package group

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for group data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Group, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*Group, error)
	GetBySystemIDAndGID(ctx context.Context, client *ent.Client, systemID int64, gid string) (*Group, error)
	GetBySystemIDAndGroupname(ctx context.Context, client *ent.Client, systemID int64, groupname string) (*Group, error)
	Update(ctx context.Context, client *ent.Client, params *UpdateParams) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*Group, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*Group, error)
	Exists(ctx context.Context, client *ent.Client, filter *QueryFilter) (bool, error)
}

// CreateParams for creating a group (repository layer).
type CreateParams struct {
	SystemID  int64
	GID       string
	Groupname string
}

// UpdateParams for updating a group (repository layer).
type UpdateParams struct {
	ID        int64
	Groupname *string
}

// QueryFilter for querying groups (repository layer).
type QueryFilter struct {
	ID        *int64
	SystemID  *int64
	GID       *string
	Groupname *string
}
