package symlink

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/symlink"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*Symlink, error) {
	entSymlink, err := client.Symlink.Create().
		SetInodeID(cmd.InodeID).
		SetTargetPath(cmd.TargetPath).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create symlink: %w", err)
	}
	return fromEnt(entSymlink), nil
}

func (r *EntRepository) GetByInodeID(ctx context.Context, client *ent.Client, inodeID int64) (*Symlink, error) {
	entSymlink, err := client.Symlink.Query().
		Where(symlink.InodeIDEQ(inodeID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get symlink: %w", err)
	}
	return fromEnt(entSymlink), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, cmd *UpdateCommand) error {
	return client.Symlink.Update().
		Where(symlink.InodeIDEQ(cmd.InodeID)).
		SetTargetPath(cmd.TargetPath).
		Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, inodeID int64) error {
	_, err := client.Symlink.Delete().
		Where(symlink.InodeIDEQ(inodeID)).
		Exec(ctx)
	return err
}

func fromEnt(e *ent.Symlink) *Symlink {
	return &Symlink{
		InodeID:    e.InodeID,
		TargetPath: e.TargetPath,
	}
}
