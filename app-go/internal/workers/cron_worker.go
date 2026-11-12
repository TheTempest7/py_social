package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/TheTempest7/social_network/app-go/internal/broker"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
)

func StartEmailCron(repo repository.UserRepo, queueName string, broker *broker.RabbitMQ) {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		ctx := context.Background()
		emails, err := repo.GetPendingEmails(ctx, 100)
		if err != nil {
			slog.Error("Cron: ошибка получения упавших писем", "error", err)
			continue
		}
		for _, e := range emails {
			payload := map[string]int{"id": e.ID}
			if err := broker.Publish(ctx, queueName, payload); err != nil {
				slog.Error("Cron: не удалось опубликовать в RabbitMQ", "error", err, "email_id", e.ID)
			} else {
				slog.Info("Cron: письмо возвращено в очередь", "email_id", e.ID)
			}
		}
	}
}
