package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	id := SessionID("test-session-id")
	userID := int64(123)
	expiresAt := time.Now().Add(24 * time.Hour)
	createdAt := time.Now()

	session := NewSession(id, userID, expiresAt, createdAt)

	assert.Equal(t, id, session.ID())
	assert.Equal(t, userID, session.UserID())
	assert.Equal(t, expiresAt.Unix(), session.ExpiresAt().Unix())
	assert.Equal(t, createdAt.Unix(), session.CreatedAt().Unix())
}

func TestSession_IsExpired_NotExpired(t *testing.T) {
	session := NewSession(
		SessionID("test-id"),
		123,
		time.Now().Add(1*time.Hour),
		time.Now(),
	)

	assert.False(t, session.IsExpired())
}

func TestSession_IsExpired_Expired(t *testing.T) {
	session := NewSession(
		SessionID("test-id"),
		123,
		time.Now().Add(-1*time.Hour),
		time.Now(),
	)

	assert.True(t, session.IsExpired())
}

func TestSession_IsExpired_ExactlyNow(t *testing.T) {
	// Session expiring right now should be considered expired
	session := NewSession(
		SessionID("test-id"),
		123,
		time.Now().Add(-1*time.Millisecond),
		time.Now(),
	)

	// Small delay to ensure time has passed
	time.Sleep(2 * time.Millisecond)
	assert.True(t, session.IsExpired())
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	// IDs should be 32 characters (16 bytes hex encoded)
	assert.Len(t, id1, 32)
	assert.Len(t, id2, 32)

	// IDs should be unique
	assert.NotEqual(t, id1, id2)

	// IDs should be hex strings
	for _, r := range id1 {
		assert.True(t, (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	}
}
