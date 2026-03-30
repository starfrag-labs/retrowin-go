package auth

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/internal/user"
)

type User struct {
	user.User
}

// UserService defines the interface for user operations.
type UserService interface {
	// FindOrCreate finds an existing user by OIDC subject or creates a new one.
	// Returns userID (string).
	FindOrCreate(ctx context.Context, subject, email, name, picture string) (string, error)
}
