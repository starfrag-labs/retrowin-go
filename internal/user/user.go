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
	JoinDate   time.Time `json:"joinDate"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// CreateCommand represents the command to create a new user.
type CreateCommand struct {
	Provider   string `json:"provider"`
	ProviderID string `json:"providerId"`
}
