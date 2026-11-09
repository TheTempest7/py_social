package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	JWTSecret      string
	JWTLifetime    time.Duration
	CookieSecure   bool
	CookieSameSite string
	DBPath         string
	MigrationsPath string
	RedisAddr      string
	RedisPassword  string
	ServerPort     string

	// MinIO
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	MinioRegion    string
	MinioPublicURL string

	DatabaseURL     string
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	ConnMaxLifetime time.Duration

	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     int
	DBName     string
	DBSSLMode  string

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	MinioMaxConcurrentRequests int
}

// Глобальная переменная, чтобы не прокидывать её везде
var AppConfig *Config

func Load() {
	AppConfig = &Config{
		JWTSecret:   getEnv("JWT_SECRET", "secret"),
		JWTLifetime: time.Duration(getEnvInt("JWT_LIFETIME_HOURS", 24)) * time.Hour,

		CookieSecure:   getEnvBool("COOKIE_SECURE", false),
		CookieSameSite: getEnv("COOKIE_SAMESITE", "Lax"),

		MigrationsPath: getEnv("MIGRATIONS_PATH", "migrations"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		ServerPort: getEnv("SERVER_PORT", "8080"),

		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getEnv("MINIO_BUCKET", "avatars"),
		MinioUseSSL:    getEnvBool("MINIO_USE_SSL", false),
		MinioRegion:    getEnv("MINIO_REGION", "localhost"),
		MinioPublicURL: getEnv("MINIO_PUBLIC_URL", "localhost:9032"),

		DatabaseURL:     getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/social_network?sslmode=disable"),
		DBMaxOpenConns:  getEnvInt("DB_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns:  getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvInt("DB_PORT", 5432),
		DBName:     getEnv("DB_NAME", "social_network"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		SMTPHost: getEnv("SMTP_HOST", "localhost"),
		SMTPPort: getEnvInt("SMTP_PORT", 1025),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		SMTPFrom: getEnv("SMTP_FROM", "noreply@social.network"),

		MinioMaxConcurrentRequests: getEnvInt("MINIO_MAX_CONCURRENT_REQUESTS", 12),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
