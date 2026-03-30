package user

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Service defines the interface for user operations.
type Service interface {
	Get(ctx context.Context, provider, providerID string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, cmd *CreateCommand) (*User, error)
	Delete(ctx context.Context, provider, providerID string) error
	FindOrCreateByOIDC(ctx context.Context, provider, subject, username string) (string, error)
}

// CreateCommand for creating a user (service layer).
type CreateCommand struct {
	Username   string
	Provider   string
	ProviderID string
}

type service struct {
	repo   Repository
	client *ent.Client
}

// NewService creates a new user service.
func NewService(repo Repository, client *ent.Client) Service {
	return &service{
		repo:   repo,
		client: client,
	}
}

func (s *service) Get(ctx context.Context, provider, providerID string) (*User, error) {
	user, err := s.repo.GetByProvider(ctx, s.client, provider, providerID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}
	return user, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}
	return user, nil
}

func (s *service) GetByUsername(ctx context.Context, username string) (*User, error) {
	user, err := s.repo.GetByUsername(ctx, s.client, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}
	return user, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*User, error) {
	if cmd.Provider == "" {
		return nil, errors.BadRequest("provider is required")
	}
	if cmd.ProviderID == "" {
		return nil, errors.BadRequest("providerId is required")
	}
	if !IsValidProvider(cmd.Provider) {
		return nil, errors.BadRequest("invalid provider")
	}

	exists, err := s.repo.ExistsByProvider(ctx, s.client, cmd.Provider, cmd.ProviderID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("user already exists")
	}

	params := &CreateParams{
		Username:   cmd.Username,
		Provider:   cmd.Provider,
		ProviderID: cmd.ProviderID,
	}
	return s.repo.Create(ctx, s.client, params)
}

func (s *service) Delete(ctx context.Context, provider, providerID string) error {
	user, err := s.repo.GetByProvider(ctx, s.client, provider, providerID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NotFound("user not found")
	}
	return s.repo.Delete(ctx, s.client, user.ID())
}

func (s *service) FindOrCreateByOIDC(ctx context.Context, provider, subject, username string) (string, error) {
	user, err := s.repo.GetByProvider(ctx, s.client, provider, subject)
	if err != nil {
		return "", err
	}

	if user != nil {
		return user.ID(), nil
	}

	createCmd := &CreateCommand{
		Username:   username,
		Provider:   provider,
		ProviderID: subject,
	}

	newUser, err := s.Create(ctx, createCmd)
	if err != nil {
		return "", err
	}

	return newUser.ID(), nil
}
