package repository

import (
	"context"
	"errors"
	"time"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/servicestatus"
	entuser "github.com/starfrag-lab/retrowin-go/ent/user"
	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// EntRepository implements the user.Repository interface using Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository creates a new EntRepository.
func NewEntRepository(client *ent.Client) user.Repository {
	return &EntRepository{client: client}
}

// Create creates a new user.
func (r *EntRepository) Create(ctx context.Context, provider, providerID string) (*user.User, error) {
	entUser, err := r.client.User.Create().
		SetProvider(provider).
		SetProviderID(providerID).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, errors.New("user already exists")
		}
		return nil, err
	}

	return &user.User{
		ID:         int64(entUser.ID),
		Provider:   entUser.Provider,
		ProviderID: entUser.ProviderID,
		CreatedAt:  entUser.CreateTime,
		UpdatedAt:  entUser.UpdateTime,
	}, nil
}

// GetByID retrieves a user by ID.
func (r *EntRepository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	entUser, err := r.client.User.Get(ctx, int(id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &user.User{
		ID:         int64(entUser.ID),
		Provider:   entUser.Provider,
		ProviderID: entUser.ProviderID,
		CreatedAt:  entUser.CreateTime,
		UpdatedAt:  entUser.UpdateTime,
	}, nil
}

// GetByProvider retrieves a user by provider and provider ID.
func (r *EntRepository) GetByProvider(ctx context.Context, provider, providerID string) (*user.User, error) {
	entUser, err := r.client.User.Query().
		Where(
			entuser.ProviderEQ(provider),
			entuser.ProviderIDEQ(providerID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if entUser == nil {
		return nil, nil
	}

	return &user.User{
		ID:         int64(entUser.ID),
		Provider:   entUser.Provider,
		ProviderID: entUser.ProviderID,
		CreatedAt:  entUser.CreateTime,
		UpdatedAt:  entUser.UpdateTime,
	}, nil
}

// Delete deletes a user by ID.
func (r *EntRepository) Delete(ctx context.Context, id int64) error {
	return r.client.User.DeleteOneID(int(id)).Exec(ctx)
}

// ExistsByProvider checks if a user exists by provider and provider ID.
func (r *EntRepository) ExistsByProvider(ctx context.Context, provider, providerID string) (bool, error) {
	return r.client.User.Query().
		Where(
			entuser.ProviderEQ(provider),
			entuser.ProviderIDEQ(providerID),
		).
		Exist(ctx)
}

// EntServiceStatusRepository implements the user.ServiceStatusRepository interface using Ent.
type EntServiceStatusRepository struct {
	client *ent.Client
}

// NewEntServiceStatusRepository creates a new EntServiceStatusRepository.
func NewEntServiceStatusRepository(client *ent.Client) user.ServiceStatusRepository {
	return &EntServiceStatusRepository{client: client}
}

// Create creates a new service status.
func (r *EntServiceStatusRepository) Create(ctx context.Context, userID int64) (*user.ServiceStatus, error) {
	status, err := r.client.ServiceStatus.Create().
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return &user.ServiceStatus{
		UserID:     status.UserID,
		Available:  status.Available,
		JoinDate:   status.JoinDate,
		UpdateDate: status.UpdateDate,
	}, nil
}

// GetByUserID retrieves a service status by user ID.
func (r *EntServiceStatusRepository) GetByUserID(ctx context.Context, userID int64) (*user.ServiceStatus, error) {
	status, err := r.client.ServiceStatus.Query().
		Where(servicestatus.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if status == nil {
		return nil, nil
	}

	return &user.ServiceStatus{
		UserID:     status.UserID,
		Available:  status.Available,
		JoinDate:   status.JoinDate,
		UpdateDate: status.UpdateDate,
	}, nil
}

// Update updates a service status.
func (r *EntServiceStatusRepository) Update(ctx context.Context, userID int64, available bool) (*user.ServiceStatus, error) {
	_, err := r.client.ServiceStatus.Update().
		Where(servicestatus.UserIDEQ(userID)).
		SetAvailable(available).
		SetUpdateDate(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.GetByUserID(ctx, userID)
}

// Delete deletes a service status by user ID.
func (r *EntServiceStatusRepository) Delete(ctx context.Context, userID int64) error {
	_, err := r.client.ServiceStatus.Delete().
		Where(servicestatus.UserIDEQ(userID)).
		Exec(ctx)
	return err
}
