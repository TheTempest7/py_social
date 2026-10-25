package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Blacklist — обертка над Redis для хранения отозванных токенов
type Blacklist struct {
	rdb *redis.Client
}

// NewBlacklist — создает подключение к Redis
func NewBlacklist(addr, password string) *Blacklist {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &Blacklist{rdb: rdb}
}

// Add — добавляет токен в черный список на время его жизни
func (b *Blacklist) Add(ctx context.Context, token string, expiration time.Duration) error {
	// Мы ставим значение "1" и TTL (например, 24 часа).
	// Redis сам удалит его, когда время выйдет.
	return b.rdb.Set(ctx, "blacklist:"+token, "1", expiration).Err()
}

// IsBlacklisted — проверяет наличие токена в Redis
func (b *Blacklist) IsBlacklisted(ctx context.Context, token string) bool {
	val, err := b.rdb.Get(ctx, "blacklist:"+token).Result()
	if err == redis.Nil {
		return false // токена нет
	}
	if err != nil {
		return false // ошибка Redis (в проде надо логировать)
	}
	return val == "1"
}
