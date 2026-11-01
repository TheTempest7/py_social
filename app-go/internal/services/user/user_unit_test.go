package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/models"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	"github.com/TheTempest7/social_network/app-go/internal/services/user"
	"github.com/TheTempest7/social_network/app-go/internal/testutils"
	"golang.org/x/crypto/bcrypt"
)

// 👇 Юнит-тест 1: email занят
func TestRegisterUser_EmailTaken_Unit(t *testing.T) {
	mock := &testutils.MockUserRepository{
		CreateUserFunc: func(ctx context.Context, user *models.User) (*models.User, error) {
			return nil, repository.ErrEmailTaken
		},
	}
	svc := user.NewUserService(mock)

	input := &user.RegisterInput{
		Email:    "taken@test.com",
		Username: "user",
		Password: "12345678",
	}
	_, err := svc.RegisterUser(context.Background(), input)

	if !errors.Is(err, repository.ErrEmailTaken) {
		t.Errorf("ожидали ErrEmailTaken, получили %v", err)
	}
}

// 👇 Юнит-тест 2: успешная регистрация (проверка логики)
func TestRegisterUser_Success_Unit(t *testing.T) {
	var captured *models.User
	mock := &testutils.MockUserRepository{
		CreateUserFunc: func(ctx context.Context, user *models.User) (*models.User, error) {
			captured = user
			user.ID = 1
			return user, nil
		},
	}
	svc := user.NewUserService(mock)

	input := &user.RegisterInput{
		Email:    "new@test.com",
		Username: "newuser",
		Password: "12345678",
	}
	result, err := svc.RegisterUser(context.Background(), input)
	if err != nil {
		t.Fatalf("ожидали nil, получили %v", err)
	}
	if result.ID != 1 {
		t.Errorf("ожидали ID=1")
	}
	// 👇 Проверяем, что сервис **захешировал** пароль (не передал сырой)
	if captured.PasswordHash == "12345678" {
		t.Error("пароль не захеширован!")
	}
}

// 👇 Юнит-тест: Успешный логин
func TestLoginUser_Success_Unit(t *testing.T) {
	// Настраиваем мок: возвращаем "валидного" юзера с захешированным паролем "12345678"
	mock := &testutils.MockUserRepository{
		GetUserByUsernameFunc: func(ctx context.Context, username string) (*models.User, error) {
			// Хеш сгенерирован заранее для пароля "12345678"
			hash, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.MinCost)
			return &models.User{
				ID:           1,
				Username:     "testuser",
				Email:        "test@test.com",
				PasswordHash: string(hash),
				IsActive:     true,
			}, nil
		},
	}

	svc := user.NewUserService(mock)

	input := user.LoginInput{
		Username: "testuser",
		Password: "12345678",
	}

	// ВАЖНО: Для теста логина (генерации JWT) нужен конфиг
	config.AppConfig = &config.Config{
		JWTSecret:   "test-secret",
		JWTLifetime: 24 * time.Hour,
	}

	user, token, err := svc.LoginUser(context.Background(), input)
	if err != nil {
		t.Fatalf("ожидали nil, получили %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("ожидали 'testuser', получили '%s'", user.Username)
	}
	if token == "" {
		t.Error("ожидали непустой токен")
	}
}

// 👇 Юнит-тест: Неактивный пользователь
func TestLoginUser_Inactive_Unit(t *testing.T) {
	mock := &testutils.MockUserRepository{
		GetUserByUsernameFunc: func(ctx context.Context, username string) (*models.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.MinCost)
			return &models.User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: string(hash),
				IsActive:     false, // 👈 Юзер забанен
			}, nil
		},
	}

	svc := user.NewUserService(mock)

	input := user.LoginInput{
		Username: "testuser",
		Password: "12345678",
	}

	_, _, err := svc.LoginUser(context.Background(), input)

	// Проверяем, что сервис вернул именно ошибку неактивности
	if !errors.Is(err, user.ErrUserInactive) {
		t.Errorf("ожидали ErrUserInactive, получили %v", err)
	}
}
