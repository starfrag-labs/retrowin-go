package handler

import (
	"context"
	"path"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/core/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// GetRootDirectory implements GET /fs/{systemId}/root.
func (h *Handler) GetRootDirectory(ctx context.Context, params apiv1.GetRootDirectoryParams) (apiv1.GetRootDirectoryRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	rootInode, err := h.fsSvc.GetRootDirectory(ctx, params.SystemId)
	if err != nil {
		return nil, h.domainError(err)
	}

	return &apiv1.InodeResponse{
		Inode: *h.toInode(rootInode),
	}, nil
}

// StatPath implements GET /fs/{systemId}/stat.
func (h *Handler) StatPath(ctx context.Context, params apiv1.StatPathParams) (apiv1.StatPathRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return nil, h.domainError(err)
	}

	return &apiv1.InodeResponse{
		Inode: *h.toInode(in),
	}, nil
}

// ReadDir implements GET /fs/{systemId}/readdir.
func (h *Handler) ReadDir(ctx context.Context, params apiv1.ReadDirParams) (apiv1.ReadDirRes, error) {
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

	resp := &apiv1.DirContentResponse{
		Entries: make([]apiv1.DirEntry, len(entries)),
	}
	for i, e := range entries {
		resp.Entries[i] = apiv1.DirEntry{
			Name:     e.Name,
			InodeId:  e.InodeID,
			FileType: int32(e.FileType),
		}
	}

	return resp, nil
}

// Mkdir implements POST /fs/{systemId}/mkdir.
func (h *Handler) Mkdir(ctx context.Context, req *apiv1.MkdirRequest, params apiv1.MkdirParams) (apiv1.MkdirRes, error) {
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

	return &apiv1.InodeResponse{
		Inode: *h.toInode(dirInode),
	}, nil
}

// CreateSymlink implements POST /fs/{systemId}/symlink.
func (h *Handler) CreateSymlink(ctx context.Context, req *apiv1.SymlinkRequest, params apiv1.CreateSymlinkParams) (apiv1.CreateSymlinkRes, error) {
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

	return &apiv1.InodeResponse{
		Inode: *h.toInode(symlinkInode),
	}, nil
}

// Chmod implements PATCH /fs/{systemId}/chmod.
func (h *Handler) Chmod(ctx context.Context, req *apiv1.ChmodRequest, params apiv1.ChmodParams) (apiv1.ChmodRes, error) {
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

	return &apiv1.InodeResponse{
		Inode: *h.toInode(updatedInode),
	}, nil
}

// Unlink implements DELETE /fs/{systemId}/unlink.
func (h *Handler) Unlink(ctx context.Context, params apiv1.UnlinkParams) (apiv1.UnlinkRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	// Resolve path to get inode and parent
	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
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

	// Unlink from parent directory first
	if err := h.fsSvc.Unlink(ctx, parentDir.ID(), entryName); err != nil {
		return nil, h.domainError(err)
	}

	// Clean up S3 object if this is an object inode (best-effort)
	_ = h.storageSvc.DeleteObjectByInode(ctx, in.ID())

	// Delete the inode
	if err := h.fsSvc.Delete(ctx, in.ID()); err != nil {
		return nil, h.domainError(err)
	}

	return &apiv1.UnlinkNoContent{}, nil
}

func (h *Handler) toInode(in *inode.Inode) *apiv1.Inode {
	return &apiv1.Inode{
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
