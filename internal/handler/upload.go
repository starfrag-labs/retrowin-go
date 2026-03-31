package handler

import (
	"context"
	"net/url"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/application/storage"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
)

// InitiateUpload implements POST /fs/{systemId}/upload/initiate.
func (h *Handler) InitiateUpload(ctx context.Context, req *apiv1.InitiateUploadRequest, params apiv1.InitiateUploadParams) (apiv1.InitiateUploadRes, error) {
	var contentType string
	if req.ContentType.Set {
		contentType = req.ContentType.Value
	}

	session, err := h.storageSvc.InitiateUpload(ctx, &storage.InitiateUploadCommand{
		SystemID:    params.SystemId,
		ContentType: contentType,
		Size:        req.Size,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "initiate_upload_failed",
				Message: err.Error(),
			},
		}, nil
	}

	uploadURL, _ := url.Parse(session.UploadURL)

	return &apiv1.UploadSessionResponse{
		UploadSession: apiv1.UploadSession{
			ObjectId:  session.ObjectID,
			UploadUrl: *uploadURL,
			ExpiresAt: toTimestamp(session.ExpiresAt),
		},
	}, nil
}

// CompleteUpload implements POST /fs/{systemId}/upload/complete.
func (h *Handler) CompleteUpload(ctx context.Context, req *apiv1.CompleteUploadRequest, params apiv1.CompleteUploadParams) (apiv1.CompleteUploadRes, error) {
	mode := inode.ModeObject | inode.PermOwnerRW | inode.PermGroupRX | inode.PermOtherR
	if req.Mode.Set {
		mode = inode.ModeObject | int(req.Mode.Value)
	}

	result, err := h.storageSvc.CompleteUpload(ctx, &storage.CompleteUploadCommand{
		ObjectID: req.ObjectId,
		SystemID: params.SystemId,
		Mode:     mode,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "complete_upload_failed",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.InodeResponse{
		Inode: *h.toInode(result.Inode),
	}, nil
}

// GetDownloadUrl implements GET /fs/{systemId}/download.
func (h *Handler) GetDownloadUrl(ctx context.Context, params apiv1.GetDownloadUrlParams) (apiv1.GetDownloadUrlRes, error) {
	// First resolve the path to get inode ID
	in, err := h.fsSvc.ResolvePath(ctx, params.SystemId, params.Path)
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "path_not_found",
				Message: err.Error(),
			},
		}, nil
	}

	downloadURL, err := h.storageSvc.GetDownloadURL(ctx, in.ID())
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "get_download_url_failed",
				Message: err.Error(),
			},
		}, nil
	}

	parsedURL, _ := url.Parse(downloadURL)

	return &apiv1.DownloadURLResponse{
		DownloadUrl: apiv1.DownloadURL{
			DownloadUrl: *parsedURL,
			ExpiresAt:   toOptTimestamp(in.Mtime()), // TODO: Use actual expiry
		},
	}, nil
}
