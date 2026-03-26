package session

import (
	"time"
)

// ID is a unique identifier for a session.
type ID string

func (id ID) String() string {
	return string(id)
}

// Session represents an authenticated user session.
type Session struct {
	id        ID
	userID    int64
	expiresAt time.Time
	createdAt time.Time
}

// NewSession creates a new session.
func NewSession(id ID, userID int64, expiresAt, createdAt time.Time) *Session {
	return &Session{
		id:        id,
		userID:    userID,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

func (s *Session) ID() ID            { return s.id }
func (s *Session) UserID() int64     { return s.userID }
func (s *Session) ExpiresAt() time.Time { return s.expiresAt }
func (s *Session) CreatedAt() time.Time { return s.createdAt }

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}
