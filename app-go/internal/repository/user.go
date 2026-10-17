package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/TheTempest7/social_network/app-go/internal/models"
)

// Ошибки репозитория
var (
	ErrUserNotFound  = errors.New("пользователь не найден")
	ErrEmailTaken    = errors.New("email уже занят")
	ErrUsernameTaken = errors.New("username уже занят")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) (*models.User, error) {
	query := `
        INSERT INTO users (email, username, password_hash, first_name, last_name, is_active)
        VALUES (?, ?, ?, ?, ?, ?)
    `

	result, err := r.db.Exec(query,
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

	return r.GetUserByID(user.ID)
}

func (r *UserRepository) GetUserByID(id int) (*models.User, error) {
	query := `
        SELECT id, email, username, password_hash, first_name, last_name, is_active, created_at, updated_at
        FROM users
        WHERE id = ?
    `

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}

	return &user, nil
}

// GetUserByEmail — ищет по email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
        SELECT id, email, username, password_hash, first_name, last_name, is_active, created_at, updated_at
        FROM users
        WHERE email = ?
    `

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
        SELECT id, email, username, password_hash, first_name, last_name, is_active, created_at, updated_at
        FROM users
        WHERE username = ?
    `

	var user models.User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UserExists(email, username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = ? OR username = ?`

	var count int
	if err := r.db.QueryRow(query, email, username).Scan(&count); err != nil {
		return false, fmt.Errorf("проверка существования: %w", err)
	}

	return count > 0, nil
}
