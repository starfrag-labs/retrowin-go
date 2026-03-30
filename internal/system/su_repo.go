package system

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// SystemUserRepository defines the interface for system-user data access.
type SystemUserRepository interface {
	Create(ctx context.Context, client *ent.Client, params *SystemUserCreateParams) (*SystemUser, error)
	GetByID(ctx context.Context, client *ent.Client, id int) (*SystemUser, error)
	Delete(ctx context.Context, client *ent.Client, id int) error
	Find(ctx context.Context, client *ent.Client, filter *SystemUserQueryFilter) ([]*SystemUser, error)
	FindOne(ctx context.Context, client *ent.Client, filter *SystemUserQueryFilter) (*SystemUser, error)
}

// SystemUserCreateParams for creating a system-user (repository layer).
type SystemUserCreateParams struct {
	UserID   string
	SystemID string
	Username string
}

// SystemUserQueryFilter for querying system-users (repository layer).
type SystemUserQueryFilter struct {
	ID       *int
	UserID   *string
	SystemID *string
	Username *string
}