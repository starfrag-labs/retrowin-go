package group

import (
	"context"
	"errors"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Errors
var (
	ErrNotFound      = errors.New("group not found")
	ErrAlreadyExists = errors.New("group already exists")
)

// CreateCommand for creating a group.
type CreateCommand struct {
	SystemID  int64
	GID       string
	Groupname string
}

// UpdateCommand for updating a group.
type UpdateCommand struct {
	ID        int64
	Groupname *string
}

// Filter for querying groups.
type Filter struct {
	ID        *int64
	SystemID  *int64
	GID       *string
	Groupname *string
}

// Filter helpers
func ByID(id int64) Filter {
	return Filter{ID: &id}
}

func BySystemID(systemID int64) Filter {
	return Filter{SystemID: &systemID}
}

func BySystemIDAndGID(systemID int64, gid string) Filter {
	return Filter{SystemID: &systemID, GID: &gid}
}

func BySystemIDAndGroupname(systemID int64, groupname string) Filter {
	return Filter{SystemID: &systemID, Groupname: &groupname}
}

// Service defines the interface for group operations.
type Service interface {
	Create(ctx context.Context, cmd *CreateCommand) (*Group, error)
	GetByID(ctx context.Context, id int64) (*Group, error)
	GetBySystemIDAndGID(ctx context.Context, systemID int64, gid string) (*Group, error)
	GetBySystemIDAndGroupname(ctx context.Context, systemID int64, groupname string) (*Group, error)
	Update(ctx context.Context, cmd *UpdateCommand) error
	Delete(ctx context.Context, id int64) error
	Find(ctx context.Context, filter Filter) ([]*Group, error)
	FindOne(ctx context.Context, filter Filter) (*Group, error)
}

type service struct {
	repo   Repository
	client *ent.Client
}

// NewService creates a new Service.
func NewService(repo Repository, client *ent.Client) Service {
	return &service{repo: repo, client: client}
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*Group, error) {
	return s.repo.Create(ctx, s.client, cmd)
}

func (s *service) GetByID(ctx context.Context, id int64) (*Group, error) {
	g, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrNotFound
	}
	return g, nil
}

func (s *service) GetBySystemIDAndGID(ctx context.Context, systemID int64, gid string) (*Group, error) {
	g, err := s.repo.GetBySystemIDAndGID(ctx, s.client, systemID, gid)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrNotFound
	}
	return g, nil
}

func (s *service) GetBySystemIDAndGroupname(ctx context.Context, systemID int64, groupname string) (*Group, error) {
	g, err := s.repo.GetBySystemIDAndGroupname(ctx, s.client, systemID, groupname)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrNotFound
	}
	return g, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCommand) error {
	return s.repo.Update(ctx, s.client, cmd)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, s.client, id)
}

func (s *service) Find(ctx context.Context, filter Filter) ([]*Group, error) {
	return s.repo.Find(ctx, s.client, filter)
}

func (s *service) FindOne(ctx context.Context, filter Filter) (*Group, error) {
	return s.repo.FindOne(ctx, s.client, filter)
}
