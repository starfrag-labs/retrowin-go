package user

import (
	"time"
)

// Provider types
const (
	ProviderKeycloak = "keycloak"
	ProviderGoogle   = "google"
)

// IsValidProvider checks if the provider is valid.
func IsValidProvider(provider string) bool {
	switch provider {
	case ProviderKeycloak, ProviderGoogle:
		return true
	default:
		return false
	}
}

// User represents a user in the system.
type User struct {
	ID         int64     `json:"id"`
	Provider   string    `json:"provider"`
	ProviderID string    `json:"providerId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// ServiceStatus represents the service status for a user.
type ServiceStatus struct {
	UserID     int64     `json:"userId"`
	Available  bool      `json:"available"`
	JoinDate   time.Time `json:"joinDate"`
	UpdateDate time.Time `json:"updateDate"`
}

// CreateCommand represents the command to create a new user.
type CreateCommand struct {
	Provider   string `json:"provider"`
	ProviderID string `json:"providerId"`
}
