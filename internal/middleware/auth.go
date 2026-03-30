package middleware

import (
	"context"
	"net/http"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// ContextKey type for context keys.
type ContextKey string

const (
	// UserIDKey is the context key for user ID.
	UserIDKey ContextKey = "user_id"
	// SessionIDKey is the context key for session ID.
	SessionIDKey ContextKey = "session_id"
)

// GetUserID extracts user ID from context.
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSessionID extracts session ID from context.
func GetSessionID(ctx context.Context) string {
	if id, ok := ctx.Value(SessionIDKey).(string); ok {
		return id
	}
	return ""
}

// SetUserID adds user ID to context.
func SetUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

// SetSessionID adds session ID to context.
func SetSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, SessionIDKey, id)
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, err *errors.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	_, _ = w.Write([]byte(`{"code":"` + err.Code + `","message":"` + err.Message + `"}`))
}
