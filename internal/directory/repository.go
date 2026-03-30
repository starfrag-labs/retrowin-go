package directory

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for directory entry data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Entry, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*Entry, error)
	Update(ctx context.Context, client *ent.Client, params *UpdateParams) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*Entry, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*Entry, error)
	Exists(ctx context.Context, client *ent.Client, filter *QueryFilter) (bool, error)
}

// CreateParams for creating a directory entry (repository layer).
type CreateParams struct {
	ParentID int64
	Name     string
	ChildID  int64
}

// UpdateParams for updating a directory entry (repository layer).
type UpdateParams struct {
	ID       int64
	ParentID *int64
	Name     *string
	ChildID  *int64
}

// QueryFilter for querying directory entries (repository layer).
type QueryFilter struct {
	ID       *int64
	ParentID *int64
	Name     *string
	ChildID  *int64
}
