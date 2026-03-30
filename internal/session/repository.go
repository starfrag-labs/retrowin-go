package session

import "context"

// SessionRepository defines the interface for session data access.
type SessionRepository interface {
	// Save saves a session.
	Save(ctx context.Context, session *Session) error

	// Get retrieves a session by ID.
	Get(ctx context.Context, id SessionID) (*Session, error)

	// Delete deletes a session by ID.
	Delete(ctx context.Context, id SessionID) error

	// DeleteByUserID deletes all sessions for a user.
	DeleteByUserID(ctx context.Context, userID string) error
}
