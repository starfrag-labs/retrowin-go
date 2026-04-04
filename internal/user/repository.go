package user

import (
	"context"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByProvider(ctx context.Context, provider, providerID string) (*User, error)
	Delete(ctx context.Context, id string) error
	ExistsByProvider(ctx context.Context, provider, providerID string) (bool, error)
}
