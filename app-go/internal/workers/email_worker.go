package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/TheTempest7/social_network/app-go/internal/broker"
	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
)

type EmailTask struct {
	ID int
}

// StartEmailWorkers запускает пул воркеров

func StartEmailWorkers(numWorkers int, rabbit *broker.RabbitMQ, repo repository.UserRepo, queueName string) {
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			// 👇 Имя очереди берем из аргумента queueName
			err := rabbit.Consume(context.Background(), queueName, func(body []byte) error {

				// 👇 Парсим JSON из RabbitMQ
				var data struct {
					ID int `json:"id"`
				}
				if err := json.Unmarshal(body, &data); err != nil {
					return fmt.Errorf("не удалось распарсить: %w", err)
				}

				ctx := context.Background()
				// 1. Достаем письмо из БД
				email, err := repo.GetEmailByID(ctx, data.ID)
				if err != nil {
					return err
				}

				// 2. Отправляем SMTP
				msg := fmt.Appendf(nil, "To: %s\r\nSubject: %s\r\n\r\n%s\r\n", email.Email, email.Subject, email.Body)
				addr := fmt.Sprintf("%s:%d", config.AppConfig.SMTPHost, config.AppConfig.SMTPPort)

				var auth smtp.Auth
				if config.AppConfig.SMTPUser != "" {
					auth = smtp.PlainAuth("", config.AppConfig.SMTPUser, config.AppConfig.SMTPPass, config.AppConfig.SMTPHost)
				}

				err = smtp.SendMail(addr, auth, config.AppConfig.SMTPFrom, []string{email.Email}, msg)
				if err != nil {
					repo.IncrementEmailRetry(ctx, data.ID) // Ставим failed в БД
					return err                             // NACK
				}

				// 3. Успех
				repo.UpdateEmailStatus(ctx, data.ID, "sent") // ACK
				return nil
			})

			if err != nil {
				slog.Error("Воркер остановлен", "error", err, "worker_id", workerID)
			}
		}(i)
	}
}
