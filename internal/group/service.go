package group

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

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

// CreateCommand for creating a group (service layer).
type CreateCommand struct {
	SystemID  int64
	GID       string
	Groupname string
}

// UpdateCommand for updating a group (service layer).
type UpdateCommand struct {
	ID        int64
	Groupname *string
}

// Filter for querying groups (service layer).
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

// toQueryFilter converts service Filter to repository QueryFilter.
func (f Filter) toQueryFilter() *QueryFilter {
	return &QueryFilter{
		ID:        f.ID,
		SystemID:  f.SystemID,
		GID:       f.GID,
		Groupname: f.Groupname,
	}
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
	params := &CreateParams{
		SystemID:  cmd.SystemID,
		GID:       cmd.GID,
		Groupname: cmd.Groupname,
	}
	return s.repo.Create(ctx, s.client, params)
}

func (s *service) GetByID(ctx context.Context, id int64) (*Group, error) {
	g, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.NotFound("group not found")
	}
	return g, nil
}

func (s *service) GetBySystemIDAndGID(ctx context.Context, systemID int64, gid string) (*Group, error) {
	g, err := s.repo.GetBySystemIDAndGID(ctx, s.client, systemID, gid)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.NotFound("group not found")
	}
	return g, nil
}

func (s *service) GetBySystemIDAndGroupname(ctx context.Context, systemID int64, groupname string) (*Group, error) {
	g, err := s.repo.GetBySystemIDAndGroupname(ctx, s.client, systemID, groupname)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.NotFound("group not found")
	}
	return g, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCommand) error {
	params := &UpdateParams{
		ID:        cmd.ID,
		Groupname: cmd.Groupname,
	}
	return s.repo.Update(ctx, s.client, params)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, s.client, id)
}

func (s *service) Find(ctx context.Context, filter Filter) ([]*Group, error) {
	return s.repo.Find(ctx, s.client, filter.toQueryFilter())
}

func (s *service) FindOne(ctx context.Context, filter Filter) (*Group, error) {
	return s.repo.FindOne(ctx, s.client, filter.toQueryFilter())
}
