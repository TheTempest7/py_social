package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/TheTempest7/social_network/app-go/internal/models"
)

const queryColumns = "id, email, username, password_hash, first_name, last_name, is_active, created_at, updated_at"

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
		&user.FirstName, &user.LastName, &user.IsActive,
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

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
        INSERT INTO users (email, username, password_hash, first_name, last_name, is_active)
        VALUES (?, ?, ?, ?, ?, ?)
    `

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.IsActive,
	)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "UNIQUE constraint failed") {
			if strings.Contains(errStr, "users.email") {
				return nil, ErrEmailTaken
			}
			if strings.Contains(errStr, "users.username") {
				return nil, ErrUsernameTaken
			}
		}
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("получение ID: %w", err)
	}

	user.ID = int(id)

	return r.GetUserByID(ctx, user.ID)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `
        SELECT ` + queryColumns + `
        FROM users
        WHERE id = ?
    `

	row := r.db.QueryRowContext(ctx, query, id)

	return scanUser(row)
}

// GetUserByEmail — ищет по email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT ` + queryColumns + `
        FROM users
        WHERE email = ?
    `

	row := r.db.QueryRowContext(ctx, query, email)

	return scanUser(row)
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
        SELECT ` + queryColumns + `
        FROM users
        WHERE username = ?
    `

	row := r.db.QueryRowContext(ctx, query, username)

	return scanUser(row)
}

func (r *UserRepository) UserExists(ctx context.Context, email, username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = ? OR username = ?`

	var count int
	if err := r.db.QueryRowContext(ctx, query, email, username).Scan(&count); err != nil {
		return false, fmt.Errorf("проверка существования: %w", err)
	}

	return count > 0, nil
}
