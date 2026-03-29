package group

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for group data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*Group, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*Group, error)
	GetBySystemIDAndGID(ctx context.Context, client *ent.Client, systemID int64, gid string) (*Group, error)
	GetBySystemIDAndGroupname(ctx context.Context, client *ent.Client, systemID int64, groupname string) (*Group, error)
	Update(ctx context.Context, client *ent.Client, cmd *UpdateCommand) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter Filter) ([]*Group, error)
	FindOne(ctx context.Context, client *ent.Client, filter Filter) (*Group, error)
	Exists(ctx context.Context, client *ent.Client, filter Filter) (bool, error)
}
