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
func (a *userService) FindOrCreate(ctx context.Context, subject, email, name, picture string) (string, error) {
	// Find or create user by provider subject (always use keycloak)
	// Use name as username, fallback to email
	username := name
	if username == "" {
		username = email
	}
	return a.userSvc.FindOrCreateByOIDC(ctx, "keycloak", subject, username)
}
