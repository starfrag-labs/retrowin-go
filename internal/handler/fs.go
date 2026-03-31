package handler

import (
	"context"
	"errors"
	"net/http"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/core/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	domainerrors "github.com/starfrag-lab/retrowin-go/internal/errors"
)

// GetRootDirectory implements GET /fs/{systemId}/root.
func (h *Handler) GetRootDirectory(ctx context.Context, params apiv1.GetRootDirectoryParams) (apiv1.GetRootDirectoryRes, error) {
	rootInode, err := h.fsSvc.GetRootDirectory(ctx, params.SystemId)
	if err != nil {
		return nil, h.fsError(err)
	}

	return &apiv1.InodeResponse{
		Inode: *h.toInode(rootInode),
	}, nil
}

// StatPath implements GET /fs/{systemId}/stat.
func (h *Handler) StatPath(ctx context.Context, params apiv1.StatPathParams) (apiv1.StatPathRes, error) {
	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return nil, h.fsError(err)
	}

	return &apiv1.InodeResponse{
		Inode: *h.toInode(in),
	}, nil
}

// ReadDir implements GET /fs/{systemId}/readdir.
func (h *Handler) ReadDir(ctx context.Context, params apiv1.ReadDirParams) (apiv1.ReadDirRes, error) {
	// First resolve the directory path
	dirInode, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return nil, h.fsError(err)
	}

	// Read directory entries
	entries, err := h.fsSvc.ReadDir(ctx, dirInode.ID())
	if err != nil {
		return nil, h.fsError(err)
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
	mode := inode.ModeDirectory | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherRX
	if req.Mode.Set {
		mode = inode.ModeDirectory | int(req.Mode.Value)
	}

	dirInode, err := h.fsSvc.CreateDirectory(ctx, &fs.CreateDirectoryCommand{
		SystemID: params.SystemId,
		Mode:     mode,
	})
	if err != nil {
		return nil, h.fsError(err)
	}

	// TODO: Link the directory to its parent based on req.Path

	return &apiv1.InodeResponse{
		Inode: *h.toInode(dirInode),
	}, nil
}

// CreateSymlink implements POST /fs/{systemId}/symlink.
func (h *Handler) CreateSymlink(ctx context.Context, req *apiv1.SymlinkRequest, params apiv1.CreateSymlinkParams) (apiv1.CreateSymlinkRes, error) {
	symlinkInode, err := h.fsSvc.CreateSymlink(ctx, &fs.CreateSymlinkCommand{
		SystemID: params.SystemId,
		Target:   req.Target,
	})
	if err != nil {
		return nil, h.fsError(err)
	}

	return &apiv1.InodeResponse{
		Inode: *h.toInode(symlinkInode),
	}, nil
}

// Chmod implements PATCH /fs/{systemId}/chmod.
func (h *Handler) Chmod(ctx context.Context, req *apiv1.ChmodRequest, params apiv1.ChmodParams) (apiv1.ChmodRes, error) {
	// Resolve path to get inode
	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, req.Path)
	if err != nil {
		return nil, h.fsError(err)
	}

	// Update mode (preserve file type bits, update permission bits)
	newMode := (in.Mode() & inode.ModeTypeMask) | int(req.Mode)
	if err := h.fsSvc.UpdateMode(ctx, &fs.UpdateModeCommand{
		ID:   in.ID(),
		Mode: newMode,
	}); err != nil {
		return nil, h.fsError(err)
	}

	// Get updated inode
	updatedInode, err := h.fsSvc.Get(ctx, in.ID())
	if err != nil {
		return nil, h.fsError(err)
	}

	return &apiv1.InodeResponse{
		Inode: *h.toInode(updatedInode),
	}, nil
}

// Unlink implements DELETE /fs/{systemId}/unlink.
func (h *Handler) Unlink(ctx context.Context, params apiv1.UnlinkParams) (apiv1.UnlinkRes, error) {
	// Resolve path to get inode
	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return nil, h.fsError(err)
	}

	if err := h.fsSvc.Delete(ctx, in.ID()); err != nil {
		return nil, h.fsError(err)
	}

	return &apiv1.UnlinkNoContent{}, nil
}

// fsError converts domain errors to HTTP errors.
func (h *Handler) fsError(err error) error {
	var domainErr *domainerrors.Error
	if errors.As(err, &domainErr) {
		return &apiv1.ErrorStatusCode{
			StatusCode: domainErr.StatusCode,
			Response: apiv1.Error{
				Error: apiv1.ErrorError{
					Type:    domainErr.Code,
					Message: domainErr.Message,
				},
			},
		}
	}

	return &apiv1.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "internal_error",
				Message: err.Error(),
			},
		},
	}
}

func (h *Handler) toInode(in *inode.Inode) *apiv1.Inode {
	return &apiv1.Inode{
		ID:         in.ID(),
		SystemId:   in.SystemID(),
		Mode:       int32(in.Mode()),
		UID:        int32(in.UID()),
		Gid:        int32(in.GID()),
		Size:       in.Size(),
		LinkCount:  int32(in.LinkCount()),
		Flags:      int32(in.Flags()),
		Atime:      toOptTimestamp(in.Atime()),
		Mtime:      toOptTimestamp(in.Mtime()),
		Ctime:      toOptTimestamp(in.Ctime()),
		CreatedAt:  toOptTimestamp(in.CreatedAt()),
	}
}
