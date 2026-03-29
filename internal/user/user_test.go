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
		{
			name:     "case sensitive - Keycloak is invalid",
			provider: "Keycloak",
			expected: false,
		},
		{
			name:     "case sensitive - GOOGLE is invalid",
			provider: "GOOGLE",
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
	user := User{
		ID:         123,
		Provider:   ProviderKeycloak,
		ProviderID: "keycloak-123",
	}

	assert.Equal(t, int64(123), user.ID)
	assert.Equal(t, ProviderKeycloak, user.Provider)
	assert.Equal(t, "keycloak-123", user.ProviderID)
}

func TestCreateCommand_Struct(t *testing.T) {
	cmd := CreateCommand{
		Provider:   ProviderKeycloak,
		ProviderID: "keycloak-123",
	}

	assert.Equal(t, ProviderKeycloak, cmd.Provider)
	assert.Equal(t, "keycloak-123", cmd.ProviderID)
}
