package v1

import (
	"net/url"
	"time"

	"github.com/google/uuid"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/auth"
	"github.com/starfrag-lab/retrowin-go/internal/file"
	"github.com/starfrag-lab/retrowin-go/internal/upload"
	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// Handler implements the ogen API handler interface.
type Handler struct {
	userSvc   user.Service
	fileSvc   file.Service
	uploadSvc upload.Service
	authSvc   auth.Service
}

// NewHandler creates a new Handler.
func NewHandler(
	userSvc user.Service,
	fileSvc file.Service,
	uploadSvc upload.Service,
	authSvc auth.Service,
) *Handler {
	return &Handler{
		userSvc:   userSvc,
		fileSvc:   fileSvc,
		uploadSvc: uploadSvc,
		authSvc:   authSvc,
	}
}

// Ensure Handler implements the ogen Handler interface.
var _ apiv1.Handler = (*Handler)(nil)

// Helper functions

func toTimestamp(t time.Time) apiv1.Timestamp {
	return apiv1.Timestamp(t)
}

func (h *Handler) toUser(u *user.User) *apiv1.User {
	return &apiv1.User{
		ID:         u.ID,
		Provider:   apiv1.Provider(u.Provider),
		ProviderId: u.ProviderID,
		JoinDate:   apiv1.NewOptTimestamp(toTimestamp(u.JoinDate)),
		CreatedAt:  apiv1.NewOptTimestamp(toTimestamp(u.CreatedAt)),
		UpdatedAt:  apiv1.NewOptTimestamp(toTimestamp(u.UpdatedAt)),
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
