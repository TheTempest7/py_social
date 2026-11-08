package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/TheTempest7/social_network/app-go/internal/models"
)

const queryColumns = "id, email, username, password_hash, first_name, last_name, avatar_url, is_active, created_at, updated_at"

// Ошибки репозитория
var (
	ErrUserNotFound  = errors.New("пользователь не найден")
	ErrEmailTaken    = errors.New("email уже занят")
	ErrUsernameTaken = errors.New("username уже занят")
)

type UserRepository struct {
	db *sql.DB
}

func scanUser(row *sql.Row) (*models.User, error) {
	var user models.User

	// 👇 Вызываем Scan на объекте *sql.Row
	err := row.Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.AvatarURL, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)

	// 👇 Вот тут твои ошибки
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}

	return &user, nil
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpdateProfile(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		UPDATE users 
		SET first_name = $1, last_name = $2, username = $3, updated_at = NOW() 
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, user.FirstName, user.LastName, user.Username, user.ID)
	if err != nil {
		return nil, ParseDBError(err)
	}

	// Возвращаем обновленного пользователя
	return r.GetUserByID(ctx, user.ID)
}

func (r *UserRepository) UpdateAvatar(ctx context.Context, id int, url string) (*models.User, error) {
	query := `UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2`
	res, err := r.db.ExecContext(ctx, query, url, id)

	if err != nil {
		return nil, fmt.Errorf("обновление аватара: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return nil, ErrUserNotFound
	}
	return r.GetUserByID(ctx, id)
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
        INSERT INTO users (email, username, password_hash, first_name, last_name, is_active)
        VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
    `

	var id int

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.IsActive,
	).Scan(&id)

	if err != nil {
		return nil, ParseDBError(err)
	}

	user.ID = int(id)

	return r.GetUserByID(ctx, user.ID)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `
        SELECT ` + queryColumns + `
        FROM users
        WHERE id = $1
    `

	row := r.db.QueryRowContext(ctx, query, id)

	return scanUser(row)
}

// GetUserByEmail — ищет по email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT ` + queryColumns + `
        FROM users
        WHERE email = $1
    `

	row := r.db.QueryRowContext(ctx, query, email)

	return scanUser(row)
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
        SELECT ` + queryColumns + `
        FROM users
        WHERE username = $1
    `

	row := r.db.QueryRowContext(ctx, query, username)

	return scanUser(row)
}

func (r *UserRepository) UserExists(ctx context.Context, email, username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = $1 OR username = $2`

	var count int
	if err := r.db.QueryRowContext(ctx, query, email, username).Scan(&count); err != nil {
		return false, fmt.Errorf("проверка существования: %w", err)
	}

	return count > 0, nil
}

// Outbox

// CreateEmailOutbox — сохраняет письмо в БД (статус pending)
func (r *UserRepository) CreateEmailOutbox(ctx context.Context, userID int, email, subject, body string) (int, error) {
	query := `
		INSERT INTO email_outbox (user_id, email, subject, body, status) 
		VALUES ($1, $2, $3, $4, 'pending') 
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query, userID, email, subject, body).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("создание email_outbox: %w", err)
	}
	return id, nil
}

// GetEmailByID — достает письмо по ID (для воркера)
func (r *UserRepository) GetEmailByID(ctx context.Context, id int) (*models.EmailOutbox, error) {
	query := `SELECT id, user_id, email, subject, body, status, retries FROM email_outbox WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var e models.EmailOutbox
	err := row.Scan(&e.ID, &e.UserID, &e.Email, &e.Subject, &e.Body, &e.Status, &e.Retries)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateEmailStatus — обновляет статус письма после попытки отправки
func (r *UserRepository) UpdateEmailStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE email_outbox SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// IncrementEmailRetry — увеличивает счетчик попыток и ставит статус failed
func (r *UserRepository) IncrementEmailRetry(ctx context.Context, id int) error {
	query := `UPDATE email_outbox SET retries = retries + 1, status = 'failed', updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetPendingEmails — для Cron-воркера (находит упавшие письма для повторной отправки)
func (r *UserRepository) GetPendingEmails(ctx context.Context, limit int) ([]models.EmailOutbox, error) {
	query := `
		SELECT id, user_id, email, subject, body, status, retries 
		FROM email_outbox 
		WHERE (status = 'failed' AND retries < 5) 
		   OR (status = 'pending' AND created_at < NOW() - INTERVAL '5 minutes')
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []models.EmailOutbox
	if rows.Err() == nil {
		for rows.Next() {
			var e models.EmailOutbox
			if err := rows.Scan(&e.ID, &e.UserID, &e.Email, &e.Subject, &e.Body, &e.Status, &e.Retries); err != nil {
				return nil, err
			}
			emails = append(emails, e)
		}
	}
	return emails, nil
}

// Outbox
