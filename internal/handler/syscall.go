package handler

import (
	"context"
	"path"

	api "github.com/starfrag-lab/retrowin-go/pkg/api"

	"github.com/starfrag-lab/retrowin-go/internal/application/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/dentry"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Ls implements GET /syscall/{systemId}/ls — getdents64
func (h *Handler) Ls(ctx context.Context, params api.LsParams) (api.LsRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	dirInode, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return nil, h.domainError(err)
	}

	entries, err := h.dentrySvc.ReadDir(ctx, dirInode.ID())
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

// Mkdir implements POST /syscall/{systemId}/mkdir — mkdirat
func (h *Handler) Mkdir(ctx context.Context, req *api.MkdirRequest, params api.MkdirParams) (api.MkdirRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	mode := inode.ModeDirectory | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherRX
	if req.Mode.Set {
		mode = inode.ModeDirectory | int(req.Mode.Value)
	}

	dirPath := path.Dir(req.Path)
	dirName := path.Base(req.Path)

	if dirPath == "/" && dirName == "/" {
		return nil, h.domainError(errors.BadRequest("cannot create root directory"))
	}

	parentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, dirPath)
	if err != nil {
		return nil, h.domainError(err)
	}

	dirInode, err := h.fsSvc.CreateDirectory(ctx, &fs.CreateDirectoryCommand{
		SystemID: params.SystemId,
		Mode:     mode,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	entry := dentry.DirEntry{
		Name:     dirName,
		InodeID:  dirInode.ID(),
		FileType: uint8(inode.ModeDirectory >> 12),
	}
	if err := h.dentrySvc.Link(ctx, parentDir.ID(), entry); err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(dirInode),
	}, nil
}

// Ln implements POST /syscall/{systemId}/ln — symlinkat
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

	linkDir := path.Dir(req.LinkPath)
	linkName := path.Base(req.LinkPath)

	parentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, linkDir)
	if err != nil {
		return nil, h.domainError(err)
	}

	symlinkInode, err := h.fsSvc.CreateSymlink(ctx, &fs.CreateSymlinkCommand{
		SystemID: params.SystemId,
		Target:   req.Target,
		Mode:     inode.ModeSymlink | 0777,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	entry := dentry.DirEntry{
		Name:     linkName,
		InodeID:  symlinkInode.ID(),
		FileType: uint8(inode.ModeSymlink >> 12),
	}
	if err := h.dentrySvc.Link(ctx, parentDir.ID(), entry); err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(symlinkInode),
	}, nil
}

// Unlink implements DELETE /syscall/{systemId}/unlink — unlinkat
func (h *Handler) Unlink(ctx context.Context, params api.UnlinkParams) (api.UnlinkRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	dirPath := path.Dir(params.Path)
	entryName := path.Base(params.Path)

	parentDir, err := h.fsSvc.ResolvePath(ctx, params.SystemId, dirPath)
	if err != nil {
		return nil, h.domainError(err)
	}

	entries, err := h.dentrySvc.ReadDir(ctx, parentDir.ID())
	if err != nil {
		return nil, h.domainError(err)
	}

	var targetEntry *dentry.DirEntry
	for i := range entries {
		if entries[i].Name == entryName {
			targetEntry = &entries[i]
			break
		}
	}
	if targetEntry == nil {
		return nil, h.domainError(errors.NotFound("path not found: " + params.Path))
	}

	if err := h.dentrySvc.Unlink(ctx, parentDir.ID(), entryName); err != nil {
		return nil, h.domainError(err)
	}

	if err := h.fsSvc.Delete(ctx, targetEntry.InodeID); err != nil && !errors.IsNotFound(err) {
		return nil, h.domainError(err)
	}

	return &api.UnlinkNoContent{}, nil
}

// Rename implements POST /syscall/{systemId}/rename — renameat
func (h *Handler) Rename(ctx context.Context, req *api.RenameRequest, params api.RenameParams) (api.RenameRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	renamedInode, err := h.fsSvc.Rename(ctx, &fs.RenameCommand{
		SystemID: params.SystemId,
		Path:     req.Path,
		NewName:  req.NewName,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	return &api.InodeResponse{
		Inode: *h.toInode(renamedInode),
	}, nil
}

// Mv implements POST /syscall/{systemId}/mv — move multiple paths
func (h *Handler) Mv(ctx context.Context, req *api.MvRequest, params api.MvParams) (api.MvRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	result, err := h.fsSvc.Mv(ctx, &fs.MvCommand{
		SystemID:    params.SystemId,
		Sources:     req.Sources,
		Destination: req.Destination,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	resp := &api.MvResult{
		Moved:  result.Moved,
		Errors: make([]api.MvResultErrorsItem, len(result.Errors)),
	}
	for i, e := range result.Errors {
		resp.Errors[i] = api.MvResultErrorsItem{
			Path:  e.Path,
			Error: e.Error.Error(),
		}
	}

	return resp, nil
}

// Rm implements POST /syscall/{systemId}/rm — remove multiple paths
func (h *Handler) Rm(ctx context.Context, req *api.RmRequest, params api.RmParams) (api.RmRes, error) {
	if err := h.checkSystemAccess(ctx, params.SystemId); err != nil {
		return nil, h.domainError(err)
	}

	result, err := h.fsSvc.Rm(ctx, &fs.RmCommand{
		SystemID: params.SystemId,
		Paths:    req.Paths,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	resp := &api.RmResult{
		Deleted: result.Deleted,
		Errors:  make([]api.RmResultErrorsItem, len(result.Errors)),
	}
	for i, e := range result.Errors {
		resp.Errors[i] = api.RmResultErrorsItem{
			Path:  e.Path,
			Error: e.Error.Error(),
		}
	}

	return resp, nil
}
