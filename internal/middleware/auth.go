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
	// UserUIDKey is the context key for user UID.
	UserUIDKey ContextKey = "user_uid"
	// SessionIDKey is the context key for session ID.
	SessionIDKey ContextKey = "session_id"
)

// GetUserID extracts user ID from context.
func GetUserID(ctx context.Context) int64 {
	if id, ok := ctx.Value(UserIDKey).(int64); ok {
		return id
	}
	return 0
}

// GetUserUID extracts user UID from context.
func GetUserUID(ctx context.Context) string {
	if uid, ok := ctx.Value(UserUIDKey).(string); ok {
		return uid
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
func SetUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

// SetUserUID adds user UID to context.
func SetUserUID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, UserUIDKey, uid)
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
