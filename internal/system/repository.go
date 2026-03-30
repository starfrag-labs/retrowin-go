package system

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// SystemRepository defines the interface for system data access.
type SystemRepository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*System, error)
	GetByID(ctx context.Context, client *ent.Client, id string) (*System, error)
	GetByName(ctx context.Context, client *ent.Client, name string) (*System, error)
	Update(ctx context.Context, client *ent.Client, params *UpdateParams) error
	Delete(ctx context.Context, client *ent.Client, id string) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*System, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*System, error)
	Exists(ctx context.Context, client *ent.Client, filter *QueryFilter) (bool, error)
}

// SystemUserRepository defines the interface for system-user data access.
type SystemUserRepository interface {
	Create(ctx context.Context, client *ent.Client, params *SystemUserCreateParams) (*SystemUser, error)
	GetByID(ctx context.Context, client *ent.Client, id int) (*SystemUser, error)
	Delete(ctx context.Context, client *ent.Client, id int) error
	Find(ctx context.Context, client *ent.Client, filter *SystemUserQueryFilter) ([]*SystemUser, error)
	FindOne(ctx context.Context, client *ent.Client, filter *SystemUserQueryFilter) (*SystemUser, error)
}

// CreateParams for creating a system (repository layer).
type CreateParams struct {
	Name        string
	Description *string
	Status      Status
}

// UpdateParams for updating a system (repository layer).
type UpdateParams struct {
	ID          string
	Name        *string
	Description *string
	Status      *Status
}

// QueryFilter for querying systems (repository layer).
type QueryFilter struct {
	ID     *string
	Name   *string
	Status *Status
}

// SystemUserCreateParams for creating a system-user (repository layer).
type SystemUserCreateParams struct {
	UserID   string
	SystemID string
	Username string
	UID      int
}

// SystemUserQueryFilter for querying system-users (repository layer).
type SystemUserQueryFilter struct {
	ID       *int
	UserID   *string
	SystemID *string
	Username *string
}
