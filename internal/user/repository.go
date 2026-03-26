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

// ServiceStatusRepository defines the interface for service status data access.
type ServiceStatusRepository interface {
	// Create creates a service status for a user.
	Create(ctx context.Context, userID int64) (*ServiceStatus, error)

	// GetByUserID retrieves service status by user ID.
	GetByUserID(ctx context.Context, userID int64) (*ServiceStatus, error)

	// Update updates the service status.
	Update(ctx context.Context, userID int64, available bool) (*ServiceStatus, error)

	// Delete deletes service status by user ID.
	Delete(ctx context.Context, userID int64) error
}
