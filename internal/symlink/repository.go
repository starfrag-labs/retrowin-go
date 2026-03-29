package symlink

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for symlink data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*Symlink, error)
	GetByInodeID(ctx context.Context, client *ent.Client, inodeID int64) (*Symlink, error)
	Update(ctx context.Context, client *ent.Client, cmd *UpdateCommand) error
	Delete(ctx context.Context, client *ent.Client, inodeID int64) error
}
