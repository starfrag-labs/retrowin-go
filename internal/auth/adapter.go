package auth

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// UserAdapter adapts user.Service to auth.UserService interface.
type UserAdapter struct {
	userSvc   user.Service
	provider  string
}

// NewUserAdapter creates a new user adapter.
func NewUserAdapter(userSvc user.Service, provider string) *UserAdapter {
	return &UserAdapter{
		userSvc:  userSvc,
		provider: provider,
	}
}

// FindOrCreate finds an existing user by OIDC subject or creates a new one.
func (a *UserAdapter) FindOrCreate(ctx context.Context, subject, email, name, picture string) (int64, error) {
	// Find or create user by provider subject
	userID, err := a.userSvc.FindOrCreateByOIDC(ctx, a.provider, subject, email, name, picture)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
