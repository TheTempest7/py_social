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
}
