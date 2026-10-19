package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	JWTSecret      string
	JWTLifetime    time.Duration
	CookieSecure   bool
	CookieSameSite string
}

// Глобальная переменная, чтобы не прокидывать её везде
var AppConfig *Config

func Load() {
	AppConfig = &Config{
		JWTSecret:      getEnv("JWT_SECRET", "secret"),
		JWTLifetime:    time.Duration(getEnvInt("JWT_LIFETIME_HOURS", 24)) * time.Hour,
		CookieSecure:   getEnvBool("COOKIE_SECURE", false),
		CookieSameSite: getEnv("COOKIE_SAMESITE", "Lax"),
	}
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
