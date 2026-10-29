package testutils

import (
	"testing"

	"github.com/TheTempest7/social_network/app-go/internal/storage"
)

// SetupStorage — создает фейковый Storage для тестов
func SetupStorage(t *testing.T) *storage.Storage {
	t.Helper()
	// Возвращаем nil или пустую структуру, так как для тестов GetByID Storage не нужен
	return nil
}
