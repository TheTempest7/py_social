package services_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/database"
	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	"github.com/TheTempest7/social_network/app-go/internal/services"
)

// setupService — создаёт чистый Service с изолированной БД в памяти.
// БД автоматически закрывается после завершения теста.
func setupService(t *testing.T) *services.UserService {
	t.Helper() // помечает функцию как хелпер, чтобы ошибки указывали на строку вызова

	// 👇 Уникальное имя БД на каждый тест, чтобы тесты не влияли друг на друга
	dbName := fmt.Sprintf("file:test_%d?mode=memory&cache=shared", time.Now().UnixNano())

	config.AppConfig = &config.Config{
		JWTSecret:      "test-secret",
		JWTLifetime:    24 * time.Hour,
		CookieSecure:   false,
		CookieSameSite: "Lax",
	}

	db, err := database.ConnectDB(dbName)
	if err != nil {
		t.Fatalf("БД не поднялась: %v", err)
	}
	t.Cleanup(func() { db.Close() }) // 👈 закроется автоматически после теста

	if err := database.RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("миграции: %v", err)
	}

	repo := repository.NewUserRepository(db)
	return services.NewUserService(repo)
}

func TestRegisterUserRaceCondition(t *testing.T) {
	svc := setupService(t)

	// 3. 100 горутин — одинаковый email
	const goroutines = 100
	var wg sync.WaitGroup
	var (
		successes int
		conflicts int
		others    int
		mu        sync.Mutex
	)

	req := dto.RegisterRequest{
		Email:    "race@test.com",
		Username: "race",
		Password: "12345678",
	}

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.RegisterUser(req)

			mu.Lock()
			defer mu.Unlock()

			switch {
			case err == nil:
				successes++
			case errors.Is(err, repository.ErrEmailTaken),
				errors.Is(err, repository.ErrUsernameTaken):
				conflicts++
			default:
				others++
				t.Logf("неожиданная ошибка: %v", err)
			}
		}()
	}
	wg.Wait()

	// 4. Проверяем
	if successes != 1 {
		t.Errorf("ожидали 1 успех, получили %d", successes)
	}
	if conflicts != goroutines-1 {
		t.Errorf("ожидали %d конфликтов, получили %d", goroutines-1, conflicts)
	}
	if others != 0 {
		t.Errorf("ожидали 0 прочих ошибок, получили %d", others)
	}
}
func TestLoginUserSuccess(t *testing.T) {
	svc := setupService(t)

	// Сначала регистрируем
	regReq := dto.RegisterRequest{
		Email:    "login@test.com",
		Username: "loginuser",
		Password: "12345678",
	}
	_, err := svc.RegisterUser(regReq)
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	// Теперь логинимся
	user, token, err := svc.LoginUser(dto.LoginRequest{
		Username: "loginuser",
		Password: "12345678",
	})
	if err != nil {
		t.Fatalf("логин: %v", err)
	}

	if user == nil {
		t.Fatal("user == nil")
	}
	if user.Username != "loginuser" {
		t.Errorf("ожидали 'loginuser', получили '%s'", user.Username)
	}
	if token == "" {
		t.Error("токен пустой")
	}
}

func TestLoginUserWrongPassword(t *testing.T) {
	svc := setupService(t)

	_, err := svc.RegisterUser(dto.RegisterRequest{
		Email:    "wrong@test.com",
		Username: "wrongpass",
		Password: "correct123",
	})
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	_, _, err = svc.LoginUser(dto.LoginRequest{
		Username: "wrongpass",
		Password: "wrong123",
	})

	if !errors.Is(err, services.ErrInvalidCreds) {
		t.Errorf("ожидали ErrInvalidCreds, получили %v", err)
	}
}

func TestGetByIDSuccess(t *testing.T) {
	svc := setupService(t)

	created, err := svc.RegisterUser(dto.RegisterRequest{
		Email:    "getbyid@test.com",
		Username: "getbyid",
		Password: "12345678",
	})
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	found, err := svc.GetByID(created.ID)
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
	svc := setupService(t)

	_, err := svc.GetByID(99999)
	if err == nil {
		t.Fatal("ожидали ошибку, получили nil")
	}
}
