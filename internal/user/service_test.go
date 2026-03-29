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
			ID:         123,
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
		assert.Equal(t, user.ErrUserNotFound, err)
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

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns user when found", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		expectedUser := &user.User{
			ID:         123,
			Provider:   user.ProviderKeycloak,
			ProviderID: "keycloak-123",
		}
		userRepo.EXPECT().GetByID(mock.Anything, int64(123)).Return(expectedUser, nil)

		result, err := svc.GetByID(ctx, 123)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
	})

	t.Run("returns error when user not found", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		userRepo.EXPECT().GetByID(mock.Anything, int64(123)).Return(nil, nil)

		result, err := svc.GetByID(ctx, 123)

		assert.Error(t, err)
		assert.Equal(t, user.ErrUserNotFound, err)
		assert.Nil(t, result)
	})
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("creates user successfully", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		expectedUser := &user.User{
			ID:         123,
			Provider:   user.ProviderKeycloak,
			ProviderID: "keycloak-123",
		}
		userRepo.EXPECT().ExistsByProvider(mock.Anything, user.ProviderKeycloak, "keycloak-123").Return(false, nil)
		userRepo.EXPECT().Create(mock.Anything, user.ProviderKeycloak, "keycloak-123").Return(expectedUser, nil)

		cmd := &user.CreateCommand{
			Provider:   user.ProviderKeycloak,
			ProviderID: "keycloak-123",
		}
		result, err := svc.Create(ctx, cmd)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
	})

	t.Run("returns error when provider is empty", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		cmd := &user.CreateCommand{
			Provider:   "",
			ProviderID: "keycloak-123",
		}

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, "provider is required", err.Error())
		assert.Nil(t, result)
	})

	t.Run("returns error when providerID is empty", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		cmd := &user.CreateCommand{
			Provider:   user.ProviderKeycloak,
			ProviderID: "",
		}

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, "providerId is required", err.Error())
		assert.Nil(t, result)
	})

	t.Run("returns error when provider is invalid", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		cmd := &user.CreateCommand{
			Provider:   "invalid",
			ProviderID: "keycloak-123",
		}

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, user.ErrInvalidProvider, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when user already exists", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		cmd := &user.CreateCommand{
			Provider:   user.ProviderKeycloak,
			ProviderID: "keycloak-123",
		}

		userRepo.EXPECT().ExistsByProvider(mock.Anything, user.ProviderKeycloak, "keycloak-123").Return(true, nil)

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, user.ErrUserAlreadyExists, err)
		assert.Nil(t, result)
	})
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes user successfully", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		u := &user.User{
			ID:         123,
			Provider:   user.ProviderKeycloak,
			ProviderID: "keycloak-123",
		}
		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "keycloak-123").Return(u, nil)
		userRepo.EXPECT().Delete(mock.Anything, int64(123)).Return(nil)

		err := svc.Delete(ctx, user.ProviderKeycloak, "keycloak-123")

		assert.NoError(t, err)
	})

	t.Run("returns error when user not found", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "keycloak-123").Return(nil, nil)

		err := svc.Delete(ctx, user.ProviderKeycloak, "keycloak-123")

		assert.Error(t, err)
		assert.Equal(t, user.ErrUserNotFound, err)
	})
}

func TestService_FindOrCreateByOIDC(t *testing.T) {
	ctx := context.Background()

	t.Run("returns existing user", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		existingUser := &user.User{
			ID:         123,
			UID:        "user-uid-123",
			Provider:   user.ProviderKeycloak,
			ProviderID: "subject-123",
		}
		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "subject-123").Return(existingUser, nil)

		userID, userUID, err := svc.FindOrCreateByOIDC(ctx, user.ProviderKeycloak, "subject-123", "test@example.com", "Test User", "")

		assert.NoError(t, err)
		assert.Equal(t, int64(123), userID)
		assert.NotEmpty(t, userUID)
	})

	t.Run("creates new user when not found", func(t *testing.T) {
		userRepo := userMocks.NewRepositoryMock(t)
		svc := user.NewService(userRepo)

		userRepo.EXPECT().GetByProvider(mock.Anything, user.ProviderKeycloak, "subject-123").Return(nil, nil)
		userRepo.EXPECT().ExistsByProvider(mock.Anything, user.ProviderKeycloak, "subject-123").Return(false, nil)
		newUser := &user.User{
			ID:         456,
			UID:        "user-uid-456",
			Provider:   user.ProviderKeycloak,
			ProviderID: "subject-123",
		}
		userRepo.EXPECT().Create(mock.Anything, user.ProviderKeycloak, "subject-123").Return(newUser, nil)

	userID, userUID, err := svc.FindOrCreateByOIDC(ctx, user.ProviderKeycloak, "subject-123", "test@example.com", "Test User", "")

		assert.NoError(t, err)
		assert.Equal(t, int64(456), userID)
		assert.NotEmpty(t, userUID)
	})
}
