package testutils

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib" // 👈 Драйвер
)

func resetSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	if err != nil {
		t.Fatalf("очистка схемы БД: %v", err)
	}
}

// SetupDB — создает in-memory БД, применяет миграции и возвращает репозиторий.
// Путь к миграциям указывается относительно корня проекта (где go.mod).
func SetupDB(t *testing.T) *repository.UserRepository {
	t.Helper()

	// 1. Загружаем .env (переменные попадут в os.Getenv)
	LoadTestEnv(t)

	// Загружаем конфиг (он нужен для JWT в сервисах)
	// 2. ЧИСТОЕ РЕШЕНИЕ: На время теста принудительно меняем хост на локальный.
	// t.Setenv автоматически вернет всё назад после завершения теста.
	t.Setenv("DB_NAME", "social_network_test")

	// Теперь хост меняем на локальный (для Docker)
	t.Setenv("DB_HOST", "127.0.0.1")
	// 3. Загружаем конфиг (он возьмет DB_HOST=127.0.0.1 из окружения)
	config.Load()

	dsn := config.AppConfig.DSN()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("БД не открылась: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// 4. Проверяем, что реально подключились (иначе truncate упадет с непонятной ошибкой)
	if err := db.Ping(); err != nil {
		t.Fatalf("БД не отвечает (%s): %v", dsn, err)
	}

	resetSchema(t, db)

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
