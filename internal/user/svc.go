package user

import (
	"context"
	"errors"
)

// Service defines the interface for user operations.
type Service interface {
	// Get retrieves a user by provider and provider ID.
	Get(ctx context.Context, provider, providerID string) (*User, error)

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id int64) (*User, error)

	// Create creates a new user.
	Create(ctx context.Context, cmd *CreateCommand) (*User, error)

	// Delete deletes a user.
	Delete(ctx context.Context, provider, providerID string) error

	// FindOrCreateByOIDC finds an existing user by OIDC subject or creates a new one.
	FindOrCreateByOIDC(ctx context.Context, provider, subject, email, name, picture string) (int64, error)
}

// Errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidProvider   = errors.New("invalid provider")
)

type service struct {
	userRepo Repository
}

// NewService creates a new user service.
func NewService(userRepo Repository) Service {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) Get(ctx context.Context, provider, providerID string) (*User, error) {
	user, err := s.userRepo.GetByProvider(ctx, provider, providerID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*User, error) {
	// Validate input
	if cmd.Provider == "" {
		return nil, errors.New("provider is required")
	}
	if cmd.ProviderID == "" {
		return nil, errors.New("providerId is required")
	}
	if !IsValidProvider(cmd.Provider) {
		return nil, ErrInvalidProvider
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByProvider(ctx, cmd.Provider, cmd.ProviderID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// Create user
	user, err := s.userRepo.Create(ctx, cmd.Provider, cmd.ProviderID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Delete(ctx context.Context, provider, providerID string) error {
	user, err := s.userRepo.GetByProvider(ctx, provider, providerID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *service) FindOrCreateByOIDC(ctx context.Context, provider, subject, email, name, picture string) (int64, error) {
	// Try to find existing user
	user, err := s.userRepo.GetByProvider(ctx, provider, subject)
	if err != nil {
		return 0, err
	}

	if user != nil {
		return user.ID, nil
	}

	// Create new user
	createCmd := &CreateCommand{
		Provider:   provider,
		ProviderID: subject,
	}

	newUser, err := s.Create(ctx, createCmd)
	if err != nil {
		return 0, err
	}

	return newUser.ID, nil
}
