package database

import (
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbPath, migrationsPath string) error {
	// Если путь уже содержит "file:" (in-memory для тестов) — не добавляем sqlite://
	dsn := dbPath
	if !strings.HasPrefix(dbPath, "file:") && !strings.HasPrefix(dbPath, "sqlite://") {
		dsn = "sqlite://" + dbPath
	}

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("инициализация мигратора: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("применение миграций: %w", err)
	}

	return nil
}
