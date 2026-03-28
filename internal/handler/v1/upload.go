package v1

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

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
