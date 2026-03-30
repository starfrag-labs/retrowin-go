package user

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/utils"
)

// SystemUser represents a user's membership in a system.
type SystemUser struct {
	id       int
	userID   string
	systemID string
	username string
	uid      int
}

// NewSystemUser creates a new SystemUser.
func NewSystemUser(
	id int,
	userID string,
	systemID string,
	username string,
	uid int,
) *SystemUser {
	return &SystemUser{
		id:       id,
		userID:   userID,
		systemID: systemID,
		username: username,
		uid:      uid,
	}
}

// Getters
func (su *SystemUser) ID() int          { return su.id }
func (su *SystemUser) UserID() string   { return su.userID }
func (su *SystemUser) SystemID() string { return su.systemID }
func (su *SystemUser) Username() string { return su.username }
func (su *SystemUser) UID() int         { return su.uid }

// UserService defines the interface for user identity resolution.
type UserService interface {
	// ResolveUID extracts userID from context and looks up the UNIX uid
	// for the user in the specified system.
	ResolveUID(ctx context.Context, systemID string) (int, error)

	// ResolveUIDAndGIDs resolves both UID and all GIDs for permission checking.
	// GIDs are resolved from /etc/group in the filesystem.
	ResolveUIDAndGIDs(ctx context.Context, systemID string) (uid int, gids []int, err error)

	// GetByUserAndSystem retrieves the SystemUser for detailed info.
	GetByUserAndSystem(ctx context.Context, userID, systemID string) (*SystemUser, error)

	// Create creates a new system user.
	Create(ctx context.Context, cmd *CreateCommand) (*SystemUser, error)

	// GetByID retrieves a system user by ID.
	GetByID(ctx context.Context, id int) (*SystemUser, error)

	// Delete removes a system user.
	Delete(ctx context.Context, id int) error

	// Find retrieves system users matching the filter.
	Find(ctx context.Context, filter Filter) ([]*SystemUser, error)

	// FindOne retrieves a single system user matching the filter.
	FindOne(ctx context.Context, filter Filter) (*SystemUser, error)
}

// CreateCommand for creating a system-user (service layer).
type CreateCommand struct {
	UserID   string
	SystemID string
	Username string
	UID      int
}

// Filter for querying system-users (service layer).
type Filter = QueryFilter

// Filter helpers
func ByUserID(userID string) Filter {
	return Filter{UserID: &userID}
}

func BySystemID(systemID string) Filter {
	return Filter{SystemID: &systemID}
}

func ByUsername(username string) Filter {
	return Filter{Username: &username}
}

func ByUserAndSystem(userID, systemID string) Filter {
	return Filter{UserID: &userID, SystemID: &systemID}
}

type service struct {
	repo   SystemUserRepository
	client *ent.Client
}

// NewService creates a new UserService.
func NewService(repo SystemUserRepository, client *ent.Client) UserService {
	return &service{repo: repo, client: client}
}

// ResolveUID extracts userID from context and looks up UID for the system.
func (s *service) ResolveUID(ctx context.Context, systemID string) (int, error) {
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == "" {
		return 0, nil // No user in context, return 0 (skip permission check)
	}

	su, err := s.repo.FindOne(ctx, s.client, &QueryFilter{
		UserID:   &userID,
		SystemID: &systemID,
	})
	if err != nil {
		return 0, errors.WrapInternal(err, "failed to resolve uid")
	}
	if su == nil {
		return 0, errors.NotFound("user not found in system")
	}
	return su.UID(), nil
}

// ResolveUIDAndGIDs resolves both UID and GIDs for permission checking.
// GIDs will be resolved by fs service using /etc/group.
func (s *service) ResolveUIDAndGIDs(ctx context.Context, systemID string) (int, []int, error) {
	uid, err := s.ResolveUID(ctx, systemID)
	if err != nil {
		return 0, nil, err
	}
	// GIDs are resolved separately by fs service from /etc/group
	return uid, nil, nil
}

func (s *service) GetByUserAndSystem(ctx context.Context, userID, systemID string) (*SystemUser, error) {
	return s.repo.FindOne(ctx, s.client, &QueryFilter{
		UserID:   &userID,
		SystemID: &systemID,
	})
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*SystemUser, error) {
	if cmd.UserID == "" {
		return nil, errors.BadRequest("user_id is required")
	}
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}
	if cmd.Username == "" {
		return nil, errors.BadRequest("username is required")
	}

	params := &CreateParams{
		UserID:   cmd.UserID,
		SystemID: cmd.SystemID,
		Username: cmd.Username,
		UID:      cmd.UID,
	}
	return s.repo.Create(ctx, s.client, params)
}

func (s *service) GetByID(ctx context.Context, id int) (*SystemUser, error) {
	su, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if su == nil {
		return nil, errors.NotFound("system user not found")
	}
	return su, nil
}

func (s *service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, s.client, id)
}

func (s *service) Find(ctx context.Context, filter Filter) ([]*SystemUser, error) {
	return s.repo.Find(ctx, s.client, &filter)
}

func (s *service) FindOne(ctx context.Context, filter Filter) (*SystemUser, error) {
	su, err := s.repo.FindOne(ctx, s.client, &filter)
	if err != nil {
		return nil, err
	}
	if su == nil {
		return nil, errors.NotFound("system user not found")
	}
	return su, nil
}
