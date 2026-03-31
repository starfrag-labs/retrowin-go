package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	initsvc "github.com/starfrag-lab/retrowin-go/internal/service/init"
	"github.com/starfrag-lab/retrowin-go/internal/system"
)

// CreateSystem implements POST /systems.
func (h *Handler) CreateSystem(ctx context.Context, req *apiv1.CreateSystemRequest) (apiv1.CreateSystemRes, error) {
	var description *string
	if req.Description.Set {
		description = &req.Description.Value
	}

	result, err := h.initSvc.InitSystem(ctx, &initsvc.InitSystemCommand{
		Name:        req.Name,
		Description: description,
	})
	if err != nil {
		return nil, h.domainError(err)
	}

	return &apiv1.SystemResponse{
		System: *h.toSystem(result.System),
	}, nil
}

// ListSystems implements GET /systems.
func (h *Handler) ListSystems(ctx context.Context) (apiv1.ListSystemsRes, error) {
	systems, err := h.systemSvc.Find(ctx, system.Filter{})
	if err != nil {
		return nil, h.domainError(err)
	}

	resp := &apiv1.SystemListResponse{
		Systems: make([]apiv1.System, len(systems)),
	}
	for i, sys := range systems {
		resp.Systems[i] = *h.toSystem(sys)
	}

	return resp, nil
}

// GetSystem implements GET /systems/{systemId}.
func (h *Handler) GetSystem(ctx context.Context, params apiv1.GetSystemParams) (apiv1.GetSystemRes, error) {
	sys, err := h.systemSvc.GetByID(ctx, params.SystemId)
	if err != nil {
		return nil, h.domainError(err)
	}

	return &apiv1.SystemResponse{
		System: *h.toSystem(sys),
	}, nil
}

func (h *Handler) toSystem(sys *system.System) *apiv1.System {
	resp := &apiv1.System{
		ID:        sys.ID(),
		Name:      sys.Name(),
		Status:    apiv1.SystemStatus(sys.Status()),
		CreatedAt: toOptTimestamp(sys.CreatedAt()),
		UpdatedAt: toOptTimestamp(sys.UpdatedAt()),
	}
	if desc := sys.Description(); desc != nil {
		resp.Description.Set = true
		resp.Description.Value = *desc
	}
	return resp
}
