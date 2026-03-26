package v1

import (
	"context"
	"net/url"
	"time"

	"github.com/google/uuid"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/config"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/file"
	"github.com/starfrag-lab/retrowin-go/internal/middleware"
	"github.com/starfrag-lab/retrowin-go/internal/upload"
	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// Handler implements the ogen API handler interface.
type Handler struct {
	userSvc   user.Service
	fileSvc   file.Service
	uploadSvc upload.Service
	config    *config.Config
}

// NewHandler creates a new Handler.
func NewHandler(
	userSvc user.Service,
	fileSvc file.Service,
	uploadSvc upload.Service,
	cfg *config.Config,
) *Handler {
	return &Handler{
		userSvc:   userSvc,
		fileSvc:   fileSvc,
		uploadSvc: uploadSvc,
		config:    cfg,
	}
}

// Health check endpoints

// GetHealth implements GET /health.
func (h *Handler) GetHealth(ctx context.Context) (*apiv1.HealthStatus, error) {
	return &apiv1.HealthStatus{
		Status:  apiv1.HealthStatusStatusHealthy,
		Version: apiv1.NewOptString(h.config.App.Version),
	}, nil
}

// User endpoints

// GetUser implements GET /user.
func (h *Handler) GetUser(ctx context.Context) (apiv1.GetUserRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return &apiv1.GetUserUnauthorized{}, nil
	}

	u, err := h.userSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetUserNotFound{}, nil
		}
		return &apiv1.GetUserUnauthorized{}, nil
	}

	return &apiv1.UserResponse{
		User: *h.toUser(u),
	}, nil
}

// CreateUser implements POST /user.
func (h *Handler) CreateUser(ctx context.Context, req *apiv1.CreateUserRequest) (apiv1.CreateUserRes, error) {
	cmd := &user.CreateCommand{
		Provider:   string(req.Provider),
		ProviderID: req.ProviderId,
	}

	u, err := h.userSvc.Create(ctx, cmd)
	if err != nil {
		if errors.IsConflict(err) {
			return &apiv1.CreateUserConflict{}, nil
		}
		return &apiv1.CreateUserBadRequest{}, nil
	}

	return &apiv1.UserResponse{
		User: *h.toUser(u),
	}, nil
}

// DeleteUser implements DELETE /user.
func (h *Handler) DeleteUser(ctx context.Context) (apiv1.DeleteUserRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	// Get user first to obtain provider info for deletion
	u, err := h.userSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.DeleteUserNotFound{}, nil
		}
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	err = h.userSvc.Delete(ctx, u.Provider, u.ProviderID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.DeleteUserNotFound{}, nil
		}
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	return &apiv1.DeleteUserNoContent{}, nil
}

// GetServiceStatus implements GET /user/status.
func (h *Handler) GetServiceStatus(ctx context.Context) (apiv1.GetServiceStatusRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return &apiv1.GetServiceStatusUnauthorized{}, nil
	}

	// Get user first to obtain provider info
	u, err := h.userSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetServiceStatusNotFound{}, nil
		}
		return &apiv1.GetServiceStatusUnauthorized{}, nil
	}

	status, err := h.userSvc.GetServiceStatus(ctx, u.Provider, u.ProviderID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetServiceStatusNotFound{}, nil
		}
		return &apiv1.GetServiceStatusUnauthorized{}, nil
	}

	return &apiv1.ServiceStatusResponse{
		Status: *h.toServiceStatus(status),
	}, nil
}

// File endpoints

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

// Upload endpoints

// GetUploadToken implements GET /file/upload/write-token/{fileKey}.
func (h *Handler) GetUploadToken(ctx context.Context, params apiv1.GetUploadTokenParams) (apiv1.GetUploadTokenRes, error) {
	uploadURL, err := h.uploadSvc.GetUploadURL(ctx, params.FileKey.String())
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetUploadTokenNotFound{}, nil
		}
		return &apiv1.GetUploadTokenForbidden{}, nil
	}

	return &apiv1.UploadTokenResponse{
		UploadToken: *h.toUploadURL(uploadURL),
	}, nil
}

// CompleteUpload implements PATCH /file/upload/complete/{fileKey}.
func (h *Handler) CompleteUpload(ctx context.Context, req apiv1.OptCompleteUploadRequest, params apiv1.CompleteUploadParams) (apiv1.CompleteUploadRes, error) {
	var byteSize int64
	if r, ok := req.Get(); ok {
		if bs, ok := r.ByteSize.Get(); ok {
			byteSize = bs
		}
	}

	f, err := h.uploadSvc.CompleteUpload(ctx, params.FileKey.String(), byteSize)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.CompleteUploadNotFound{}, nil
		}
		return &apiv1.CompleteUploadForbidden{}, nil
	}

	return h.toFileResponse(f), nil
}

// Stream endpoints

// GetStreamToken implements GET /file/stream/read-token/{fileKey}.
func (h *Handler) GetStreamToken(ctx context.Context, params apiv1.GetStreamTokenParams) (apiv1.GetStreamTokenRes, error) {
	streamURL, err := h.uploadSvc.GetStreamURL(ctx, params.FileKey.String())
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetStreamTokenNotFound{}, nil
		}
		return &apiv1.GetStreamTokenForbidden{}, nil
	}

	return &apiv1.StreamTokenResponse{
		StreamToken: *h.toStreamURL(streamURL),
	}, nil
}

// Helper functions

func toTimestamp(t time.Time) apiv1.Timestamp {
	return apiv1.Timestamp(t)
}

func (h *Handler) toUser(u *user.User) *apiv1.User {
	return &apiv1.User{
		ID:         u.ID,
		Provider:   apiv1.Provider(u.Provider),
		ProviderId: u.ProviderID,
		CreatedAt:  apiv1.NewOptTimestamp(toTimestamp(u.CreatedAt)),
		UpdatedAt:  apiv1.NewOptTimestamp(toTimestamp(u.UpdatedAt)),
	}
}

func (h *Handler) toServiceStatus(s *user.ServiceStatus) *apiv1.ServiceStatus {
	return &apiv1.ServiceStatus{
		UserId:     s.UserID,
		Available:  s.Available,
		JoinDate:   apiv1.NewOptTimestamp(toTimestamp(s.JoinDate)),
		UpdateDate: apiv1.NewOptTimestamp(toTimestamp(s.UpdateDate)),
	}
}

func (h *Handler) toFile(f *file.File) *apiv1.File {
	fileKey, _ := uuid.Parse(f.FileKey)
	return &apiv1.File{
		ID:        f.ID,
		FileKey:   fileKey,
		Type:      apiv1.FileType(f.Type),
		FileName:  f.FileName,
		OwnerId:   apiv1.NewOptInt64(f.OwnerID),
		ParentId:  toOptNilInt64(f.ParentID),
		ByteSize:  apiv1.NewOptInt64(f.ByteSize),
		CreatedAt: apiv1.NewOptTimestamp(toTimestamp(f.CreatedAt)),
		UpdatedAt: apiv1.NewOptTimestamp(toTimestamp(f.UpdatedAt)),
		Path:      f.Path,
		Roles:     f.Roles,
	}
}

func (h *Handler) toFileResponse(f *file.File) *apiv1.FileResponse {
	return &apiv1.FileResponse{
		File: *h.toFile(f),
	}
}

func (h *Handler) toUploadURL(u *upload.UploadURL) *apiv1.UploadToken {
	uploadURL, _ := url.Parse(u.UploadURL)
	return &apiv1.UploadToken{
		UploadUrl: *uploadURL,
		ExpiresAt: toTimestamp(u.ExpiresAt),
	}
}

func (h *Handler) toStreamURL(s *upload.StreamURL) *apiv1.StreamToken {
	downloadURL, _ := url.Parse(s.DownloadURL)
	return &apiv1.StreamToken{
		DownloadUrl: *downloadURL,
		ExpiresAt:   toTimestamp(s.ExpiresAt),
	}
}

// Utility functions

func toOptNilInt64(v *int64) apiv1.OptNilInt64 {
	if v == nil {
		return apiv1.OptNilInt64{}
	}
	return apiv1.NewOptNilInt64(*v)
}
