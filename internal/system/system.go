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
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
