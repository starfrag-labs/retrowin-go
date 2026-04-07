package handler

import (
	"context"

	api "github.com/starfrag-lab/retrowin-go/pkg/api"

	"github.com/starfrag-lab/retrowin-go/internal/application/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
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
