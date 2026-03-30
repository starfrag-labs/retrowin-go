package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected bool
	}{
		{
			name:     "keycloak provider is valid",
			provider: ProviderKeycloak,
			expected: true,
		},
		{
			name:     "google provider is valid",
			provider: ProviderGoogle,
			expected: true,
		},
		{
			name:     "empty provider is invalid",
			provider: "",
			expected: false,
		},
		{
			name:     "unknown provider is invalid",
			provider: "unknown",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidProvider(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_Struct(t *testing.T) {
	u := User{
		ID:         "user-123",
		Username:   "testuser",
		Provider:   ProviderKeycloak,
		ProviderID: "keycloak-123",
	}

	assert.Equal(t, "user-123", u.ID())
	assert.Equal(t, "testuser", u.Username())
	assert.Equal(t, ProviderKeycloak, u.Provider())
	assert.Equal(t, "keycloak-123", u.ProviderID())
}
