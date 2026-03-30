package system

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for system data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*System, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*System, error)
	GetByName(ctx context.Context, client *ent.Client, name string) (*System, error)
	Update(ctx context.Context, client *ent.Client, params *UpdateParams) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*System, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*System, error)
	Exists(ctx context.Context, client *ent.Client, filter *QueryFilter) (bool, error)
}

// CreateParams for creating a system (repository layer).
type CreateParams struct {
	Name        string
	Description *string
	Status      Status
}

// UpdateParams for updating a system (repository layer).
type UpdateParams struct {
	ID          int64
	Name        *string
	Description *string
	Status      *Status
}

// QueryFilter for querying systems (repository layer).
type QueryFilter struct {
	ID     *int64
	Name   *string
	Status *Status
}
