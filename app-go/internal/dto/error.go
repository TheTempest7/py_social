package dto

// ErrorResponse — стандартный ответ при ошибке
type ErrorResponse struct {
	Error string `json:"error" example:"текст ошибки"`
}
