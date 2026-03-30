package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/starfrag-lab/retrowin-go/internal/user"
	userMocks "github.com/starfrag-lab/retrowin-go/internal/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("returns user when found", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		expectedUser := &user.User{
			ID:         "user-123",
			Username:   "testuser",
			Provider:   user.ProviderKeycloak,
			ProviderID: "keycloak-123",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "keycloak-123").Return(expectedUser, nil)

		result, err := svc.Get(ctx, user.ProviderKeycloak, "keycloak-123")

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
	})

	t.Run("returns error when user not found", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "keycloak-123").Return(nil, nil)

		result, err := svc.Get(ctx, user.ProviderKeycloak, "keycloak-123")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "keycloak-123").Return(nil, errors.New("db error"))

		result, err := svc.Get(ctx, user.ProviderKeycloak, "keycloak-123")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestService_FindOrCreateByOIDC(t *testing.T) {
	ctx := context.Background()

	t.Run("returns existing user", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		existingUser := &user.User{
			ID:         "user-123",
			Username:   "testuser",
			Provider:   user.ProviderKeycloak,
			ProviderID: "subject-123",
		}
		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "subject-123").Return(existingUser, nil)

		userID, err := svc.FindOrCreateByOIDC(ctx, user.ProviderKeycloak, "subject-123", "testuser")

		assert.NoError(t, err)
		assert.Equal(t, "user-123", userID)
	})

	t.Run("creates new user when not found", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "subject-123").Return(nil, nil)
		userRepo.EXPECT().ExistsByProvider(mock.Anything, user.ProviderKeycloak, "subject-123").Return(false, nil)
		newUser := &user.User{
			ID:         "user-456",
			Username:   "testuser",
			Provider:   user.ProviderKeycloak,
			ProviderID: "subject-123",
		}
		userRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(newUser, nil)

		userID, err := svc.FindOrCreateByOIDC(ctx, user.ProviderKeycloak, "subject-123", "testuser")

		assert.NoError(t, err)
		assert.Equal(t, "user-456", userID)
	})
}
