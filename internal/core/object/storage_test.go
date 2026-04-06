package object

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpiryForSize(t *testing.T) {
	const mb = 1 << 20

	tests := []struct {
		name     string
		size     int64
		expected time.Duration
	}{
		{"zero", 0, 15 * time.Minute},
		{"1 byte", 1, 15 * time.Minute},
		{"5MB", 5 * mb, 15 * time.Minute},
		{"exactly 10MB", 10 * mb, 15 * time.Minute},
		{"10MB + 1", 10*mb + 1, 1 * time.Hour},
		{"50MB", 50 * mb, 1 * time.Hour},
		{"exactly 100MB", 100 * mb, 1 * time.Hour},
		{"100MB + 1", 100*mb + 1, 3 * time.Hour},
		{"500MB", 500 * mb, 3 * time.Hour},
		{"exactly 1GB", 1024 * mb, 3 * time.Hour},
		{"1GB + 1", 1024*mb + 1, 6 * time.Hour},
		{"10GB", 10 * 1024 * mb, 6 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ExpiryForSize(tt.size))
		})
	}
}

func TestExpiryForSize_WithinBounds(t *testing.T) {
	const mb = 1 << 20

	// All returned values should be within [MinExpiry, MaxExpiry]
	sizes := []int64{0, 1, 10 * mb, 100 * mb, 1024 * mb, 10 * 1024 * mb}
	for _, size := range sizes {
		expiry := ExpiryForSize(size)
		assert.GreaterOrEqual(t, expiry, MinExpiry, "size=%d", size)
		assert.LessOrEqual(t, expiry, MaxExpiry, "size=%d", size)
	}
}
