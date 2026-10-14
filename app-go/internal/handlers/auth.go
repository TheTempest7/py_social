package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/services"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// Register — регистрация нового пользователя
// @Summary      Регистрация
// @Description  Создаёт нового пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Данные регистрации"
// @Success      201 {object} dto.UserResponse
// @Failure      400 {string} string "Невалидный JSON"
// @Failure      409 {string} string "Email или username заняты"
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Читаем JSON
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	// 2. Вызываем сервис
	user, err := h.userService.RegisterUser(req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailTaken):
			respondError(w, http.StatusConflict, "email уже занят")
		case errors.Is(err, services.ErrUsernameTaken):
			respondError(w, http.StatusConflict, "username уже занят")
		default:
			respondError(w, http.StatusInternalServerError, "ошибка регистрации")
		}
		return
	}

	// 3. Отвечаем
	respondJSON(w, http.StatusCreated, user)
}

// Login — вход
// @Summary      Вход
// @Description  Авторизация пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Данные логина"
// @Success      200 {object} map[string]string
// @Failure      401 {string} string "Неверный логин или пароль"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Читаем JSON
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	// 2. Вызываем сервис
	user, token, err := h.userService.LoginUser(req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCreds) {
			respondError(w, http.StatusUnauthorized, "неверный логин или пароль")
			return
		}
		respondError(w, http.StatusInternalServerError, "ошибка логина")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 часа
		HttpOnly: true,         // JS не прочитает
		Secure:   false,        // true в проде (HTTPS)
		SameSite: http.SameSiteLaxMode,
	})

	// 3. Пока отвечаем просто "ок" (JWT добавим позже)
	respondJSON(w, http.StatusOK, dto.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	})
}

// Logout — выход
// @Summary      Выход
// @Description  Удаляет cookie с токеном
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /api/auth/logout [post]
// Logout — удаляет Cookie с токеном
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // 👈 -1 = удалить cookie
		HttpOnly: true,
	})
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "вы вышли",
	})
}

// respondJSON — пишет JSON-ответ
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError — пишет JSON-ошибку
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
