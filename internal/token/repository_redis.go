package token

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRepository implements Repository using Redis.
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis token repository.
func NewRedisRepository(client *redis.Client) Repository {
	return &RedisRepository{client: client}
}

// Set stores a token with expiration.
func (r *RedisRepository) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a token value by key.
func (r *RedisRepository) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", err
	}
	return result, nil
}

// Delete deletes a token by key.
func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
