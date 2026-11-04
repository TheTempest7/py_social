package repository

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	// Проверяем, является ли ошибка ошибкой Postgres
	if errors.As(err, &pgErr) {
		// 23505 — это код unique_violation
		if pgErr.Code == "23505" {
			// pgErr.ConstraintName содержит имя индекса, например "users_email_key"
			if strings.Contains(pgErr.ConstraintName, "email") {
				return ErrEmailTaken
			}
			if strings.Contains(pgErr.ConstraintName, "username") {
				return ErrUsernameTaken
			}
		}
	}

	// Если это не известная нам ошибка, возвращаем её как есть
	return err
}
