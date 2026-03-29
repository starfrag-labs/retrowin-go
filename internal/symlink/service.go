package symlink

import (
	"context"
	"errors"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Errors
var (
	ErrNotFound = errors.New("symlink not found")
)

// CreateCommand for creating a symlink.
type CreateCommand struct {
	InodeID    int64
	TargetPath string
}

// UpdateCommand for updating a symlink.
type UpdateCommand struct {
	InodeID    int64
	TargetPath string
}

// Service defines the interface for symlink operations.
type Service interface {
	Create(ctx context.Context, cmd *CreateCommand) (*Symlink, error)
	GetByInodeID(ctx context.Context, inodeID int64) (*Symlink, error)
	Update(ctx context.Context, cmd *UpdateCommand) error
	Delete(ctx context.Context, inodeID int64) error
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
	return s.repo.Create(ctx, s.client, cmd)
}

func (s *service) GetByInodeID(ctx context.Context, inodeID int64) (*Symlink, error) {
	sl, err := s.repo.GetByInodeID(ctx, s.client, inodeID)
	if err != nil {
		return nil, err
	}
	if sl == nil {
		return nil, ErrNotFound
	}
	return sl, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCommand) error {
	return s.repo.Update(ctx, s.client, cmd)
}

func (s *service) Delete(ctx context.Context, inodeID int64) error {
	return s.repo.Delete(ctx, s.client, inodeID)
}
