package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRepository implements Repository using Redis.
type RedisRepository struct {
	client *redis.Client
	prefix string
}

// NewRedisRepository creates a new Redis session repository.
func NewRedisRepository(client *redis.Client, prefix string) Repository {
	return &RedisRepository{
		client: client,
		prefix: prefix,
	}
}

// Save saves a session.
func (r *RedisRepository) Save(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := r.getKey(session.ID())
	ttl := time.Until(session.ExpiresAt())

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a session by ID.
func (r *RedisRepository) Get(ctx context.Context, id ID) (*Session, error) {
	key := r.getKey(id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// Delete deletes a session by ID.
func (r *RedisRepository) Delete(ctx context.Context, id ID) error {
	key := r.getKey(id)
	return r.client.Del(ctx, key).Err()
}

// DeleteByUserID deletes all sessions for a user.
func (r *RedisRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	pattern := fmt.Sprintf("%s:user:%d:*", r.prefix, userID)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *RedisRepository) getKey(id ID) string {
	return fmt.Sprintf("%s:session:%s", r.prefix, id)
}
