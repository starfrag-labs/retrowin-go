package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

type sessionService struct {
	repo SessionRepository
	ttl  time.Duration
}

// NewSessionService creates a new session service.
func NewSessionService(repo SessionRepository, ttl time.Duration) SessionService {
	return &sessionService{
		repo: repo,
		ttl:  ttl,
	}
}

// Create creates a new session for a user.
func (s *sessionService) Create(ctx context.Context, userID string) (*Session, error) {
	now := time.Now()
	id := SessionID(generateSessionID())
	session := NewSession(id, userID, now.Add(s.ttl), now)

	if err := s.repo.Save(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// Get retrieves a session by ID.
func (s *sessionService) Get(ctx context.Context, id SessionID) (*Session, error) {
	return s.repo.Get(ctx, id)
}

// Validate validates if session is still valid.
func (s *sessionService) Validate(ctx context.Context, id SessionID) (*Session, error) {
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
func (s *sessionService) Delete(ctx context.Context, id SessionID) error {
	return s.repo.Delete(ctx, id)
}

// DeleteByUserID deletes all sessions for a user.
func (s *sessionService) DeleteByUserID(ctx context.Context, userID string) error {
	return s.repo.DeleteByUserID(ctx, userID)
}

// generateSessionID generates a random session ID.
func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fall back to timestamp-based ID if crypto/rand fails
		return hex.EncodeToString([]byte(time.Now().String()))[:32]
	}
	return hex.EncodeToString(b)
}
