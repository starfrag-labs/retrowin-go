package user

import (
	"context"
)

// SystemUserRepository defines the interface for system-user data access.
type SystemUserRepository interface {
	Create(ctx context.Context, systemUser *SystemUser) (*SystemUser, error)
	GetByID(ctx context.Context, id int) (*SystemUser, error)
	Delete(ctx context.Context, id int) error
	Find(ctx context.Context, filter *QueryFilter) ([]*SystemUser, error)
	FindOne(ctx context.Context, filter *QueryFilter) (*SystemUser, error)
	// GetNextUID returns the next available UID for the system.
	GetNextUID(ctx context.Context, systemID string) (int, error)
}

// SystemGroupRepository defines the interface for system-group data access.
type SystemGroupRepository interface {
	Create(ctx context.Context, group *SystemGroup) (*SystemGroup, error)
	GetByID(ctx context.Context, id int) (*SystemGroup, error)
	Delete(ctx context.Context, id int) error
	Find(ctx context.Context, filter *GroupQueryFilter) ([]*SystemGroup, error)
	FindOne(ctx context.Context, filter *GroupQueryFilter) (*SystemGroup, error)
	// GetNextGID returns the next available GID for the system.
	GetNextGID(ctx context.Context, systemID string) (int, error)
	// Group membership operations
	AddUserToGroup(ctx context.Context, userSystemID, groupID int) error
	RemoveUserFromGroup(ctx context.Context, userSystemID, groupID int) error
	FindGIDsByUserSystemID(ctx context.Context, userSystemID int) ([]int, error)
}

// QueryFilter for querying system-users (repository layer).
type QueryFilter struct {
	ID       *int
	UserID   *string
	SystemID *string
	Username *string
	UID      *int
}

// GroupQueryFilter for querying system-groups (repository layer).
type GroupQueryFilter struct {
	ID       *int
	SystemID *string
	Name     *string
	GID      *int
}
