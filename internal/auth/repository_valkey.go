package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// ValkeySessionRepository implements SessionRepository using Valkey/Redis.
type ValkeySessionRepository struct {
	client *redis.Client
	prefix string
}

// NewValkeySessionRepository creates a new ValkeySessionRepository.
func NewValkeySessionRepository(client *redis.Client, prefix string) SessionRepository {
	return &ValkeySessionRepository{
		client: client,
		prefix: prefix,
	}
}

// Save saves a session.
func (r *ValkeySessionRepository) Save(ctx context.Context, session *Session) error {
	key := r.getKey(session.ID())

	data := &sessionData{
		UserID:    session.UserID(),
		ExpiresAt: session.ExpiresAt(),
		CreatedAt: session.CreatedAt(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ttl := time.Until(session.ExpiresAt())
	if ttl < 0 {
		ttl = time.Minute // Minimum TTL
	}

	return r.client.Set(ctx, key, jsonData, ttl).Err()
}

// Get retrieves a session by ID.
func (r *ValkeySessionRepository) Get(ctx context.Context, id SessionID) (*Session, error) {
	key := r.getKey(id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, errors.NotFound("session not found")
	}
	if err != nil {
		return nil, err
	}

	var sessionData sessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, err
	}

	return NewSession(id, sessionData.UserID, sessionData.ExpiresAt, sessionData.CreatedAt), nil
}

// Delete deletes a session by ID.
func (r *ValkeySessionRepository) Delete(ctx context.Context, id SessionID) error {
	key := r.getKey(id)
	return r.client.Del(ctx, key).Err()
}

// DeleteByUserID deletes all sessions for a user.
// Note: This implementation scans for all sessions with the user ID prefix.
// In production, consider using a separate index or set to track user sessions.
func (r *ValkeySessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	pattern := r.prefix + ":session:*"
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var sessionData sessionData
		if err := json.Unmarshal(data, &sessionData); err != nil {
			continue
		}

		if sessionData.UserID == userID {
			r.client.Del(ctx, key)
		}
	}
	return iter.Err()
}

func (r *ValkeySessionRepository) getKey(id SessionID) string {
	return r.prefix + ":session:" + string(id)
}

type sessionData struct {
	UserID    int64     `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}
