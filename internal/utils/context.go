package utils

import (
	"context"
)

// Context keys for storing values in context.
type ctxKey string

const (
	UserIDKey    ctxKey = "user_id"
	SessionIDKey ctxKey = "session_id"
)

// ContextWithUserID returns a context with user ID.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// ContextWithSession returns a context with session information.
func ContextWithSession(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// GetUserID retrieves the user ID from context.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetSessionID retrieves the session ID from context.
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(SessionIDKey).(string)
	return sessionID, ok
}
