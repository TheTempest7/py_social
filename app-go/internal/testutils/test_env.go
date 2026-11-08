package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

// LoadTestEnv ищет .env вверх по дереву и загружает его в окружение.
func LoadTestEnv(t *testing.T) {
	t.Helper()
	wd, _ := os.Getwd()
	for {
		// Ищем .env (он лежит в корне app-go)
		envPath := filepath.Join(wd, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				t.Fatalf("ошибка загрузки .env (%s): %v", envPath, err)
			}
			return // Нашли и загрузили, выходим
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("не удалось найти .env вверх по дереву от теста")
		}
		wd = parent
	}
}
