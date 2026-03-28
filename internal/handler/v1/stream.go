package v1

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

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
