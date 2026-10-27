package profile_test

import (
	"context"
	"testing"

	"github.com/TheTempest7/social_network/app-go/internal/models"
	"github.com/TheTempest7/social_network/app-go/internal/services/profile"
	"github.com/TheTempest7/social_network/app-go/internal/services/user"
	"github.com/TheTempest7/social_network/app-go/internal/testutils"
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

func TestUpdateProfile_Success_Unit(t *testing.T) {
	// 1. Настраиваем мок
	mock := &testutils.MockUserRepository{
		// Когда сервис попросит юзера, отдадим старого
		GetUserByIDFunc: func(ctx context.Context, id int) (*models.User, error) {
			return &models.User{ID: 1, FirstName: "Old", LastName: "Name"}, nil
		},
		// Когда сервис попросит обновить, проверим, что он передал новые данные
		UpdateProfileFunc: func(ctx context.Context, user *models.User) (*models.User, error) {
			if user.FirstName != "New" {
				t.Errorf("ожидали 'New', получили '%s'", user.FirstName)
			}
			return user, nil
		},
	}

	svc := profile.NewProfileService(mock)

	// 2. Даем новые данные (указатели)
	newName := "New"
	input := profile.UpdateProfileInput{
		FirstName: &newName,
		LastName:  nil, // LastName не меняем
	}

	// 3. Вызываем
	result, err := svc.UpdateProfile(context.Background(), 1, input)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if result.FirstName != "New" {
		t.Errorf("ожидали 'New', получили '%s'", result.FirstName)
	}
}
