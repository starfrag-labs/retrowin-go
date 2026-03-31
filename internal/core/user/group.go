package user

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// SystemGroup represents a group in a system.
type SystemGroup struct {
	id       int
	systemID string
	name     string
	gid      int
}

// NewSystemGroup creates a new SystemGroup.
func NewSystemGroup(
	id int,
	systemID string,
	name string,
	gid int,
) *SystemGroup {
	return &SystemGroup{
		id:       id,
		systemID: systemID,
		name:     name,
		gid:      gid,
	}
}

// Getters
func (g *SystemGroup) ID() int          { return g.id }
func (g *SystemGroup) SystemID() string { return g.systemID }
func (g *SystemGroup) Name() string     { return g.name }
func (g *SystemGroup) GID() int         { return g.gid }

// GroupService defines the interface for group operations.
type GroupService interface {
	// Create creates a new system group.
	Create(ctx context.Context, cmd *GroupCreateCommand) (*SystemGroup, error)

	// GetByID retrieves a system group by ID.
	GetByID(ctx context.Context, id int) (*SystemGroup, error)

	// GetByGID retrieves a system group by GID within a system.
	GetByGID(ctx context.Context, systemID string, gid int) (*SystemGroup, error)

	// Delete removes a system group.
	Delete(ctx context.Context, id int) error

	// Find retrieves system groups matching the filter.
	Find(ctx context.Context, filter GroupFilter) ([]*SystemGroup, error)

	// FindOne retrieves a single system group matching the filter.
	FindOne(ctx context.Context, filter GroupFilter) (*SystemGroup, error)

	// AddUserToGroup adds a user to a group.
	AddUserToGroup(ctx context.Context, userSystemID, groupID int) error

	// RemoveUserFromGroup removes a user from a group.
	RemoveUserFromGroup(ctx context.Context, userSystemID, groupID int) error
}

// GroupCreateCommand for creating a system group (service layer).
type GroupCreateCommand struct {
	SystemID string
	Name     string
	GID      int
}

// GroupFilter for querying system groups (service layer).
type GroupFilter = GroupQueryFilter

// Group filter helpers
func GroupBySystemID(systemID string) GroupFilter {
	return GroupFilter{SystemID: &systemID}
}

func GroupByName(name string) GroupFilter {
	return GroupFilter{Name: &name}
}

func GroupByGID(gid int) GroupFilter {
	return GroupFilter{GID: &gid}
}

func GroupBySystemAndName(systemID, name string) GroupFilter {
	return GroupFilter{SystemID: &systemID, Name: &name}
}

func GroupBySystemAndGID(systemID string, gid int) GroupFilter {
	return GroupFilter{SystemID: &systemID, GID: &gid}
}

type groupService struct {
	repo   SystemGroupRepository
	client *ent.Client
}

// NewGroupService creates a new GroupService.
func NewGroupService(repo SystemGroupRepository, client *ent.Client) GroupService {
	return &groupService{repo: repo, client: client}
}

func (s *groupService) Create(ctx context.Context, cmd *GroupCreateCommand) (*SystemGroup, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}
	if cmd.Name == "" {
		return nil, errors.BadRequest("name is required")
	}

	params := &GroupCreateParams{
		SystemID: cmd.SystemID,
		Name:     cmd.Name,
		GID:      cmd.GID,
	}
	return s.repo.Create(ctx, s.client, params)
}

func (s *groupService) GetByID(ctx context.Context, id int) (*SystemGroup, error) {
	g, err := s.repo.GetByID(ctx, s.client, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.NotFound("system group not found")
	}
	return g, nil
}

func (s *groupService) GetByGID(ctx context.Context, systemID string, gid int) (*SystemGroup, error) {
	g, err := s.repo.FindOne(ctx, s.client, &GroupQueryFilter{
		SystemID: &systemID,
		GID:      &gid,
	})
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.NotFound("system group not found")
	}
	return g, nil
}

func (s *groupService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, s.client, id)
}

func (s *groupService) Find(ctx context.Context, filter GroupFilter) ([]*SystemGroup, error) {
	return s.repo.Find(ctx, s.client, &filter)
}

func (s *groupService) FindOne(ctx context.Context, filter GroupFilter) (*SystemGroup, error) {
	g, err := s.repo.FindOne(ctx, s.client, &filter)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.NotFound("system group not found")
	}
	return g, nil
}

func (s *groupService) AddUserToGroup(ctx context.Context, userSystemID, groupID int) error {
	return s.repo.AddUserToGroup(ctx, s.client, userSystemID, groupID)
}

func (s *groupService) RemoveUserFromGroup(ctx context.Context, userSystemID, groupID int) error {
	return s.repo.RemoveUserFromGroup(ctx, s.client, userSystemID, groupID)
}
