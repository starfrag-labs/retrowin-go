package directory

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for directory entry data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*Entry, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*Entry, error)
	Update(ctx context.Context, client *ent.Client, entry *Entry) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter Filter) ([]*Entry, error)
	FindOne(ctx context.Context, client *ent.Client, filter Filter) (*Entry, error)
	Exists(ctx context.Context, client *ent.Client, filter Filter) (bool, error)
}
