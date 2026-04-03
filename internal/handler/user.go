package handler

import (
	"context"

	api "github.com/starfrag-lab/retrowin-go/pkg/api"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/middleware"
	extuser "github.com/starfrag-lab/retrowin-go/internal/user"
)

// GetUser implements GET /user.
func (h *Handler) GetUser(ctx context.Context) (api.GetUserRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		return &api.GetUserUnauthorized{}, nil
	}

	u, err := h.extUserSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &api.GetUserNotFound{}, nil
		}
		return &api.GetUserUnauthorized{}, nil
	}

	return &api.UserResponse{
		User: *h.toExtUser(u),
	}, nil
}

// DeleteUser implements DELETE /user.
func (h *Handler) DeleteUser(ctx context.Context) (api.DeleteUserRes, error) {
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		return &api.DeleteUserUnauthorized{}, nil
	}

	// Get user first to obtain provider info for deletion
	u, err := h.extUserSvc.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &api.DeleteUserNotFound{}, nil
		}
		return &api.DeleteUserUnauthorized{}, nil
	}

	err = h.extUserSvc.Delete(ctx, u.Provider(), u.ProviderID())
	if err != nil {
		if errors.IsNotFound(err) {
			return &api.DeleteUserNotFound{}, nil
		}
		return &api.DeleteUserUnauthorized{}, nil
	}

	return &api.DeleteUserNoContent{}, nil
}

func (h *Handler) toExtUser(u *extuser.User) *api.User {
	return &api.User{
		ID:         u.ID(),
		Provider:   api.Provider(u.Provider()),
		ProviderId: u.ProviderID(),
		JoinDate:   toOptTimestamp(u.JoinDate()),
		CreatedAt:  toOptTimestamp(u.CreatedAt()),
		UpdatedAt:  toOptTimestamp(u.UpdatedAt()),
	}
}
