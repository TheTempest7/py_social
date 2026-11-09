package app

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/database"
	"github.com/TheTempest7/social_network/app-go/internal/storage"
)

// InitDB подключается к базе и прогоняет миграции
func InitDB() (*sql.DB, error) {
	db, err := database.ConnectDB(config.AppConfig.DSN())
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := database.RunMigrations(config.AppConfig.DSN(), config.AppConfig.MigrationsPath); err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка миграций: %w", err)
	}

	slog.Info("✅ Подключено к PostgreSQL и миграции применены")
	return db, nil
}

// InitStorage подключается к MinIO
func InitStorage() (*storage.Storage, error) {
	s, err := storage.NewStorage(
		config.AppConfig.MinioEndpoint,
		config.AppConfig.MinioAccessKey,
		config.AppConfig.MinioSecretKey,
		config.AppConfig.MinioBucket,
		config.AppConfig.MinioUseSSL,
		config.AppConfig.MinioRegion,
		config.AppConfig.MinioPublicURL,
		config.AppConfig.MinioMaxConcurrentRequests,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MinIO: %w", err)
	}
	slog.Info("✅ Подключено к MinIO")
	return s, nil
}
