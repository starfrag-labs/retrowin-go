package object

import (
	"context"
	"io"
	"time"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Provider represents the storage provider type.
type Provider string

const (
	ProviderS3 Provider = "s3"
)

// Object represents a tracked object in external storage.
type Object struct {
	id         string
	provider   Provider
	bucket     string
	systemID   string
	storageKey string
	createdAt  time.Time
	updatedAt  time.Time
}

// NewObject creates a new Object.
func NewObject(
	id string,
	provider Provider,
	bucket string,
	systemID string,
	storageKey string,
	createdAt time.Time,
	updatedAt time.Time,
) *Object {
	return &Object{
		id:         id,
		provider:   provider,
		bucket:     bucket,
		systemID:   systemID,
		storageKey: storageKey,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// Getters
func (o *Object) ID() string           { return o.id }
func (o *Object) Provider() Provider   { return o.provider }
func (o *Object) Bucket() string       { return o.bucket }
func (o *Object) SystemID() string     { return o.systemID }
func (o *Object) StorageKey() string   { return o.storageKey }
func (o *Object) CreatedAt() time.Time { return o.createdAt }
func (o *Object) UpdatedAt() time.Time { return o.updatedAt }

// ObjectService defines the interface for object operations.
type ObjectService interface {
	// Create streams data to storage and creates an Object DB record atomically.
	Create(ctx context.Context, cmd *CreateCommand) (*Object, error)
	GetByID(ctx context.Context, id string) (*Object, error)
	GetByStorageKey(ctx context.Context, systemID string, provider Provider, bucket string, storageKey string) (*Object, error)
	// Delete atomically removes from external storage and DB.
	Delete(ctx context.Context, id string) error
	Find(ctx context.Context, filter Filter) ([]*Object, error)
	FindOne(ctx context.Context, filter Filter) (*Object, error)
	GetDownloadURL(ctx context.Context, id string) (string, error)
}

// CreateCommand for creating a new object (service layer).
type CreateCommand struct {
	Provider   Provider
	Bucket     string
	SystemID   string
	StorageKey string
	Reader     io.Reader
	Size       int64
}

// Filter for querying objects (service layer).
type Filter = QueryFilter

// Filter helpers
func ByID(id string) Filter {
	return Filter{ID: &id}
}

func BySystemID(systemID string) Filter {
	return Filter{SystemID: &systemID}
}

type service struct {
	repo    ObjectRepository
	storage Storage
	client  *ent.Client
}

// NewService creates a new ObjectService.
func NewService(repo ObjectRepository, storage Storage, client *ent.Client) ObjectService {
	return &service{repo: repo, storage: storage, client: client}
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*Object, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}
	if cmd.StorageKey == "" {
		return nil, errors.BadRequest("storage_key is required")
	}

	provider := cmd.Provider
	if provider == "" {
		provider = ProviderS3
	}

	// Stream upload to storage
	if err := s.storage.PutObject(ctx, cmd.Bucket, cmd.StorageKey, cmd.Reader, cmd.Size); err != nil {
		return nil, errors.WrapInternal(err, "failed to upload to storage")
	}

	// Create object record in DB
	params := &CreateParams{
		Provider:   provider,
		Bucket:     cmd.Bucket,
		SystemID:   cmd.SystemID,
		StorageKey: cmd.StorageKey,
	}
	obj, err := s.repo.Create(ctx, s.client, params)
	if err != nil {
		// Attempt cleanup on DB failure
		_ = s.storage.DeleteObject(ctx, cmd.Bucket, cmd.StorageKey)
		return nil, errors.WrapInternal(err, "failed to create object record")
	}

	return obj, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Object, error) {
	obj, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.NotFound("object not found")
	}
	return obj, nil
}

func (s *service) GetByStorageKey(ctx context.Context, systemID string, provider Provider, bucket string, storageKey string) (*Object, error) {
	obj, err := s.repo.GetByStorageKey(ctx, s.client, systemID, string(provider), bucket, storageKey)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.NotFound("object not found")
	}
	return obj, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	obj, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return err
	}
	if obj == nil {
		return errors.NotFound("object not found")
	}

	if err := s.storage.DeleteObject(ctx, obj.Bucket(), obj.StorageKey()); err != nil {
		return errors.WrapInternal(err, "failed to delete from storage")
	}

	return s.repo.Delete(ctx, s.client, id)
}

func (s *service) Find(ctx context.Context, filter Filter) ([]*Object, error) {
	return s.repo.Find(ctx, s.client, &filter)
}

func (s *service) FindOne(ctx context.Context, filter Filter) (*Object, error) {
	obj, err := s.repo.FindOne(ctx, s.client, &filter)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.NotFound("object not found")
	}
	return obj, nil
}

func (s *service) GetDownloadURL(ctx context.Context, id string) (string, error) {
	obj, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return "", err
	}
	if obj == nil {
		return "", errors.NotFound("object not found")
	}

	url, err := s.storage.GetPresignedDownloadURL(ctx, obj.Bucket(), obj.StorageKey(), DefaultDownloadExpiry)
	if err != nil {
		return "", errors.WrapInternal(err, "failed to generate download URL")
	}
	return url, nil
}
