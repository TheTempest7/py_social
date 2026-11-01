package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/repository"

	_ "modernc.org/sqlite"
)

// SetupDB — создает in-memory БД, применяет миграции и возвращает репозиторий.
// Путь к миграциям указывается относительно корня проекта (где go.mod).
func SetupDB(t *testing.T) *repository.UserRepository {
	t.Helper()

	// Загружаем конфиг (он нужен для JWT в сервисах)
	config.AppConfig = &config.Config{
		JWTSecret:      "test-secret",
		JWTLifetime:    24 * time.Hour,
		CookieSecure:   false,
		CookieSameSite: "Lax",
	}

	// Создаем уникальную БД
	dbName := fmt.Sprintf("file:test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		t.Fatalf("БД не открылась: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Применяем миграции
	applyMigrations(t, db)

	// Возвращаем репозиторий
	return repository.NewUserRepository(db)
}

// applyMigrations — ищет файлы миграций от корня проекта.
func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	// Идем от текущей папки вверх, пока не найдем папку migrations.
	// Это нужно, потому что тесты лежат на разной глубине (services/user, services/profile).
	wd, _ := os.Getwd()
	migrationsPath := ""

	for {
		path := filepath.Join(wd, "migrations")
		if _, err := os.Stat(path); err == nil {
			migrationsPath = path
			break
		}
		// Поднимаемся на уровень выше
		parent := filepath.Dir(wd)
		if parent == wd {
			// Дошли до корня диска, а папку не нашли
			t.Fatalf("не удалось найти папку migrations от %s", wd)
		}
		wd = parent
	}

	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		t.Fatalf("поиск миграций: %v", err)
	}
	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("чтение %s: %v", file, err)
		}

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
