package inode

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Service defines the interface for inode operations.
type Service interface {
	Create(ctx context.Context, cmd *CreateCommand) (*Inode, error)
	GetByID(ctx context.Context, id int64) (*Inode, error)
	Update(ctx context.Context, cmd *UpdateCommand) error
	Delete(ctx context.Context, id int64) error
	Find(ctx context.Context, filter Filter) ([]*Inode, error)
	FindOne(ctx context.Context, filter Filter) (*Inode, error)
	UpdateLinkCount(ctx context.Context, id int64, delta int) error
}

// CreateCommand for creating a new inode (service layer).
type CreateCommand struct {
	SystemID string
	Mode     int
	UID      int64
	GID      int64
	Flags    int
	Content  []byte
}

// UpdateCommand for updating an inode (service layer).
type UpdateCommand = UpdateParams

// Filter for querying inodes (service layer).
type Filter = QueryFilter

// Filter helpers
func ByID(id int64) Filter {
	return Filter{ID: &id}
}

func BySystemID(systemID string) Filter {
	return Filter{SystemID: &systemID}
}

func ByUID(uid int64) Filter {
	return Filter{UID: &uid}
}

type service struct {
	repo   Repository
	client *ent.Client
}

// NewService creates a new Service.
func NewService(repo Repository, client *ent.Client) Service {
	return &service{repo: repo, client: client}
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*Inode, error) {
	params := &CreateParams{
		SystemID: cmd.SystemID,
		Mode:     cmd.Mode,
		UID:      cmd.UID,
		GID:      cmd.GID,
		Flags:    cmd.Flags,
		Content:  cmd.Content,
	}
	return s.repo.Create(ctx, s.client, params)
}

func (s *service) GetByID(ctx context.Context, id int64) (*Inode, error) {
	inode, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if inode == nil {
		return nil, errors.NotFound("inode not found")
	}
	return inode, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCommand) error {
	return s.repo.Update(ctx, s.client, cmd)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, s.client, id)
}

func (s *service) Find(ctx context.Context, filter Filter) ([]*Inode, error) {
	return s.repo.Find(ctx, s.client, &filter)
}

func (s *service) FindOne(ctx context.Context, filter Filter) (*Inode, error) {
	inode, err := s.repo.FindOne(ctx, s.client, &filter)
	if err != nil {
		return nil, err
	}
	if inode == nil {
		return nil, errors.NotFound("inode not found")
	}
	return inode, nil
}

func (s *service) UpdateLinkCount(ctx context.Context, id int64, delta int) error {
	return s.repo.UpdateLinkCount(ctx, s.client, id, delta)
}