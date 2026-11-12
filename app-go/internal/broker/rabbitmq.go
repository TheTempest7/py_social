package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/TheTempest7/social_network/app-go/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

const QueueName = "email_tasks"

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQ создает подключение к RabbitMQ
func NewRabbitMQ() (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		config.AppConfig.RabbitMQUser,
		config.AppConfig.RabbitMQPassword,
		config.AppConfig.RabbitMQHost,
		config.AppConfig.RabbitMQPort,
		config.AppConfig.RabbitMQVhost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ошибка открытия канала: %w", err)
	}

	// Объявляем очередь (если ее нет, она создастся)
	_, err = ch.QueueDeclare(
		QueueName, // name
		true,      // durable (сохраняется на диске)
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("ошибка объявления очереди: %w", err)
	}

	slog.Info("✅ Подключено к RabbitMQ")
	return &RabbitMQ{conn: conn, channel: ch}, nil
}

// PublishEmailTask отправляет ID письма в очередь
func (r *RabbitMQ) Publish(ctx context.Context, queueName string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	err = r.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key (имя очереди)
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("ошибка публикации в %s: %w", queueName, err)
	}
	return nil
}

func (r *RabbitMQ) Consume(ctx context.Context, queueName string, handler func(body []byte) error) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer tag
		false,     // auto-ack (false = ручной ACK)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("ошибка запуска consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("канал RabbitMQ закрыт")
			}

			err := handler(msg.Body)
			if err != nil {
				slog.Error("Ошибка обработки сообщения", "queue", queueName, "error", err)
				msg.Nack(false, true) // Вернуть в очередь
			} else {
				msg.Ack(false)
			}
		}
	}
}

// DeclareQueue создает очередь, если ее нет (вызывается для каждой очереди отдельно)
func (r *RabbitMQ) DeclareQueue(queueName string) error {
	_, err := r.channel.QueueDeclare(
		queueName, // name (приходит снаружи!)
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	return err
}

// Close закрывает соединения при выключении сервера
func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}
