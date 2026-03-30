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
	id         int64
	uid        string
	provider   string
	providerID string
	joinDate   time.Time
	createdAt  time.Time
	updatedAt  time.Time
}

// NewUser creates a new User.
func NewUser(
	id int64,
	uid string,
	provider string,
	providerID string,
	joinDate time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:         id,
		uid:        uid,
		provider:   provider,
		providerID: providerID,
		joinDate:   joinDate,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// Getters
func (u *User) ID() int64            { return u.id }
func (u *User) UID() string          { return u.uid }
func (u *User) Provider() string     { return u.provider }
func (u *User) ProviderID() string   { return u.providerID }
func (u *User) JoinDate() time.Time  { return u.joinDate }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
