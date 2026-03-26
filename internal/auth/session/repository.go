package session

import (
	"context"
)

// Repository defines the interface for session storage.
type Repository interface {
	// Save saves a session.
	Save(ctx context.Context, session *Session) error

	// Get retrieves a session by ID.
	Get(ctx context.Context, id ID) (*Session, error)

	// Delete deletes a session by ID.
	Delete(ctx context.Context, id ID) error

	// DeleteByUserID deletes all sessions for a user.
	DeleteByUserID(ctx context.Context, userID int64) error
}
