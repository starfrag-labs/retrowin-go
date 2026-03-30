package symlink

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Service defines the interface for symlink operations.
type Service interface {
	Create(ctx context.Context, cmd *CreateCommand) (*Symlink, error)
	GetByInodeID(ctx context.Context, inodeID int64) (*Symlink, error)
	Update(ctx context.Context, cmd *UpdateCommand) error
	Delete(ctx context.Context, inodeID int64) error
}

// CreateCommand for creating a symlink (service layer).
type CreateCommand struct {
	InodeID    int64
	TargetPath string
}

// UpdateCommand for updating a symlink (service layer).
type UpdateCommand struct {
	InodeID    int64
	TargetPath string
}

type service struct {
	repo   Repository
	client *ent.Client
}

// NewService creates a new Service.
func NewService(repo Repository, client *ent.Client) Service {
	return &service{repo: repo, client: client}
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*Symlink, error) {
	params := &CreateParams{
		InodeID:    cmd.InodeID,
		TargetPath: cmd.TargetPath,
	}
	return s.repo.Create(ctx, s.client, params)
}

func (s *service) GetByInodeID(ctx context.Context, inodeID int64) (*Symlink, error) {
	sl, err := s.repo.GetByInodeID(ctx, s.client, inodeID)
	if err != nil {
		return nil, err
	}
	if sl == nil {
		return nil, errors.NotFound("symlink not found")
	}
	return sl, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCommand) error {
	params := &UpdateParams{
		InodeID:    cmd.InodeID,
		TargetPath: cmd.TargetPath,
	}
	return s.repo.Update(ctx, s.client, params)
}

func (s *service) Delete(ctx context.Context, inodeID int64) error {
	return s.repo.Delete(ctx, s.client, inodeID)
}
