package filedata

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Service defines the interface for file data operations.
// This is an internal service, not exposed via API directly.
type Service interface {
	Create(ctx context.Context, cmd *CreateCommand) (*FileData, error)
	GetByInodeID(ctx context.Context, inodeID int64) (*FileData, error)
	Update(ctx context.Context, cmd *UpdateCommand) error
	Delete(ctx context.Context, inodeID int64) error
}

// CreateCommand for creating file data (service layer).
type CreateCommand struct {
	InodeID     int64
	StorageType StorageType
	Location    string
	Checksum    *string
}

// UpdateCommand for updating file data (service layer).
type UpdateCommand struct {
	InodeID     int64
	StorageType *StorageType
	Location    *string
	Checksum    *string
}

type service struct {
	repo   Repository
	client *ent.Client
}

// NewService creates a new file data service.
func NewService(repo Repository, client *ent.Client) Service {
	return &service{
		repo:   repo,
		client: client,
	}
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*FileData, error) {
	if cmd.InodeID <= 0 {
		return nil, errors.BadRequest("inodeId is required")
	}
	if cmd.StorageType == "" {
		return nil, errors.BadRequest("storageType is required")
	}
	if cmd.Location == "" {
		return nil, errors.BadRequest("location is required")
	}

	params := &CreateParams{
		InodeID:     cmd.InodeID,
		StorageType: cmd.StorageType,
		Location:    cmd.Location,
		Checksum:    cmd.Checksum,
	}
	return s.repo.Create(ctx, s.client, params)
}

func (s *service) GetByInodeID(ctx context.Context, inodeID int64) (*FileData, error) {
	fileData, err := s.repo.GetByInodeID(ctx, s.client, inodeID)
	if err != nil {
		return nil, err
	}
	if fileData == nil {
		return nil, errors.NotFound("file data not found")
	}
	return fileData, nil
}

func (s *service) Update(ctx context.Context, cmd *UpdateCommand) error {
	params := &UpdateParams{
		InodeID:     cmd.InodeID,
		StorageType: cmd.StorageType,
		Location:    cmd.Location,
		Checksum:    cmd.Checksum,
	}
	return s.repo.Update(ctx, s.client, params)
}

func (s *service) Delete(ctx context.Context, inodeID int64) error {
	return s.repo.Delete(ctx, s.client, inodeID)
}
