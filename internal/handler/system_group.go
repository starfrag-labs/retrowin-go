package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/core/user"
)

// CreateSystemGroup implements POST /systems/{systemId}/groups.
func (h *Handler) CreateSystemGroup(ctx context.Context, req *apiv1.CreateSystemGroupRequest, params apiv1.CreateSystemGroupParams) (apiv1.CreateSystemGroupRes, error) {
	createParams := &user.GroupCreateParams{
		SystemID: params.SystemId,
		Name:     req.Name,
	}
	if req.Gid.Set {
		createParams.GID = int(req.Gid.Value)
	}

	// TODO: Need group service method to create group
	// For now, we'll use the user service's group repository
	group, err := h.userSvc.CreateGroup(ctx, createParams)
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "create_system_group_failed",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.SystemGroupResponse{
		Group: *h.toSystemGroup(group),
	}, nil
}

// ListSystemGroups implements GET /systems/{systemId}/groups.
func (h *Handler) ListSystemGroups(ctx context.Context, params apiv1.ListSystemGroupsParams) (apiv1.ListSystemGroupsRes, error) {
	groups, err := h.userSvc.FindGroups(ctx, user.GroupFilter{
		SystemID: &params.SystemId,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "list_system_groups_failed",
				Message: err.Error(),
			},
		}, nil
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
	group, err := h.userSvc.FindOneGroup(ctx, user.GroupFilter{
		SystemID: &params.SystemId,
		GID:      &params.Gid,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "system_group_not_found",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.SystemGroupResponse{
		Group: *h.toSystemGroup(group),
	}, nil
}

// DeleteSystemGroup implements DELETE /systems/{systemId}/groups/{gid}.
func (h *Handler) DeleteSystemGroup(ctx context.Context, params apiv1.DeleteSystemGroupParams) (apiv1.DeleteSystemGroupRes, error) {
	// Find group first
	group, err := h.userSvc.FindOneGroup(ctx, user.GroupFilter{
		SystemID: &params.SystemId,
		GID:      &params.Gid,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "system_group_not_found",
				Message: err.Error(),
			},
		}, nil
	}

	if err := h.userSvc.DeleteGroup(ctx, group.ID()); err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "delete_system_group_failed",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.DeleteSystemGroupNoContent{}, nil
}

// AddGroupMember implements POST /systems/{systemId}/groups/{gid}/members/{uid}.
func (h *Handler) AddGroupMember(ctx context.Context, params apiv1.AddGroupMemberParams) (apiv1.AddGroupMemberRes, error) {
	// Find group and user first
	group, err := h.userSvc.FindOneGroup(ctx, user.GroupFilter{
		SystemID: &params.SystemId,
		GID:      &params.Gid,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "system_group_not_found",
				Message: err.Error(),
			},
		}, nil
	}

	sysUser, err := h.userSvc.FindOne(ctx, user.Filter{
		SystemID: &params.SystemId,
		UID:      &params.Uid,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "system_user_not_found",
				Message: err.Error(),
			},
		}, nil
	}

	if err := h.userSvc.AddUserToGroup(ctx, sysUser.ID(), group.ID()); err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "add_group_member_failed",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.AddGroupMemberNoContent{}, nil
}

// RemoveGroupMember implements DELETE /systems/{systemId}/groups/{gid}/members/{uid}.
func (h *Handler) RemoveGroupMember(ctx context.Context, params apiv1.RemoveGroupMemberParams) (apiv1.RemoveGroupMemberRes, error) {
	// Find group and user first
	group, err := h.userSvc.FindOneGroup(ctx, user.GroupFilter{
		SystemID: &params.SystemId,
		GID:      &params.Gid,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "system_group_not_found",
				Message: err.Error(),
			},
		}, nil
	}

	sysUser, err := h.userSvc.FindOne(ctx, user.Filter{
		SystemID: &params.SystemId,
		UID:      &params.Uid,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "system_user_not_found",
				Message: err.Error(),
			},
		}, nil
	}

	if err := h.userSvc.RemoveUserFromGroup(ctx, sysUser.ID(), group.ID()); err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "remove_group_member_failed",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.RemoveGroupMemberNoContent{}, nil
}

func (h *Handler) toSystemGroup(g *user.SystemGroup) *apiv1.SystemGroup {
	return &apiv1.SystemGroup{
		Id:        int64(g.ID()),
		SystemId:  g.SystemID(),
		Name:      g.Name(),
		Gid:       g.GID(),
		Members:   g.Members(),
		CreatedAt: toOptTimestamp(g.CreatedAt()),
	}
}
