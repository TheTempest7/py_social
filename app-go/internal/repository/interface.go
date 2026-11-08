package repository

import (
	"context"

	"github.com/TheTempest7/social_network/app-go/internal/models"
)

// UserRepo — интерфейс репозитория (по месту использования)
type UserRepo interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateProfile(ctx context.Context, newData *models.User) (*models.User, error)
	UpdateAvatar(ctx context.Context, id int, url string) (*models.User, error)
	// 👇 Новые методы для работы с Email Outbox
	CreateEmailOutbox(ctx context.Context, userID int, email, subject, body string) (int, error)
	GetEmailByID(ctx context.Context, id int) (*models.EmailOutbox, error)
	UpdateEmailStatus(ctx context.Context, id int, status string) error
	IncrementEmailRetry(ctx context.Context, id int) error
	GetPendingEmails(ctx context.Context, limit int) ([]models.EmailOutbox, error)
}
