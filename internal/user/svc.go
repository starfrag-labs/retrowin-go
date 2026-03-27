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

	// Delete deletes a user and all associated data.
	Delete(ctx context.Context, provider, providerID string) error

	// GetServiceStatus retrieves the service status for a user.
	GetServiceStatus(ctx context.Context, provider, providerID string) (*ServiceStatus, error)

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
	userRepo   Repository
	statusRepo ServiceStatusRepository
}

// NewService creates a new user service.
func NewService(userRepo Repository, statusRepo ServiceStatusRepository) Service {
	return &service{
		userRepo:   userRepo,
		statusRepo: statusRepo,
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

	// Create service status
	_, err = s.statusRepo.Create(ctx, user.ID)
	if err != nil {
		// Rollback user creation on failure
		_ = s.userRepo.Delete(ctx, user.ID)
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

	// Delete service status first
	if err := s.statusRepo.Delete(ctx, user.ID); err != nil {
		return err
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *service) GetServiceStatus(ctx context.Context, provider, providerID string) (*ServiceStatus, error) {
	user, err := s.userRepo.GetByProvider(ctx, provider, providerID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	status, err := s.statusRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, errors.New("service status not found")
	}

	return status, nil
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
