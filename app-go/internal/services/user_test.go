package services_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	"github.com/TheTempest7/social_network/app-go/internal/services"

	_ "modernc.org/sqlite"
)

// setupService — создаёт чистый Service с изолированной БД в памяти.
// БД автоматически закрывается после завершения теста.
func setupService(t *testing.T) *services.UserService {
	t.Helper() // помечает функцию как хелпер, чтобы ошибки указывали на строку вызова

	config.AppConfig = &config.Config{
		JWTSecret:      "test-secret",
		JWTLifetime:    24 * time.Hour,
		CookieSecure:   false,
		CookieSameSite: "Lax",
	}

	// Создаём уникальную in-memory БД для каждого теста
	dbName := fmt.Sprintf("file:test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		t.Fatalf("БД не открылась: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Применяем миграции. ВАЖНО: передаем dbName прямо как есть (с file:...)
	// И указываем путь к папке с миграциями
	applyMigrationsForTest(t, db)
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

	req := &services.RegisterInput{
		Email:    "race@test.com",
		Username: "race",
		Password: "12345678",
	}

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.RegisterUser(context.Background(), req)

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
	regReq := &services.RegisterInput{
		Email:    "login@test.com",
		Username: "loginuser",
		Password: "12345678",
	}
	_, err := svc.RegisterUser(context.Background(), regReq)
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	// Теперь логинимся
	user, token, err := svc.LoginUser(context.Background(), services.LoginInput{
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

	_, err := svc.RegisterUser(context.Background(), &services.RegisterInput{
		Email:    "wrong@test.com",
		Username: "wrongpass",
		Password: "correct123",
	})
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	_, _, err = svc.LoginUser(context.Background(), services.LoginInput{
		Username: "wrongpass",
		Password: "wrong123",
	})

	if !errors.Is(err, services.ErrInvalidCreds) {
		t.Errorf("ожидали ErrInvalidCreds, получили %v", err)
	}
}

func TestGetByIDSuccess(t *testing.T) {
	svc := setupService(t)

	created, err := svc.RegisterUser(context.Background(), &services.RegisterInput{
		Email:    "getbyid@test.com",
		Username: "getbyid",
		Password: "12345678",
	})
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	found, err := svc.GetByID(context.Background(), created.ID)
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

	_, err := svc.GetByID(context.Background(), 99999)
	if err == nil {
		t.Fatal("ожидали ошибку, получили nil")
	}
}

func applyMigrationsForTest(t *testing.T, db *sql.DB) {
	t.Helper()

	// Читаем файлы миграций
	files, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatalf("поиск миграций: %v", err)
	}
	sort.Strings(files) // сортируем по порядку

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("чтение %s: %v", file, err)
		}

		// Выполняем SQL по одной команде
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("выполнение %s: %v", file, err)
			}
		}
	}
}
