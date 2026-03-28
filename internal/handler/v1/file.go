package v1

import (
	"context"

	"github.com/google/uuid"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/file"
	"github.com/starfrag-lab/retrowin-go/internal/middleware"
)

// GetRootContainer implements GET /file/root.
func (h *Handler) GetRootContainer(ctx context.Context) (apiv1.GetRootContainerRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return &apiv1.GetRootContainerUnauthorized{}, nil
	}

	f, err := h.fileSvc.GetRoot(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetRootContainerNotFound{}, nil
		}
		return &apiv1.GetRootContainerUnauthorized{}, nil
	}

	return h.toFileResponse(f), nil
}

// GetHomeContainer implements GET /file/home.
func (h *Handler) GetHomeContainer(ctx context.Context) (apiv1.GetHomeContainerRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return &apiv1.GetHomeContainerUnauthorized{}, nil
	}

	f, err := h.fileSvc.GetHome(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetHomeContainerNotFound{}, nil
		}
		return &apiv1.GetHomeContainerUnauthorized{}, nil
	}

	return h.toFileResponse(f), nil
}

// GetTrashContainer implements GET /file/trash.
func (h *Handler) GetTrashContainer(ctx context.Context) (apiv1.GetTrashContainerRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return &apiv1.GetTrashContainerUnauthorized{}, nil
	}

	f, err := h.fileSvc.GetTrash(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetTrashContainerNotFound{}, nil
		}
		return &apiv1.GetTrashContainerUnauthorized{}, nil
	}

	return h.toFileResponse(f), nil
}

// GetFileInfo implements GET /file/info/{fileKey}.
func (h *Handler) GetFileInfo(ctx context.Context, params apiv1.GetFileInfoParams) (apiv1.GetFileInfoRes, error) {
	f, err := h.fileSvc.Get(ctx, params.FileKey.String())
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetFileInfoNotFound{}, nil
		}
		return &apiv1.GetFileInfoForbidden{}, nil
	}

	return h.toFileResponse(f), nil
}

// GetFileChildren implements GET /file/children/{fileKey}.
func (h *Handler) GetFileChildren(ctx context.Context, params apiv1.GetFileChildrenParams) (apiv1.GetFileChildrenRes, error) {
	children, err := h.fileSvc.GetChildren(ctx, params.FileKey.String())
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetFileChildrenNotFound{}, nil
		}
		return &apiv1.GetFileChildrenForbidden{}, nil
	}

	files := make([]apiv1.File, len(children))
	for i, f := range children {
		files[i] = *h.toFile(f)
	}

	return &apiv1.FileListResponse{Files: files}, nil
}

// CreateFile implements POST /file.
func (h *Handler) CreateFile(ctx context.Context, req *apiv1.CreateFileRequest) (apiv1.CreateFileRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return &apiv1.CreateFileUnauthorized{}, nil
	}

	cmd := &file.CreateCommand{
		Type:     file.FileType(req.Type),
		FileName: req.FileName,
		OwnerID:  userID,
	}

	if parentKey, ok := req.ParentKey.Get(); ok {
		if parentKey != uuid.Nil {
			pk := parentKey.String()
			cmd.ParentKey = &pk
		}
	}

	f, err := h.fileSvc.Create(ctx, cmd)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.CreateFileNotFound{}, nil
		}
		if errors.IsUnauthorized(err) {
			return &apiv1.CreateFileUnauthorized{}, nil
		}
		return &apiv1.CreateFileBadRequest{}, nil
	}

	return h.toFileResponse(f), nil
}

// UpdateFile implements PATCH /file/{fileKey}.
func (h *Handler) UpdateFile(ctx context.Context, req *apiv1.UpdateFileRequest, params apiv1.UpdateFileParams) (apiv1.UpdateFileRes, error) {
	cmd := &file.UpdateCommand{}
	if fileName, ok := req.FileName.Get(); ok {
		cmd.FileName = &fileName
	}

	f, err := h.fileSvc.Update(ctx, params.FileKey.String(), cmd)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.UpdateFileNotFound{}, nil
		}
		return &apiv1.UpdateFileForbidden{}, nil
	}

	return h.toFileResponse(f), nil
}

// DeleteFile implements DELETE /file/{fileKey}.
func (h *Handler) DeleteFile(ctx context.Context, params apiv1.DeleteFileParams) (apiv1.DeleteFileRes, error) {
	permanent := false
	if p, ok := params.Permanent.Get(); ok {
		permanent = p
	}

	err := h.fileSvc.Delete(ctx, params.FileKey.String(), permanent)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.DeleteFileNotFound{}, nil
		}
		return &apiv1.DeleteFileForbidden{}, nil
	}

	return &apiv1.DeleteFileNoContent{}, nil
}

// MoveFile implements POST /file/{fileKey}/move.
func (h *Handler) MoveFile(ctx context.Context, req *apiv1.MoveFileRequest, params apiv1.MoveFileParams) (apiv1.MoveFileRes, error) {
	cmd := &file.MoveCommand{
		TargetKey: req.TargetKey.String(),
	}

	f, err := h.fileSvc.Move(ctx, params.FileKey.String(), cmd)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.MoveFileNotFound{}, nil
		}
		return &apiv1.MoveFileForbidden{}, nil
	}

	return h.toFileResponse(f), nil
}

// CopyFile implements POST /file/{fileKey}/copy.
func (h *Handler) CopyFile(ctx context.Context, req *apiv1.CopyFileRequest, params apiv1.CopyFileParams) (apiv1.CopyFileRes, error) {
	cmd := &file.CopyCommand{
		TargetKey: req.TargetKey.String(),
	}

	f, err := h.fileSvc.Copy(ctx, params.FileKey.String(), cmd)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.CopyFileNotFound{}, nil
		}
		return &apiv1.CopyFileForbidden{}, nil
	}

	return h.toFileResponse(f), nil
}
