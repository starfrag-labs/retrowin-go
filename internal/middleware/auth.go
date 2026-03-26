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
)

// GetUserID extracts user ID from context.
func GetUserID(ctx context.Context) int64 {
	if id, ok := ctx.Value(UserIDKey).(int64); ok {
		return id
	}
	return 0
}

// SetUserID adds user ID to context.
func SetUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, err *errors.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	_, _ = w.Write([]byte(`{"code":"` + err.Code + `","message":"` + err.Message + `"}`))
}
