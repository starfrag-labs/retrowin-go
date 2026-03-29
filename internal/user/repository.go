package user

import (
	"context"
)

// Repository defines the interface for user data access.
type Repository interface {
	// Create creates a new user.
	Create(ctx context.Context, provider, providerID string) (*User, error)

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByProvider retrieves a user by provider and provider ID.
	GetByProvider(ctx context.Context, provider, providerID string) (*User, error)

	// Delete deletes a user by ID.
	Delete(ctx context.Context, id int64) error

	// ExistsByProvider checks if a user exists by provider and provider ID.
	ExistsByProvider(ctx context.Context, provider, providerID string) (bool, error)
}
