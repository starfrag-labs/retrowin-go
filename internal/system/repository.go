package system

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for system data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*System, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*System, error)
	GetByName(ctx context.Context, client *ent.Client, name string) (*System, error)
	Update(ctx context.Context, client *ent.Client, cmd *UpdateCommand) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter Filter) ([]*System, error)
	FindOne(ctx context.Context, client *ent.Client, filter Filter) (*System, error)
	Exists(ctx context.Context, client *ent.Client, filter Filter) (bool, error)
}
