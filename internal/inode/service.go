package inode

import (
	"context"
	"time"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// CreateCommand for creating a new inode.
type CreateCommand struct {
	SystemID   *int64
	FileType   FileType
	OwnerUID   string
	OwnerGID   string
	PermOwner  string
	PermGroup  string
	PermOthers string
	IsSystem   bool
	SystemType *string
}

// UpdateCommand for updating an inode.
type UpdateCommand struct {
	ID          int64
	ByteSize    *int64
	PermOwner   *string
	PermGroup   *string
	PermOthers  *string
	LinkCount   *int16
	AccessedAt  *time.Time
}

// Filter for querying inodes.
type Filter struct {
	ID         *int64
	SystemID   *int64
	OwnerUID   *string
	FileType   *FileType
	IsSystem   *bool
	SystemType *string
}

// Filter helpers
func ByID(id int64) Filter {
	return Filter{ID: &id}
}

func BySystemID(systemID int64) Filter {
	return Filter{SystemID: &systemID}
}

func ByOwner(ownerUID string) Filter {
	return Filter{OwnerUID: &ownerUID}
}

func ByOwnerAndSystemType(ownerUID, systemType string) Filter {
	isSystem := true
	return Filter{
		OwnerUID:   &ownerUID,
		IsSystem:   &isSystem,
		SystemType: &systemType,
	}
}

// Service defines the interface for inode operations.
type Service interface {
	Create(ctx context.Context, cmd *CreateCommand) (*Inode, error)
	GetByID(ctx context.Context, id int64) (*Inode, error)
	Update(ctx context.Context, cmd *UpdateCommand) error
	Delete(ctx context.Context, id int64) error
	Find(ctx context.Context, filter Filter) ([]*Inode, error)
	FindOne(ctx context.Context, filter Filter) (*Inode, error)
	FindByOwnerAndSystemType(ctx context.Context, ownerUID, systemType string) (*Inode, error)
	UpdateLinkCount(ctx context.Context, id int64, delta int16) error
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
	// Set defaults
	if cmd.PermOwner == "" {
		if cmd.FileType == FileTypeDirectory {
			cmd.PermOwner = "rwx"
		} else {
			cmd.PermOwner = "rw-"
		}
	}
	if cmd.PermGroup == "" {
		cmd.PermGroup = "r-x"
	}
	if cmd.PermOthers == "" {
		cmd.PermOthers = "r--"
	}
	return s.repo.Create(ctx, s.client, cmd)
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
	return s.repo.Find(ctx, s.client, filter)
}

func (s *service) FindOne(ctx context.Context, filter Filter) (*Inode, error) {
	return s.repo.FindOne(ctx, s.client, filter)
}

func (s *service) FindByOwnerAndSystemType(ctx context.Context, ownerUID, systemType string) (*Inode, error) {
	filter := ByOwnerAndSystemType(ownerUID, systemType)
	inode, err := s.repo.FindOne(ctx, s.client, filter)
	if err != nil {
		return nil, err
	}
	if inode == nil {
		return nil, errors.NotFound("inode not found")
	}
	return inode, nil
}

func (s *service) UpdateLinkCount(ctx context.Context, id int64, delta int16) error {
	return s.repo.UpdateLinkCount(ctx, s.client, id, delta)
}
