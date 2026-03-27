package token

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

// Errors
var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
)

// Token represents a stored token with metadata.
type Token struct {
	Key       string
	Value     string
	ExpiresAt time.Time
}

// Repository defines the interface for token storage.
type Repository interface {
	// Set stores a token with expiration.
	Set(ctx context.Context, key string, value string, expiration time.Duration) error

	// Get retrieves a token value by key.
	Get(ctx context.Context, key string) (string, error)

	// Delete deletes a token by key.
	Delete(ctx context.Context, key string) error
}

// Service defines the interface for token operations.
type Service interface {
	// Generate creates and stores a new token.
	Generate(ctx context.Context, prefix string, length int, expiration time.Duration) (*Token, error)

	// Validate checks if a token is valid and returns its value.
	Validate(ctx context.Context, key string) (string, error)

	// Consume validates and deletes a token (one-time use).
	Consume(ctx context.Context, key string) (string, error)

	// Delete removes a token.
	Delete(ctx context.Context, key string) error
}

type service struct {
	repo Repository
}

// NewService creates a new token service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Generate creates a new random token and stores it.
func (s *service) Generate(ctx context.Context, prefix string, length int, expiration time.Duration) (*Token, error) {
	token, err := generateRandomToken(length)
	if err != nil {
		return nil, err
	}

	key := prefix + ":" + token
	expiresAt := time.Now().Add(expiration)

	if err := s.repo.Set(ctx, key, token, expiration); err != nil {
		return nil, err
	}

	return &Token{
		Key:       key,
		Value:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// Validate checks if a token is valid.
func (s *service) Validate(ctx context.Context, key string) (string, error) {
	value, err := s.repo.Get(ctx, key)
	if err != nil {
		return "", ErrTokenNotFound
	}
	return value, nil
}

// Consume validates and deletes a token.
func (s *service) Consume(ctx context.Context, key string) (string, error) {
	value, err := s.repo.Get(ctx, key)
	if err != nil {
		return "", ErrTokenNotFound
	}

	if err := s.repo.Delete(ctx, key); err != nil {
		return "", err
	}

	return value, nil
}

// Delete removes a token.
func (s *service) Delete(ctx context.Context, key string) error {
	return s.repo.Delete(ctx, key)
}

// generateRandomToken generates a random hex token.
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
