package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"
)

// GetHealth implements GET /health.
func (h *Handler) GetHealth(ctx context.Context) (apiv1.GetHealthRes, error) {
	return &apiv1.HealthStatus{
		Status: apiv1.HealthStatusStatusHealthy,
	}, nil
}
