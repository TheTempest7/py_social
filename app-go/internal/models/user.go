package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // 👈 не выводим в JSON
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	AvatarURL    *string
}

type EmailOutbox struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	Retries   int       `json:"retries"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
