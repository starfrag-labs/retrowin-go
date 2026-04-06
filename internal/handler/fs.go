package handler

import (
	"context"
	"path"
	"strings"

	api "github.com/starfrag-lab/retrowin-go/pkg/api"

	"github.com/starfrag-lab/retrowin-go/internal/core/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// GetRootDirectory implements GET /fs/{systemId}/root.
func (h *Handler) GetRootDirectory(ctx context.Context, params api.GetRootDirectoryParams) (api.GetRootDirectoryRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	rootInode, err := h.fsSvc.GetRootDirectory(ctx, params.SystemId)
	if err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(rootInode),
	}, nil
}

// StatPath implements GET /fs/{systemId}/stat.
func (h *Handler) StatPath(ctx context.Context, params api.StatPathParams) (api.StatPathRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(in),
	}, nil
}

// Ls implements GET /fs/{systemId}/ls.
func (h *Handler) Ls(ctx context.Context, params api.LsParams) (api.LsRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	// First resolve the directory path
	dirInode, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Read directory entries
	entries, err := h.fsSvc.ReadDir(ctx, dirInode.ID())
	if err != nil {
		return nil, h.domainError(err)
	}

	resp := &api.DirContentResponse{
		Entries: make([]api.DirEntry, len(entries)),
	}
	for i, e := range entries {
		resp.Entries[i] = api.DirEntry{
			Name:     e.Name,
			InodeId:  e.InodeID,
			FileType: int32(e.FileType),
		}
	}

	return resp, nil
}

// Mkdir implements POST /fs/{systemId}/mkdir.
func (h *Handler) Mkdir(ctx context.Context, req *api.MkdirRequest, params api.MkdirParams) (api.MkdirRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	mode := inode.ModeDirectory | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherRX
	if req.Mode.Set {
		mode = inode.ModeDirectory | int(req.Mode.Value)
	}

	// Parse path to get parent directory and new directory name
	dirPath := path.Dir(req.Path)
	dirName := path.Base(req.Path)

	// Handle root path case
	if dirPath == "/" && dirName == "/" {
		return nil, h.domainError(errors.BadRequest("cannot create root directory"))
	}

	// Resolve parent directory
	parentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, dirPath)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Create the directory inode
	dirInode, err := h.fsSvc.CreateDirectory(ctx, &fs.CreateDirectoryCommand{
		SystemID: params.SystemId,
		Mode:     mode,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	// Link the directory to its parent
	entry := content.DirEntry{
		Name:     dirName,
		InodeID:  dirInode.ID(),
		FileType: uint8(inode.ModeDirectory >> 12), // File type for directory entry
	}
	if err := h.fsSvc.Link(ctx, parentDir.ID(), entry); err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(dirInode),
	}, nil
}

// Ln implements POST /fs/{systemId}/ln.
func (h *Handler) Ln(ctx context.Context, req *api.SymlinkRequest, params api.LnParams) (api.LnRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	if req.Target == "" {
		return nil, h.domainError(errors.BadRequest("target path is required"))
	}
	if len(req.Target) > 4096 {
		return nil, h.domainError(errors.BadRequest("target path too long"))
	}

	// Parse link path to get parent directory and link name
	linkDir := path.Dir(req.LinkPath)
	linkName := path.Base(req.LinkPath)

	// Resolve parent directory
	parentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, linkDir)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Create the symlink inode
	symlinkInode, err := h.fsSvc.CreateSymlink(ctx, &fs.CreateSymlinkCommand{
		SystemID: params.SystemId,
		Target:   req.Target,
		Mode:     inode.ModeSymlink | 0777, // Symlinks typically have 0777 permissions
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	// Link the symlink to its parent directory
	entry := content.DirEntry{
		Name:     linkName,
		InodeID:  symlinkInode.ID(),
		FileType: uint8(inode.ModeSymlink >> 12), // File type for symlink entry
	}
	if err := h.fsSvc.Link(ctx, parentDir.ID(), entry); err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(symlinkInode),
	}, nil
}

// Chmod implements PATCH /fs/{systemId}/chmod.
func (h *Handler) Chmod(ctx context.Context, req *api.ChmodRequest, params api.ChmodParams) (api.ChmodRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	if req.Mode < 0 || req.Mode > 0o777 {
		return nil, h.domainError(errors.BadRequest("mode must be between 0 and 0o777"))
	}

	// Resolve path to get inode
	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, req.Path)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Update mode (preserve file type bits, update permission bits)
	newMode := (in.Mode() & inode.ModeTypeMask) | int(req.Mode)
	if err := h.fsSvc.UpdateMode(ctx, &fs.UpdateModeCommand{
		ID:   in.ID(),
		Mode: newMode,
	}); err != nil {
		return nil, h.domainError(err)
	}

	// Get updated inode
	updatedInode, err := h.fsSvc.Get(ctx, in.ID())
	if err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(updatedInode),
	}, nil
}

// Unlink implements DELETE /fs/{systemId}/unlink.
func (h *Handler) Unlink(ctx context.Context, params api.UnlinkParams) (api.UnlinkRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	// Parse path to get parent directory and entry name
	dirPath := path.Dir(params.Path)
	entryName := path.Base(params.Path)

	// Resolve parent directory
	parentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, dirPath)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Read dir entries (triggers lazy cleanup of dangling entries)
	entries, err := h.fsSvc.ReadDir(ctx, parentDir.ID())
	if err != nil {
		return nil, h.domainError(err)
	}

	// Find target entry by name
	var targetEntry *content.DirEntry
	for i := range entries {
		if entries[i].Name == entryName {
			targetEntry = &entries[i]
			break
		}
	}
	if targetEntry == nil {
		return nil, h.domainError(errors.NotFound("path not found: " + params.Path))
	}

	// Unlink from parent directory
	if err := h.fsSvc.Unlink(ctx, parentDir.ID(), entryName); err != nil {
		return nil, h.domainError(err)
	}

	// Clean up S3 object if this is an object inode (best-effort)
	_ = h.storageSvc.DeleteObjectByInode(ctx, targetEntry.InodeID)

	// Delete the inode (may already be gone for dangling entries)
	if err := h.fsSvc.Delete(ctx, targetEntry.InodeID); err != nil && !errors.IsNotFound(err) {
		return nil, h.domainError(err)
	}

	return &api.UnlinkNoContent{}, nil
}

// Rename implements POST /fs/{systemId}/rename.
func (h *Handler) Rename(ctx context.Context, req *api.RenameReq, params api.RenameParams) (api.RenameRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	// Validate new name
	if req.NewName == "" {
		return nil, h.domainError(errors.BadRequest("new name is required"))
	}
	if path.Base(req.NewName) != req.NewName {
		return nil, h.domainError(errors.BadRequest("new name must be a simple name, not a path"))
	}

	// Resolve source path to get inode
	sourceInode, err := h.fsSvc.ResolvePath(ctx, params.SystemId, req.Path)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Get parent directory path and entry name
	sourceDirPath := path.Dir(req.Path)
	sourceEntryName := path.Base(req.Path)

	// Resolve parent directory
	sourceParentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, sourceDirPath)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Check if new name already exists in parent directory
	entries, err := h.fsSvc.ReadDir(ctx, sourceParentDir.ID())
	if err != nil {
		return nil, h.domainError(err)
	}
	for _, e := range entries {
		if e.Name == req.NewName {
			return nil, h.domainError(errors.Conflict("target already exists"))
		}
	}

	// Create new entry with same inode but new name
	newEntry := content.DirEntry{
		Name:     req.NewName,
		InodeID:  sourceInode.ID(),
		FileType: uint8(sourceInode.Mode() >> 12),
	}

	// Add new entry
	if err := h.fsSvc.Link(ctx, sourceParentDir.ID(), newEntry); err != nil {
		return nil, h.domainError(err)
	}

	// Remove old entry
	if err := h.fsSvc.Unlink(ctx, sourceParentDir.ID(), sourceEntryName); err != nil {
		return nil, h.domainError(err)
	}

	// Get updated inode
	updatedInode, err := h.fsSvc.Get(ctx, sourceInode.ID())
	if err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(updatedInode),
	}, nil
}

// Mv implements POST /fs/{systemId}/mv.
func (h *Handler) Mv(ctx context.Context, req *api.MvReq, params api.MvParams) (api.MvRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	// Cannot move to same path
	if req.Path == req.Destination {
		return nil, h.domainError(errors.BadRequest("source and destination are the same"))
	}

	// Resolve source path to get inode
	sourceInode, err := h.fsSvc.ResolvePath(ctx, params.SystemId, req.Path)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Get source parent directory and entry name
	sourceDirPath := path.Dir(req.Path)
	sourceEntryName := path.Base(req.Path)

	sourceParentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, sourceDirPath)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Determine destination directory and new entry name
	var destDirPath, newEntryName string

	// Check if destination exists
	destInode, err := h.fsSvc.ResolvePath(ctx, params.SystemId, req.Destination)
	if err == nil {
		if destInode.IsDir() {
			// Destination is an existing directory - move INTO it
			destDirPath = req.Destination
			newEntryName = sourceEntryName
		} else {
			// Destination exists and is not a directory - conflict
			return nil, h.domainError(errors.Conflict("target already exists"))
		}
	} else {
		// Destination doesn't exist, treat as full path
		destDirPath = path.Dir(req.Destination)
		newEntryName = path.Base(req.Destination)

		// Handle trailing slash case (e.g., "/home/destdir/")
		if destDirPath == "." {
			destDirPath = "/"
		}
	}

	// Normalize paths for comparison
	normalizedSource := path.Clean(req.Path)
	normalizedDest := path.Clean(destDirPath + "/" + newEntryName)
	if normalizedSource == normalizedDest {
		return nil, h.domainError(errors.BadRequest("source and destination are the same"))
	}

	// Check if moving directory into itself
	if sourceInode.IsDir() {
		if strings.HasPrefix(normalizedDest, normalizedSource+"/") {
			return nil, h.domainError(errors.BadRequest("cannot move directory into itself"))
		}
	}

	// Resolve destination parent directory
	destParentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, destDirPath)
	if err != nil {
		return nil, h.domainError(err)
	}

	// Check if entry already exists in destination parent
	destEntries, err := h.fsSvc.ReadDir(ctx, destParentDir.ID())
	if err != nil {
		return nil, h.domainError(err)
	}
	for _, e := range destEntries {
		if e.Name == newEntryName {
			return nil, h.domainError(errors.Conflict("target already exists"))
		}
	}

	// Create new entry in destination
	newEntry := content.DirEntry{
		Name:     newEntryName,
		InodeID:  sourceInode.ID(),
		FileType: uint8(sourceInode.Mode() >> 12),
	}

	if err := h.fsSvc.Link(ctx, destParentDir.ID(), newEntry); err != nil {
		return nil, h.domainError(err)
	}

	// Remove old entry from source
	if err := h.fsSvc.Unlink(ctx, sourceParentDir.ID(), sourceEntryName); err != nil {
		return nil, h.domainError(err)
	}

	// Get updated inode
	updatedInode, err := h.fsSvc.Get(ctx, sourceInode.ID())
	if err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(updatedInode),
	}, nil
}

func (h *Handler) toInode(in *inode.Inode) *api.Inode {
	return &api.Inode{
		ID:        in.ID(),
		SystemId:  in.SystemID(),
		Mode:      int32(in.Mode()),
		UID:       int32(in.UID()),
		Gid:       int32(in.GID()),
		Size:      in.Size(),
		LinkCount: int32(in.LinkCount()),
		Flags:     int32(in.Flags()),
		Atime:     toOptTimestamp(in.Atime()),
		Mtime:     toOptTimestamp(in.Mtime()),
		Ctime:     toOptTimestamp(in.Ctime()),
		CreatedAt: toOptTimestamp(in.CreatedAt()),
	}
}
