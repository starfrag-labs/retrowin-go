package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	coreuser "github.com/starfrag-lab/retrowin-go/internal/core/user"
)

// CreateSystemGroup implements POST /systems/{systemId}/groups.
func (h *Handler) CreateSystemGroup(ctx context.Context, req *apiv1.CreateSystemGroupRequest, params apiv1.CreateSystemGroupParams) (apiv1.CreateSystemGroupRes, error) {
	cmd := &coreuser.GroupCreateCommand{
		SystemID: params.SystemId,
		Name:     req.Name,
	}
	if req.Gid.Set {
		cmd.GID = int(req.Gid.Value)
	}

	group, err := h.sysGroupSvc.Create(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &apiv1.SystemGroupResponse{
		Group: *h.toSystemGroup(group),
	}, nil
}

// ListSystemGroups implements GET /systems/{systemId}/groups.
func (h *Handler) ListSystemGroups(ctx context.Context, params apiv1.ListSystemGroupsParams) (apiv1.ListSystemGroupsRes, error) {
	groups, err := h.sysGroupSvc.Find(ctx, coreuser.GroupBySystemID(params.SystemId))
	if err != nil {
		return nil, err
	}

	resp := &apiv1.SystemGroupListResponse{
		Groups: make([]apiv1.SystemGroup, len(groups)),
	}
	for i, g := range groups {
		resp.Groups[i] = *h.toSystemGroup(g)
	}

	return resp, nil
}

// GetSystemGroup implements GET /systems/{systemId}/groups/{gid}.
func (h *Handler) GetSystemGroup(ctx context.Context, params apiv1.GetSystemGroupParams) (apiv1.GetSystemGroupRes, error) {
	group, err := h.sysGroupSvc.FindOne(ctx, coreuser.GroupBySystemAndGID(params.SystemId, int(params.Gid)))
	if err != nil {
		return nil, err
	}

	return &apiv1.SystemGroupResponse{
		Group: *h.toSystemGroup(group),
	}, nil
}

// DeleteSystemGroup implements DELETE /systems/{systemId}/groups/{gid}.
func (h *Handler) DeleteSystemGroup(ctx context.Context, params apiv1.DeleteSystemGroupParams) (apiv1.DeleteSystemGroupRes, error) {
	// Find group first
	group, err := h.sysGroupSvc.FindOne(ctx, coreuser.GroupBySystemAndGID(params.SystemId, int(params.Gid)))
	if err != nil {
		return nil, err
	}

	if err := h.sysGroupSvc.Delete(ctx, group.ID()); err != nil {
		return nil, err
	}

	return &apiv1.DeleteSystemGroupNoContent{}, nil
}

// AddGroupMember implements POST /systems/{systemId}/groups/{gid}/members/{uid}.
func (h *Handler) AddGroupMember(ctx context.Context, params apiv1.AddGroupMemberParams) (apiv1.AddGroupMemberRes, error) {
	// Find group first
	group, err := h.sysGroupSvc.FindOne(ctx, coreuser.GroupBySystemAndGID(params.SystemId, int(params.Gid)))
	if err != nil {
		return nil, err
	}

	// Find user by UID within system
	users, err := h.sysUserSvc.Find(ctx, coreuser.Filter{
		SystemID: &params.SystemId,
	})
	if err != nil {
		return nil, err
	}

	// Find user with matching UID
	var targetUser *coreuser.SystemUser
	for _, u := range users {
		if u.UID() == int(params.UID) {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return &apiv1.AddGroupMemberNotFound{}, nil
	}

	if err := h.sysGroupSvc.AddUserToGroup(ctx, targetUser.ID(), group.ID()); err != nil {
		return nil, err
	}

	return &apiv1.AddGroupMemberNoContent{}, nil
}

// RemoveGroupMember implements DELETE /systems/{systemId}/groups/{gid}/members/{uid}.
func (h *Handler) RemoveGroupMember(ctx context.Context, params apiv1.RemoveGroupMemberParams) (apiv1.RemoveGroupMemberRes, error) {
	// Find group first
	group, err := h.sysGroupSvc.FindOne(ctx, coreuser.GroupBySystemAndGID(params.SystemId, int(params.Gid)))
	if err != nil {
		return nil, err
	}

	// Find user by UID within system
	users, err := h.sysUserSvc.Find(ctx, coreuser.Filter{
		SystemID: &params.SystemId,
	})
	if err != nil {
		return nil, err
	}

	// Find user with matching UID
	var targetUser *coreuser.SystemUser
	for _, u := range users {
		if u.UID() == int(params.UID) {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return &apiv1.RemoveGroupMemberNotFound{}, nil
	}

	if err := h.sysGroupSvc.RemoveUserFromGroup(ctx, targetUser.ID(), group.ID()); err != nil {
		return nil, err
	}

	return &apiv1.RemoveGroupMemberNoContent{}, nil
}

func (h *Handler) toSystemGroup(g *coreuser.SystemGroup) *apiv1.SystemGroup {
	return &apiv1.SystemGroup{
		ID:       int64(g.ID()),
		SystemId: g.SystemID(),
		Name:     g.Name(),
		Gid:      int32(g.GID()),
	}
}
