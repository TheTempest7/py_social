package workers

import (
	"log/slog"
	"time"
)

type EmailTask struct {
	To      string
	Subject string
	Body    string
}

// StartEmailWorkers запускает пул воркеров
// numWorkers — сколько горутин будет одновременно отправлять письма
// tasks — канал, откуда воркеры берут задачи
func StartEmailWorkers(numWorkers int, tasks <-chan EmailTask) {
	for i := 0; i < numWorkers; i++ {
		// Запускаем горутину для каждого воркера
		go func(workerID int) {
			// Воркер работает бесконечно, читая канал
			for task := range tasks {
				slog.Info("Воркер начал отправку email", "worker_id", workerID, "to", task.To)

				// Имитируем долгую отправку (в реальности тут smtp.SendMail)
				time.Sleep(2 * time.Second)

				slog.Info("Воркер успешно отправил email", "worker_id", workerID, "to", task.To)
			}
		}(i)
	}
}
