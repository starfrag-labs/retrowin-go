package system

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// CreateCommand for creating a system.
type CreateCommand struct {
	Name        string
	Description *string
	Status      Status
}

// UpdateCommand for updating a system.
type UpdateCommand struct {
	ID          int64
	Name        *string
	Description *string
	Status      *Status
}

// Filter for querying systems.
type Filter struct {
	ID     *int64
	Name   *string
	Status *Status
}

// Filter helpers
func ByID(id int64) Filter {
	return Filter{ID: &id}
}

func ByName(name string) Filter {
	return Filter{Name: &name}
}

func ByStatus(status Status) Filter {
	return Filter{Status: &status}
}

// Service defines the interface for system operations.
type Service interface {
	Create(ctx context.Context, cmd *CreateCommand) (*System, error)
	GetByID(ctx context.Context, id int64) (*System, error)
	GetByName(ctx context.Context, name string) (*System, error)
	Update(ctx context.Context, cmd *UpdateCommand) error
	Delete(ctx context.Context, id int64) error
	Find(ctx context.Context, filter Filter) ([]*System, error)
	FindOne(ctx context.Context, filter Filter) (*System, error)
}

type service struct {
	repo   Repository
	client *ent.Client
}

// NewService creates a new Service.
func NewService(repo Repository, client *ent.Client) Service {
	return &service{repo: repo, client: client}
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*System, error) {
	if cmd.Status == "" {
		cmd.Status = StatusActive
	}
	return s.repo.Create(ctx, s.client, cmd)
}

func (s *service) GetByID(ctx context.Context, id int64) (*System, error) {
	sys, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if sys == nil {
		return nil, errors.NotFound("system not found")
	}
	return sys, nil
}

func (s *service) GetByName(ctx context.Context, name string) (*System, error) {
	sys, err := s.repo.GetByName(ctx, s.client, name)
	if err != nil {
		return nil, err
	}
	if sys == nil {
		return nil, errors.NotFound("system not found")
	}
	return sys, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCommand) error {
	return s.repo.Update(ctx, s.client, cmd)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, s.client, id)
}

func (s *service) Find(ctx context.Context, filter Filter) ([]*System, error) {
	return s.repo.Find(ctx, s.client, filter)
}

func (s *service) FindOne(ctx context.Context, filter Filter) (*System, error) {
	return s.repo.FindOne(ctx, s.client, filter)
}
