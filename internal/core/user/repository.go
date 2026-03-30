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
}

// CreateParams for creating a system-user (repository layer).
type CreateParams struct {
	UserID   string
	SystemID string
	Username string
	UID      int
}

// QueryFilter for querying system-users (repository layer).
type QueryFilter struct {
	ID       *int
	UserID   *string
	SystemID *string
	Username *string
}
