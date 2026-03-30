package user

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Service defines the interface for user operations.
type Service interface {
	// Get retrieves a user by provider and provider ID.
	Get(ctx context.Context, provider, providerID string) (*User, error)

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByUID retrieves a user by UID.
	GetByUID(ctx context.Context, uid string) (*User, error)

	// Create creates a new user.
	Create(ctx context.Context, cmd *CreateCommand) (*User, error)

	// Delete deletes a user.
	Delete(ctx context.Context, provider, providerID string) error

	// FindOrCreateByOIDC finds an existing user by OIDC subject or creates a new one.
	// Returns userID (int64) and userUID (string).
	FindOrCreateByOIDC(ctx context.Context, provider, subject, email, name, picture string) (int64, string, error)
}

// CreateCommand for creating a user (service layer).
type CreateCommand struct {
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

func (s *service) GetByID(ctx context.Context, id int64) (*User, error) {
	user, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}
	return user, nil
}

func (s *service) GetByUID(ctx context.Context, uid string) (*User, error) {
	user, err := s.repo.GetByUID(ctx, s.client, uid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}
	return user, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*User, error) {
	// Validate input
	if cmd.Provider == "" {
		return nil, errors.BadRequest("provider is required")
	}
	if cmd.ProviderID == "" {
		return nil, errors.BadRequest("providerId is required")
	}
	if !IsValidProvider(cmd.Provider) {
		return nil, errors.BadRequest("invalid provider")
	}

	// Check if user already exists
	exists, err := s.repo.ExistsByProvider(ctx, s.client, cmd.Provider, cmd.ProviderID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("user already exists")
	}

	// Create user
	params := &CreateParams{
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

	// Delete user
	return s.repo.Delete(ctx, s.client, user.ID())
}

func (s *service) FindOrCreateByOIDC(ctx context.Context, provider, subject, email, name, picture string) (int64, string, error) {
	// Try to find existing user
	user, err := s.repo.GetByProvider(ctx, s.client, provider, subject)
	if err != nil {
		return 0, "", err
	}

	if user != nil {
		return user.ID(), user.UID(), nil
	}

	// Create new user
	createCmd := &CreateCommand{
		Provider:   provider,
		ProviderID: subject,
	}

	newUser, err := s.Create(ctx, createCmd)
	if err != nil {
		return 0, "", err
	}

	return newUser.ID(), newUser.UID(), nil
}
