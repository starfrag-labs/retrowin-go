package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	coreuser "github.com/starfrag-lab/retrowin-go/internal/core/user"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/service/sysinit"
	"github.com/starfrag-lab/retrowin-go/internal/system"
	"github.com/starfrag-lab/retrowin-go/internal/utils"
)

// CreateSystem implements POST /systems.
func (h *Handler) CreateSystem(ctx context.Context, req *apiv1.CreateSystemRequest) (apiv1.CreateSystemRes, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		return nil, h.domainError(errors.Unauthorized("user not authenticated"))
	}

	var description *string
	if req.Description.Set {
		description = &req.Description.Value
	}

	result, err := h.initSvc.InitSystem(ctx, &sysinit.InitSystemCommand{
		Name:        req.Name,
		Description: description,
		RootUserID:  userID,
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
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		return nil, h.domainError(errors.Unauthorized("user not authenticated"))
	}

	// Find all system memberships for this user
	memberships, err := h.sysUserSvc.Find(ctx, coreuser.ByUserID(userID))
	if err != nil {
		return nil, h.domainError(err)
	}

	// Collect system IDs
	systemIDs := make([]string, len(memberships))
	for i, m := range memberships {
		systemIDs[i] = m.SystemID()
	}

	// Load each system
	resp := &apiv1.SystemListResponse{
		Systems: make([]apiv1.System, 0, len(systemIDs)),
	}
	for _, sysID := range systemIDs {
		sys, err := h.systemSvc.GetByID(ctx, sysID)
		if err != nil {
			continue // Skip systems that may have been deleted
		}
		resp.Systems = append(resp.Systems, *h.toSystem(sys))
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
