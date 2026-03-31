package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/core/user"
)

// CreateSystemUser implements POST /systems/{systemId}/users.
func (h *Handler) CreateSystemUser(ctx context.Context, req *apiv1.CreateSystemUserRequest, params apiv1.CreateSystemUserParams) (apiv1.CreateSystemUserRes, error) {
	cmd := &user.CreateCommand{
		UserID:   req.UserId,
		SystemID: params.SystemId,
		Username: req.Username,
	}
	if req.Uid.Set {
		cmd.UID = int(req.Uid.Value)
	}

	sysUser, err := h.userSvc.Create(ctx, cmd)
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "create_system_user_failed",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.SystemUserResponse{
		User: *h.toSystemUser(sysUser),
	}, nil
}

// ListSystemUsers implements GET /systems/{systemId}/users.
func (h *Handler) ListSystemUsers(ctx context.Context, params apiv1.ListSystemUsersParams) (apiv1.ListSystemUsersRes, error) {
	users, err := h.userSvc.Find(ctx, user.BySystemID(params.SystemId))
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "list_system_users_failed",
				Message: err.Error(),
			},
		}, nil
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
	sysUser, err := h.userSvc.FindOne(ctx, user.Filter{
		SystemID: &params.SystemId,
		UID:      &params.Uid,
	})
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "get_system_user_failed",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.SystemUserResponse{
		User: *h.toSystemUser(sysUser),
	}, nil
}

// DeleteSystemUser implements DELETE /systems/{systemId}/users/{uid}.
func (h *Handler) DeleteSystemUser(ctx context.Context, params apiv1.DeleteSystemUserParams) (apiv1.DeleteSystemUserRes, error) {
	// First find the user by systemID and UID
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

	if err := h.userSvc.Delete(ctx, sysUser.ID()); err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "delete_system_user_failed",
				Message: err.Error(),
			},
		}, nil
	}

	return &apiv1.DeleteSystemUserNoContent{}, nil
}

func (h *Handler) toSystemUser(u *user.SystemUser) *apiv1.SystemUser {
	return &apiv1.SystemUser{
		Id:        int64(u.ID()),
		UserId:    u.UserID(),
		SystemId:  u.SystemID(),
		Username:  u.Username(),
		Uid:       u.UID(),
		Gid:       u.GID(),
		CreatedAt: toOptTimestamp(u.CreatedAt()),
	}
}
