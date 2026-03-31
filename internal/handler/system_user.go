package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	coreuser "github.com/starfrag-lab/retrowin-go/internal/core/user"
)

// CreateSystemUser implements POST /systems/{systemId}/users.
func (h *Handler) CreateSystemUser(ctx context.Context, req *apiv1.CreateSystemUserRequest, params apiv1.CreateSystemUserParams) (apiv1.CreateSystemUserRes, error) {
	cmd := &coreuser.CreateCommand{
		UserID:   req.UserId,
		SystemID: params.SystemId,
		Username: req.Username,
	}
	if req.UID.Set {
		cmd.UID = int(req.UID.Value)
	}

	sysUser, err := h.sysUserSvc.Create(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &apiv1.SystemUserResponse{
		User: *h.toSystemUser(sysUser),
	}, nil
}

// ListSystemUsers implements GET /systems/{systemId}/users.
func (h *Handler) ListSystemUsers(ctx context.Context, params apiv1.ListSystemUsersParams) (apiv1.ListSystemUsersRes, error) {
	users, err := h.sysUserSvc.Find(ctx, coreuser.BySystemID(params.SystemId))
	if err != nil {
		return nil, err
	}

	resp := &apiv1.SystemUserListResponse{
		Users: make([]apiv1.SystemUser, len(users)),
	}
	for i, u := range users {
		resp.Users[i] = *h.toSystemUser(u)
	}

	return resp, nil
}

// GetSystemUser implements GET /systems/{systemId}/users/{uid}.
func (h *Handler) GetSystemUser(ctx context.Context, params apiv1.GetSystemUserParams) (apiv1.GetSystemUserRes, error) {
	uid := int(params.UID)
	users, err := h.sysUserSvc.Find(ctx, coreuser.Filter{
		SystemID: &params.SystemId,
	})
	if err != nil {
		return nil, err
	}

	// Find user with matching UID
	for _, u := range users {
		if u.UID() == uid {
			return &apiv1.SystemUserResponse{
				User: *h.toSystemUser(u),
			}, nil
		}
	}

	return &apiv1.GetSystemUserNotFound{}, nil
}

// DeleteSystemUser implements DELETE /systems/{systemId}/users/{uid}.
func (h *Handler) DeleteSystemUser(ctx context.Context, params apiv1.DeleteSystemUserParams) (apiv1.DeleteSystemUserRes, error) {
	// Find user by UID within system
	uid := int(params.UID)
	users, err := h.sysUserSvc.Find(ctx, coreuser.Filter{
		SystemID: &params.SystemId,
	})
	if err != nil {
		return nil, err
	}

	// Find user with matching UID
	var targetUser *coreuser.SystemUser
	for _, u := range users {
		if u.UID() == uid {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return &apiv1.DeleteSystemUserNotFound{}, nil
	}

	if err := h.sysUserSvc.Delete(ctx, targetUser.ID()); err != nil {
		return nil, err
	}

	return &apiv1.DeleteSystemUserNoContent{}, nil
}

func (h *Handler) toSystemUser(u *coreuser.SystemUser) *apiv1.SystemUser {
	return &apiv1.SystemUser{
		ID:       int64(u.ID()),
		UserId:   u.UserID(),
		SystemId: u.SystemID(),
		Username: u.Username(),
		UID:      int32(u.UID()),
		Gid:      int32(u.GID()),
	}
}
