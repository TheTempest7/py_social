package workers

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
)

type EmailTask struct {
	ID int
}

// StartEmailWorkers запускает пул воркеров
// numWorkers — сколько горутин будет одновременно отправлять письма
// tasks — канал, откуда воркеры берут задачи
func StartEmailWorkers(numWorkers int, tasks <-chan EmailTask, repo repository.UserRepo) {
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for task := range tasks {
				ctx := context.Background()

				// 1. Достаем письмо из БД
				email, err := repo.GetEmailByID(ctx, task.ID)
				if err != nil {
					slog.Error("Воркер не нашел письмо", "error", err, "id", task.ID)
					continue
				}

				// 2. Формируем и отправляем
				msg := fmt.Appendf(nil, "To: %s\r\nSubject: %s\r\n\r\n%s\r\n", email.Email, email.Subject, email.Body)

				addr := fmt.Sprintf("%s:%d", config.AppConfig.SMTPHost, config.AppConfig.SMTPPort)
				auth := smtp.PlainAuth("", config.AppConfig.SMTPUser, config.AppConfig.SMTPPass, config.AppConfig.SMTPHost)

				err = smtp.SendMail(addr, auth, config.AppConfig.SMTPFrom, []string{email.Email}, msg)

				// 3. Обновляем статус
				if err != nil {
					slog.Error("Ошибка отправки email", "error", err, "id", task.ID, "worker", workerID)
					repo.IncrementEmailRetry(ctx, task.ID) // ставим failed, retries++
				} else {
					slog.Info("Email успешно отправлен", "id", task.ID, "worker", workerID)
					repo.UpdateEmailStatus(ctx, task.ID, "sent")
				}
			}
		}(i)
	}
}
