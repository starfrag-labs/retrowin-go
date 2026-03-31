package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/middleware"
	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// GetUser implements GET /user.
func (h *Handler) GetUser(ctx context.Context) (apiv1.GetUserRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		return &apiv1.GetUserUnauthorized{}, nil
	}

	u, err := h.userSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetUserNotFound{}, nil
		}
		return &apiv1.GetUserUnauthorized{}, nil
	}

	return &apiv1.UserResponse{
		User: *h.toUser(u),
	}, nil
}

// CreateUser implements POST /user.
func (h *Handler) CreateUser(ctx context.Context, req *apiv1.CreateUserRequest) (apiv1.CreateUserRes, error) {
	cmd := &user.CreateCommand{
		Provider:   string(req.Provider),
		ProviderID: req.ProviderId,
	}

	u, err := h.userSvc.Create(ctx, cmd)
	if err != nil {
		if errors.IsConflict(err) {
			return &apiv1.CreateUserConflict{}, nil
		}
		return &apiv1.CreateUserBadRequest{}, nil
	}

	return &apiv1.UserResponse{
		User: *h.toUser(u),
	}, nil
}

// DeleteUser implements DELETE /user.
func (h *Handler) DeleteUser(ctx context.Context) (apiv1.DeleteUserRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	// Get user first to obtain provider info for deletion
	u, err := h.userSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.DeleteUserNotFound{}, nil
		}
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	err = h.userSvc.Delete(ctx, u.Provider, u.ProviderID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.DeleteUserNotFound{}, nil
		}
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	return &apiv1.DeleteUserNoContent{}, nil
}
