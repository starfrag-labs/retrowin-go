package auth

import (
	"time"
)

// SessionID is the type for session identifiers.
type SessionID string

// Session represents a user session.
type Session struct {
	id        SessionID
	userID    int64
	expiresAt time.Time
	createdAt time.Time
}

// NewSession creates a new session.
func NewSession(id SessionID, userID int64, expiresAt, createdAt time.Time) *Session {
	return &Session{
		id:        id,
		userID:    userID,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

// ID returns the session ID.
func (s *Session) ID() SessionID {
	return s.id
}

// UserID returns the user ID.
func (s *Session) UserID() int64 {
	return s.userID
}

// ExpiresAt returns the expiration time.
func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

// CreatedAt returns the creation time.
func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

// IsExpired checks if the session is expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}
