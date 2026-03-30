package directory

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Service defines the interface for directory entry operations.
type Service interface {
	Create(ctx context.Context, cmd *CreateCommand) (*Entry, error)
	GetByID(ctx context.Context, id int64) (*Entry, error)
	Update(ctx context.Context, cmd *UpdateCommand) error
	Delete(ctx context.Context, id int64) error
	Find(ctx context.Context, filter Filter) ([]*Entry, error)
	FindOne(ctx context.Context, filter Filter) (*Entry, error)
	FindByParentAndName(ctx context.Context, parentID int64, name string) (*Entry, error)
	FindByParent(ctx context.Context, parentID int64) ([]*Entry, error)
	FindByChild(ctx context.Context, childID int64) ([]*Entry, error)
	Exists(ctx context.Context, parentID int64, name string) (bool, error)
}

// CreateCommand for creating a directory entry (service layer).
type CreateCommand struct {
	ParentID int64
	Name     string
	ChildID  int64
}

// UpdateCommand for updating a directory entry (service layer).
type UpdateCommand struct {
	ID       int64
	ParentID *int64
	Name     *string
	ChildID  *int64
}

// Filter for querying directory entries (service layer).
type Filter struct {
	ID       *int64
	ParentID *int64
	Name     *string
	ChildID  *int64
}

// Filter helpers
func ByID(id int64) Filter {
	return Filter{ID: &id}
}

func ByParentAndName(parentID int64, name string) Filter {
	return Filter{ParentID: &parentID, Name: &name}
}

func ByParent(parentID int64) Filter {
	return Filter{ParentID: &parentID}
}

func ByChild(childID int64) Filter {
	return Filter{ChildID: &childID}
}

// toQueryFilter converts service Filter to repository QueryFilter.
func (f Filter) toQueryFilter() *QueryFilter {
	return &QueryFilter{
		ID:       f.ID,
		ParentID: f.ParentID,
		Name:     f.Name,
		ChildID:  f.ChildID,
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

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*Entry, error) {
	params := &CreateParams{
		ParentID: cmd.ParentID,
		Name:     cmd.Name,
		ChildID:  cmd.ChildID,
	}
	return s.repo.Create(ctx, s.client, params)
}

func (s *service) GetByID(ctx context.Context, id int64) (*Entry, error) {
	entry, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, errors.NotFound("directory entry not found")
	}
	return entry, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCommand) error {
	params := &UpdateParams{
		ID:       cmd.ID,
		ParentID: cmd.ParentID,
		Name:     cmd.Name,
		ChildID:  cmd.ChildID,
	}
	return s.repo.Update(ctx, s.client, params)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, s.client, id)
}

func (s *service) Find(ctx context.Context, filter Filter) ([]*Entry, error) {
	return s.repo.Find(ctx, s.client, filter.toQueryFilter())
}

func (s *service) FindOne(ctx context.Context, filter Filter) (*Entry, error) {
	return s.repo.FindOne(ctx, s.client, filter.toQueryFilter())
}

func (s *service) FindByParentAndName(ctx context.Context, parentID int64, name string) (*Entry, error) {
	return s.repo.FindOne(ctx, s.client, ByParentAndName(parentID, name).toQueryFilter())
}

func (s *service) FindByParent(ctx context.Context, parentID int64) ([]*Entry, error) {
	return s.repo.Find(ctx, s.client, ByParent(parentID).toQueryFilter())
}

func (s *service) FindByChild(ctx context.Context, childID int64) ([]*Entry, error) {
	return s.repo.Find(ctx, s.client, ByChild(childID).toQueryFilter())
}

func (s *service) Exists(ctx context.Context, parentID int64, name string) (bool, error) {
	return s.repo.Exists(ctx, s.client, ByParentAndName(parentID, name).toQueryFilter())
}
