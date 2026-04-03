package handler

import (
	"context"

	api "github.com/starfrag-lab/retrowin-go/pkg/api"
)

// GetHealth implements GET /health.
func (h *Handler) GetHealth(ctx context.Context) (api.GetHealthRes, error) {
	return &api.HealthStatus{
		Status: api.HealthStatusStatusHealthy,
	}, nil
}
