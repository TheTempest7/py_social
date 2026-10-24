package validators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Один экземпляр на всё приложение — валидатор потокобезопасен
var v = validator.New()

// Validate — проверяет структуру по тегам `validate`
func Validate(s any) error {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	// Достаём список ошибок по полям
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		messages := make([]string, 0, len(errs))
		for _, e := range errs {
			messages = append(messages, formatError(e))
		}
		return fmt.Errorf("ошибка валидации: %s", strings.Join(messages, "; "))
	}

	return err
}

// formatError — читаемое сообщение для одного поля
func formatError(e validator.FieldError) string {
	field := e.Field()
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("поле %s обязательно", field)
	case "email":
		return fmt.Sprintf("поле %s должно быть валидным email", field)
	case "min":
		return fmt.Sprintf("поле %s должно быть не меньше %s", field, e.Param())
	case "max":
		return fmt.Sprintf("поле %s должно быть не больше %s", field, e.Param())
	default:
		return fmt.Sprintf("поле %s невалидно", field)
	}
}
