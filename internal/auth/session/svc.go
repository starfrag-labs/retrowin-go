package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Service defines the session service interface.
type Service interface {
	// Create creates a new session for a user.
	Create(ctx context.Context, userID int64) (*Session, error)

	// Get retrieves a session by ID.
	Get(ctx context.Context, id ID) (*Session, error)

	// Validate validates if session is still valid.
	Validate(ctx context.Context, id ID) (*Session, error)

	// Delete deletes a session (logout).
	Delete(ctx context.Context, id ID) error

	// DeleteByUserID deletes all sessions for a user.
	DeleteByUserID(ctx context.Context, userID int64) error
}

type service struct {
	repo Repository
	ttl  time.Duration
}

// NewService creates a new session service.
func NewService(repo Repository, ttl time.Duration) Service {
	return &service{
		repo: repo,
		ttl:  ttl,
	}
}

// Create creates a new session for a user.
func (s *service) Create(ctx context.Context, userID int64) (*Session, error) {
	now := time.Now()
	id := ID(generateSessionID())
	session := NewSession(id, userID, now.Add(s.ttl), now)

	if err := s.repo.Save(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// Get retrieves a session by ID.
func (s *service) Get(ctx context.Context, id ID) (*Session, error) {
	return s.repo.Get(ctx, id)
}

// Validate validates if session is still valid.
func (s *service) Validate(ctx context.Context, id ID) (*Session, error) {
	session, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.NotFound("session not found")
	}
	if session.IsExpired() {
		return nil, errors.Unauthorized("session expired")
	}
	return session, nil
}

// Delete deletes a session (logout).
func (s *service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}

// DeleteByUserID deletes all sessions for a user.
func (s *service) DeleteByUserID(ctx context.Context, userID int64) error {
	return s.repo.DeleteByUserID(ctx, userID)
}

// generateSessionID generates a random session ID.
func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
