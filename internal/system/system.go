package system

import (
	"time"
)

// Status represents system status.
type Status string

const (
	StatusActive      Status = "active"
	StatusInactive    Status = "inactive"
	StatusMaintenance Status = "maintenance"
)

// System represents a system/node in the cluster.
type System struct {
	id          string
	name        string
	description *string
	status      Status
	createdAt   time.Time
	updatedAt   time.Time
}

// NewSystem creates a new System.
func NewSystem(
	id string,
	name string,
	description *string,
	status Status,
	createdAt time.Time,
	updatedAt time.Time,
) *System {
	return &System{
		id:          id,
		name:        name,
		description: description,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Getters
func (s *System) ID() string            { return s.id }
func (s *System) Name() string          { return s.name }
func (s *System) Description() *string  { return s.description }
func (s *System) Status() Status        { return s.status }
func (s *System) CreatedAt() time.Time  { return s.createdAt }
func (s *System) UpdatedAt() time.Time  { return s.updatedAt }