package database

import (
	"database/sql"
	"fmt"
	"os"
)

func RunMigrations(db *sql.DB, migrationsPath string) error {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("чтение миграций: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := os.ReadFile(migrationsPath + "/" + file.Name())
		if err != nil {
			return fmt.Errorf("чтение %s: %w", file.Name(), err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("выполнение %s: %w", file.Name(), err)
		}

		fmt.Println("✅ Миграция применена:", file.Name())
	}

	return nil
}
