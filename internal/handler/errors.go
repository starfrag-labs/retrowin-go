package handler

import (
	"context"
	"errors"
	"net/http"

	domainerrors "github.com/starfrag-lab/retrowin-go/internal/errors"
	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"
)

// NewError creates error response for ogen convenient errors.
func (h *Handler) NewError(ctx context.Context, err error) *apiv1.ErrorStatusCode {
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
