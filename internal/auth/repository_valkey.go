package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
)

// sessionData is the serializable representation of Session for Valkey storage.
type sessionData struct {
	UserID    int64     `json:"user_id"`
	UserUID   string   `json:"user_uid"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// toSessionData converts Session to sessionData.
func toSessionData(s *Session) *sessionData {
	return &sessionData{
		UserID:    s.UserID(),
		UserUID:   s.UserUID(),
		ExpiresAt: s.ExpiresAt(),
		CreatedAt: s.CreatedAt(),
	}
}

// ValkeySessionRepository implements SessionRepository using Valkey.
type ValkeySessionRepository struct {
	client valkey.Client
	prefix string
}

// NewValkeySessionRepository creates a new ValkeySessionRepository.
func NewValkeySessionRepository(client valkey.Client, prefix string) SessionRepository {
	return &ValkeySessionRepository{
		client: client,
		prefix: prefix,
	}
}

func (r *ValkeySessionRepository) sessionKey(id SessionID) string {
	return fmt.Sprintf("%s:session:%s", r.prefix, id)
}

func (r *ValkeySessionRepository) userSessionsKey(userID int64) string {
	return fmt.Sprintf("%s:user:sessions:%d", r.prefix, userID)
}

// Save saves a session.
func (r *ValkeySessionRepository) Save(ctx context.Context, session *Session) error {
	data, err := json.Marshal(toSessionData(session))
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	// Calculate TTL from session expiration
	ttl := time.Until(session.ExpiresAt())
	if ttl <= 0 {
		ttl = 24 * time.Hour // Default TTL
	}

	// Store session data with expiration
	err = r.client.Do(ctx, r.client.B().Set().
		Key(r.sessionKey(session.ID())).
		Value(string(data)).
		ExSeconds(int64(ttl.Seconds())).
		Build()).Error()
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	// Add to user's session set for DeleteByUserID support
	err = r.client.Do(ctx, r.client.B().Sadd().
		Key(r.userSessionsKey(session.UserID())).
		Member(string(session.ID())).
		Build()).Error()
	if err != nil {
		return fmt.Errorf("add to user sessions: %w", err)
	}

	return nil
}

// Get retrieves a session by ID.
func (r *ValkeySessionRepository) Get(ctx context.Context, id SessionID) (*Session, error) {
	result := r.client.Do(ctx, r.client.B().Get().Key(r.sessionKey(id)).Build())
	if result.Error() != nil {
		return nil, nil // Key not found
	}

	data, err := result.ToString()
	if err != nil {
		return nil, nil
	}

	var sd sessionData
	if err := json.Unmarshal([]byte(data), &sd); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return NewSession(id, sd.UserID, sd.UserUID, sd.ExpiresAt, sd.CreatedAt), nil
}

// Delete deletes a session by ID.
func (r *ValkeySessionRepository) Delete(ctx context.Context, id SessionID) error {
	// Get session first to remove from user's set
	s, _ := r.Get(ctx, id)

	if s != nil {
		_ = r.client.Do(ctx, r.client.B().Srem().
			Key(r.userSessionsKey(s.UserID())).
			Member(string(id)).
			Build()).Error()
	}

	delErr := r.client.Do(ctx, r.client.B().Del().Key(r.sessionKey(id)).Build()).Error()
	if delErr != nil {
		return fmt.Errorf("delete session: %w", delErr)
	}
	return nil
}

// DeleteByUserID deletes all sessions for a user.
func (r *ValkeySessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	// Get all session IDs for user
	result := r.client.Do(ctx, r.client.B().Smembers().Key(r.userSessionsKey(userID)).Build())
	if result.Error() != nil {
		return fmt.Errorf("get user sessions: %w", result.Error())
	}

	members, err := result.AsStrSlice()
	if err != nil {
		return fmt.Errorf("parse user sessions: %w", err)
	}

	// Delete all sessions
	if len(members) > 0 {
		keys := make([]string, 0, len(members))
		for _, sid := range members {
			keys = append(keys, r.sessionKey(SessionID(sid)))
		}

		if err := r.client.Do(ctx, r.client.B().Del().Key(keys...).Build()).Error(); err != nil {
			return fmt.Errorf("delete sessions: %w", err)
		}
	}

	// Delete user's session set
	if err := r.client.Do(ctx, r.client.B().Del().Key(r.userSessionsKey(userID)).Build()).Error(); err != nil {
		return fmt.Errorf("delete user sessions set: %w", err)
	}
	return nil
}
