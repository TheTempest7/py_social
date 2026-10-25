package profile_test

import (
	"context"
	"testing"

	"github.com/TheTempest7/social_network/app-go/internal/services/profile"
	"github.com/TheTempest7/social_network/app-go/internal/services/user"
	testutils "github.com/TheTempest7/social_network/app-go/internal/test_utils"
)

func TestGetByIDSuccess(t *testing.T) {
	// 👇 Получаем готовый репозиторий с БД
	repo := testutils.SetupDB(t)

	// Инициализируем сервисы
	userSvc := user.NewUserService(repo)
	profileSvc := profile.NewProfileService(repo)
	// 1. Создаем юзера через UserService
	created, err := userSvc.RegisterUser(context.Background(), &user.RegisterInput{
		Email:    "getbyid@test.com",
		Username: "getbyid",
		Password: "12345678",
	})
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	// 2. Достаем его через ProfileService
	found, err := profileSvc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("ожидали ID %d, получили %d", created.ID, found.ID)
	}
	if found.Username != "getbyid" {
		t.Errorf("ожидали 'getbyid', получили '%s'", found.Username)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	// 👇 Получаем готовый репозиторий с БД
	repo := testutils.SetupDB(t)
	profileSvc := profile.NewProfileService(repo)

	_, err := profileSvc.GetByID(context.Background(), 99999)
	if err == nil {
		t.Fatal("ожидали ошибку, получили nil")
	}
}
