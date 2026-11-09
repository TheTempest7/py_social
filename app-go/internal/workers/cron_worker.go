package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/TheTempest7/social_network/app-go/internal/repository"
)

func StartEmailCron(repo repository.UserRepo, queue chan<- EmailTask) {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		ctx := context.Background()
		emails, err := repo.GetPendingEmails(ctx, 100)
		if err != nil {
			slog.Error("Cron: ошибка получения упавших писем", "error", err)
			continue
		}
		for _, e := range emails {
			select {
			case queue <- EmailTask{ID: e.ID}:
				// Успешно вернули в очередь
			default:
				// Очередь все еще забита, попробуем в следующий раз
			}
		}
	}
}
