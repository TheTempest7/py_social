// package database

// import (
// 	"database/sql"
// 	"fmt"
// 	"time"

// 	_ "modernc.org/sqlite"
// )

// func ConnectDB(path string) (*sql.DB, error) {
// 	db, err := sql.Open("sqlite", path)
// 	if err != nil {
// 		return nil, fmt.Errorf("не удалось открыть БД: %w", err)
// 	}

// 	db.SetMaxOpenConns(25)
// 	db.SetMaxIdleConns(5)
// 	db.SetConnMaxLifetime(5 * time.Minute)

// 	if err := db.Ping(); err != nil {
// 		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
// 	}

// 	return db, nil
// }

// internal/database/database.go
package database

import (
	"database/sql"
	"fmt"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib" // 👈 Драйвер
)

func ConnectDB(dsn string) (*sql.DB, error) {
	// Для pgx драйвер называется "pgx"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к Postgres: %w", err)
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Postgres не отвечает: %w", err)
	}

	db.SetMaxOpenConns(config.AppConfig.DBMaxOpenConns)
	db.SetMaxIdleConns(config.AppConfig.DBMaxIdleConns)
	db.SetConnMaxLifetime(config.AppConfig.ConnMaxLifetime)

	return db, nil
}
