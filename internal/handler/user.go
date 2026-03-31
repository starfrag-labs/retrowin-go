package handler

import (
	"context"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/middleware"
	extuser "github.com/starfrag-lab/retrowin-go/internal/user"
)

// GetUser implements GET /user.
func (h *Handler) GetUser(ctx context.Context) (apiv1.GetUserRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		return &apiv1.GetUserUnauthorized{}, nil
	}

	u, err := h.extUserSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.GetUserNotFound{}, nil
		}
		return &apiv1.GetUserUnauthorized{}, nil
	}

	return &apiv1.UserResponse{
		User: *h.toExtUser(u),
	}, nil
}

// CreateUser implements POST /user.
func (h *Handler) CreateUser(ctx context.Context, req *apiv1.CreateUserRequest) (apiv1.CreateUserRes, error) {
	cmd := &extuser.CreateCommand{
		Provider:   string(req.Provider),
		ProviderID: req.ProviderId,
	}

	u, err := h.extUserSvc.Create(ctx, cmd)
	if err != nil {
		if errors.IsConflict(err) {
			return &apiv1.CreateUserConflict{}, nil
		}
		return &apiv1.CreateUserBadRequest{}, nil
	}

	return &apiv1.UserResponse{
		User: *h.toExtUser(u),
	}, nil
}

// DeleteUser implements DELETE /user.
func (h *Handler) DeleteUser(ctx context.Context) (apiv1.DeleteUserRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	// Get user first to obtain provider info for deletion
	u, err := h.extUserSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.DeleteUserNotFound{}, nil
		}
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	err = h.extUserSvc.Delete(ctx, u.Provider(), u.ProviderID())
	if err != nil {
		if errors.IsNotFound(err) {
			return &apiv1.DeleteUserNotFound{}, nil
		}
		return &apiv1.DeleteUserUnauthorized{}, nil
	}

	return &apiv1.DeleteUserNoContent{}, nil
}

func (h *Handler) toExtUser(u *extuser.User) *apiv1.User {
	return &apiv1.User{
		ID:         u.ID(),
		Provider:   apiv1.Provider(u.Provider()),
		ProviderId: u.ProviderID(),
		JoinDate:   toOptTimestamp(u.JoinDate()),
		CreatedAt:  toOptTimestamp(u.CreatedAt()),
		UpdatedAt:  toOptTimestamp(u.UpdatedAt()),
	}
}
