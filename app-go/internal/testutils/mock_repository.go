package testutils

import (
	"context"

	"github.com/TheTempest7/social_network/app-go/internal/models"
)

// MockUserRepository — мок (уже есть?)
type MockUserRepository struct {
	CreateUserFunc        func(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByIDFunc       func(ctx context.Context, id int) (*models.User, error)
	GetUserByUsernameFunc func(ctx context.Context, username string) (*models.User, error)
	UpdateProfileFunc     func(ctx context.Context, user *models.User) (*models.User, error)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	return m.CreateUserFunc(ctx, user)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return m.GetUserByIDFunc(ctx, id)
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return m.GetUserByUsernameFunc(ctx, username)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, user *models.User) (*models.User, error) {
	return m.UpdateProfileFunc(ctx, user)
}
