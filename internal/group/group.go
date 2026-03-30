package group

import (
	"time"
)

// Group represents a group within a system.
type Group struct {
	id        int64
	systemID  int64
	gid       string
	groupname string
	createdAt time.Time
	updatedAt time.Time
}

// NewGroup creates a new Group.
func NewGroup(
	id int64,
	systemID int64,
	gid string,
	groupname string,
	createdAt time.Time,
	updatedAt time.Time,
) *Group {
	return &Group{
		id:        id,
		systemID:  systemID,
		gid:       gid,
		groupname: groupname,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Getters
func (g *Group) ID() int64            { return g.id }
func (g *Group) SystemID() int64      { return g.systemID }
func (g *Group) GID() string          { return g.gid }
func (g *Group) Groupname() string    { return g.groupname }
func (g *Group) CreatedAt() time.Time { return g.createdAt }
func (g *Group) UpdatedAt() time.Time { return g.updatedAt }
