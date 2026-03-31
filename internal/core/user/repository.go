package user

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// SystemUserRepository defines the interface for system-user data access.
type SystemUserRepository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*SystemUser, error)
	GetByID(ctx context.Context, client *ent.Client, id int) (*SystemUser, error)
	Delete(ctx context.Context, client *ent.Client, id int) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*SystemUser, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*SystemUser, error)
	// GetNextUID returns the next available UID for the system.
	GetNextUID(ctx context.Context, client *ent.Client, systemID string) (int, error)
}

// SystemGroupRepository defines the interface for system-group data access.
type SystemGroupRepository interface {
	Create(ctx context.Context, client *ent.Client, params *GroupCreateParams) (*SystemGroup, error)
	GetByID(ctx context.Context, client *ent.Client, id int) (*SystemGroup, error)
	Delete(ctx context.Context, client *ent.Client, id int) error
	Find(ctx context.Context, client *ent.Client, filter *GroupQueryFilter) ([]*SystemGroup, error)
	FindOne(ctx context.Context, client *ent.Client, filter *GroupQueryFilter) (*SystemGroup, error)
	// Group membership operations
	AddUserToGroup(ctx context.Context, client *ent.Client, userSystemID, groupID int) error
	RemoveUserFromGroup(ctx context.Context, client *ent.Client, userSystemID, groupID int) error
	FindGIDsByUserSystemID(ctx context.Context, client *ent.Client, userSystemID int) ([]int, error)
}

// CreateParams for creating a system-user (repository layer).
type CreateParams struct {
	UserID   string
	SystemID string
	Username string
	UID      int
	GID      int
}

// QueryFilter for querying system-users (repository layer).
type QueryFilter struct {
	ID       *int
	UserID   *string
	SystemID *string
	Username *string
}

// GroupCreateParams for creating a system-group (repository layer).
type GroupCreateParams struct {
	SystemID string
	Name     string
	GID      int
}

// GroupQueryFilter for querying system-groups (repository layer).
type GroupQueryFilter struct {
	ID       *int
	SystemID *string
	Name     *string
	GID      *int
}
