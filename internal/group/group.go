package group

import (
	"time"
)

// Group represents a group within a system.
type Group struct {
	ID          int64     `json:"id"`
	SystemID    int64     `json:"systemId"`
	GID         string    `json:"gid"`
	Groupname   string    `json:"groupname"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
