package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/TheTempest7/social_network/app-go/internal/auth"
	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	services_user "github.com/TheTempest7/social_network/app-go/internal/services/user"
	"github.com/TheTempest7/social_network/app-go/internal/utils"
	"github.com/TheTempest7/social_network/app-go/internal/validators"
)

type AuthHandler struct {
	userService *services_user.UserService
	blacklist   *auth.Blacklist
}

func NewAuthHandler(userService *services_user.UserService, blacklist *auth.Blacklist) *AuthHandler {
	return &AuthHandler{userService: userService, blacklist: blacklist}
}

// Register — регистрация нового пользователя
// @Summary      Регистрация
// @Description  Создаёт нового пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Данные регистрации"
// @Success      201 {object} dto.UserResponse
// @Failure      400 {object} dto.ErrorResponse "Невалидный JSON"
// @Failure      409 {object} dto.ErrorResponse "Email или username заняты"
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Читаем JSON
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	if err := validators.Validate(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := services_user.RegisterInput{
		Email:     strings.ToLower(strings.Trim(req.Email, " ")),
		Username:  strings.ToLower(strings.Trim(req.Username, " ")),
		Password:  req.Password,
		FirstName: strings.Trim(req.FirstName, " "),
		LastName:  strings.Trim(req.LastName, " "),
	}

	// 2. Вызываем сервис
	user, err := h.userService.RegisterUser(r.Context(), &input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmailTaken):
			respondError(w, http.StatusConflict, "email уже занят")
		case errors.Is(err, repository.ErrUsernameTaken):
			respondError(w, http.StatusConflict, "username уже занят")
		default:
			respondError(w, http.StatusInternalServerError, "ошибка регистрации")
		}
		return
	}

	// 👇 Конвертируем UserOutput в dto.UserResponse для ответа
	resp := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
	}
	// 3. Отвечаем
	respondJSON(w, http.StatusCreated, resp)
}

// Login — вход
// @Summary      Вход
// @Description  Авторизация пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Данные логина"
// @Success      200 {object} map[string]string
// @Failure      401 {object} dto.ErrorResponse "Неверный логин или пароль"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Читаем JSON
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	if err := validators.Validate(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := services_user.LoginInput{
		Username: req.Username,
		Password: req.Password,
	}

	// 2. Вызываем сервис
	user, token, err := h.userService.LoginUser(r.Context(), input)
	if err != nil {
		if errors.Is(err, services_user.ErrInvalidCreds) {
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
		MaxAge:   int(config.AppConfig.JWTLifetime.Seconds()),
		HttpOnly: true, // JS не прочитает
		Secure:   config.AppConfig.CookieSecure,
		SameSite: utils.ParseSameSite(config.AppConfig.CookieSameSite),
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
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /api/auth/logout [post]
// Logout — удаляет Cookie с токеном
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 👇 1. Достаем токен из куки
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		// 👇 2. Добавляем в Redis (черный список) на 24 часа
		// Время жизни берем из конфига (или ставим 24*time.Hour)
		_ = h.blacklist.Add(r.Context(), cookie.Value, config.AppConfig.JWTLifetime)
	}

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
