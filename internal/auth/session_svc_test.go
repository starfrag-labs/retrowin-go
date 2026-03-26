package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/starfrag-lab/retrowin-go/internal/auth"
	authMocks "github.com/starfrag-lab/retrowin-go/internal/auth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSessionService_Create(t *testing.T) {
	ctx := context.Background()
	userID := int64(123)
	ttl := 24 * time.Hour

	t.Run("creates session successfully", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*auth.Session")).Return(nil)

		session, err := svc.Create(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, userID, session.UserID())
		assert.NotEmpty(t, session.ID())
		assert.True(t, session.ExpiresAt().After(time.Now()))
	})

	t.Run("returns error when save fails", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*auth.Session")).Return(assert.AnError)

		session, err := svc.Create(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, session)
	})
}

func TestSessionService_Get(t *testing.T) {
	ctx := context.Background()
	sessionID := auth.SessionID("test-session-id")
	ttl := 24 * time.Hour

	t.Run("returns session when found", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		expectedSession := auth.NewSession(sessionID, 123, time.Now().Add(ttl), time.Now())
		repo.EXPECT().Get(mock.Anything, sessionID).Return(expectedSession, nil)

		session, err := svc.Get(ctx, sessionID)

		assert.NoError(t, err)
		assert.Equal(t, expectedSession, session)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		repo.EXPECT().Get(mock.Anything, sessionID).Return(nil, nil)

		session, err := svc.Get(ctx, sessionID)

		assert.NoError(t, err)
		assert.Nil(t, session)
	})
}

func TestSessionService_Validate(t *testing.T) {
	ctx := context.Background()
	sessionID := auth.SessionID("test-session-id")
	ttl := 24 * time.Hour

	t.Run("returns session when valid", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		validSession := auth.NewSession(sessionID, 123, time.Now().Add(ttl), time.Now())
		repo.EXPECT().Get(mock.Anything, sessionID).Return(validSession, nil)

		session, err := svc.Validate(ctx, sessionID)

		assert.NoError(t, err)
		assert.Equal(t, validSession, session)
	})

	t.Run("returns error when session not found", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		repo.EXPECT().Get(mock.Anything, sessionID).Return(nil, nil)

		session, err := svc.Validate(ctx, sessionID)

		assert.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("returns error when session expired", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		expiredSession := auth.NewSession(sessionID, 123, time.Now().Add(-1*time.Hour), time.Now())
		repo.EXPECT().Get(mock.Anything, sessionID).Return(expiredSession, nil)

		session, err := svc.Validate(ctx, sessionID)

		assert.Error(t, err)
		assert.Nil(t, session)
	})
}

func TestSessionService_Delete(t *testing.T) {
	ctx := context.Background()
	sessionID := auth.SessionID("test-session-id")
	ttl := 24 * time.Hour

	t.Run("deletes session successfully", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		repo.EXPECT().Delete(mock.Anything, sessionID).Return(nil)

		err := svc.Delete(ctx, sessionID)

		assert.NoError(t, err)
	})

	t.Run("returns error when delete fails", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		repo.EXPECT().Delete(mock.Anything, sessionID).Return(assert.AnError)

		err := svc.Delete(ctx, sessionID)

		assert.Error(t, err)
	})
}

func TestSessionService_DeleteByUserID(t *testing.T) {
	ctx := context.Background()
	userID := int64(123)
	ttl := 24 * time.Hour

	t.Run("deletes all user sessions successfully", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		repo.EXPECT().DeleteByUserID(mock.Anything, userID).Return(nil)

		err := svc.DeleteByUserID(ctx, userID)

		assert.NoError(t, err)
	})

	t.Run("returns error when delete fails", func(t *testing.T) {
		repo := authMocks.NewSessionRepositoryMock(t)
		svc := auth.NewSessionService(repo, ttl)

		repo.EXPECT().DeleteByUserID(mock.Anything, userID).Return(assert.AnError)

		err := svc.DeleteByUserID(ctx, userID)

		assert.Error(t, err)
	})
}
