package system

import (
	"context"
)

// SystemRepository defines the interface for system data access.
type SystemRepository interface {
	Create(ctx context.Context, system *System) (*System, error)
	GetByID(ctx context.Context, id string) (*System, error)
	GetByName(ctx context.Context, name string) (*System, error)
	Update(ctx context.Context, system *System) error
	Delete(ctx context.Context, id string) error
	Find(ctx context.Context, filter *QueryFilter) ([]*System, error)
	FindOne(ctx context.Context, filter *QueryFilter) (*System, error)
	Exists(ctx context.Context, filter *QueryFilter) (bool, error)
}

// SystemUserRepository defines the interface for system-user data access.
type SystemUserRepository interface {
	Create(ctx context.Context, systemUser *SystemUser) (*SystemUser, error)
	GetByID(ctx context.Context, id int) (*SystemUser, error)
	Delete(ctx context.Context, id int) error
	Find(ctx context.Context, filter *SystemUserQueryFilter) ([]*SystemUser, error)
	FindOne(ctx context.Context, filter *SystemUserQueryFilter) (*SystemUser, error)
}

// QueryFilter for querying systems (repository layer).
type QueryFilter struct {
	ID     *string
	Name   *string
	Status *Status
}

// SystemUserQueryFilter for querying system-users (repository layer).
type SystemUserQueryFilter struct {
	ID       *int
	UserID   *string
	SystemID *string
	Username *string
}
