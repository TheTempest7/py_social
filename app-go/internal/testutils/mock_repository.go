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
	UpdateAvatarFunc      func(ctx context.Context, id int, url string) (*models.User, error)

	CreateEmailOutboxFunc   func(ctx context.Context, userID int, email, subject, body string) (int, error)
	GetEmailByIDFunc        func(ctx context.Context, id int) (*models.EmailOutbox, error)
	UpdateEmailStatusFunc   func(ctx context.Context, id int, status string) error
	IncrementEmailRetryFunc func(ctx context.Context, id int) error
	GetPendingEmailsFunc    func(ctx context.Context, limit int) ([]models.EmailOutbox, error)
}

func (m *MockUserRepository) IncrementEmailRetry(ctx context.Context, id int) error {
	return m.IncrementEmailRetryFunc(ctx, id)
}

func (m *MockUserRepository) UpdateEmailStatus(ctx context.Context, id int, status string) error {
	return m.UpdateEmailStatusFunc(ctx, id, status)
}

func (m *MockUserRepository) GetPendingEmails(ctx context.Context, limit int) ([]models.EmailOutbox, error) {
	return m.GetPendingEmailsFunc(ctx, limit)
}

func (m *MockUserRepository) GetEmailByID(ctx context.Context, id int) (*models.EmailOutbox, error) {
	return m.GetEmailByIDFunc(ctx, id)
}

func (m *MockUserRepository) CreateEmailOutbox(ctx context.Context, userID int, email, subject, body string) (int, error) {
	return m.CreateEmailOutboxFunc(ctx, userID, email, subject, body)
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

func (m *MockUserRepository) UpdateAvatar(ctx context.Context, id int, url string) (*models.User, error) {
	return m.UpdateAvatarFunc(ctx, id, url)
}
