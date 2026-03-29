package auth

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// userService adapts user.Service to auth.UserService interface.
type userService struct {
	userSvc user.Service
}

// NewUserService creates a new user adapter.
func NewUserService(userSvc user.Service) UserService {
	return &userService{
		userSvc: userSvc,
	}
}

// FindOrCreate finds an existing user by OIDC subject or creates a new one.
func (a *userService) FindOrCreate(ctx context.Context, subject, email, name, picture string) (int64, string, error) {
	// Find or create user by provider subject (always use keycloak)
	userID, userUID, err := a.userSvc.FindOrCreateByOIDC(ctx, "keycloak", subject, email, name, picture)
	if err != nil {
		return 0, "", err
	}
	return userID, userUID, nil
}
