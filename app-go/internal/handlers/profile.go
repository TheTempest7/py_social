package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/TheTempest7/social_network/app-go/internal/auth"
	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/middleware"
	"github.com/TheTempest7/social_network/app-go/internal/services/profile"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	profileService *profile.ProfileService
}

func NewProfileHandler(profileService *profile.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

// Profile — текущий пользователь
// @Summary      Текущий пользователь
// @Description  Возвращает данные авторизованного пользователя
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.UserResponse
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Router       /api/auth/profile [get]
func (h *ProfileHandler) Profile(w http.ResponseWriter, r *http.Request) {
	// Достаём claims из context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		respondError(w, http.StatusUnauthorized, "не авторизован")
		return
	}

	// Ищем пользователя в БД
	user, err := h.profileService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "пользователь не найден")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *ProfileHandler) PutProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)

	if !ok {
		respondError(w, http.StatusUnauthorized, "не авторизован")
		return
	}

	var req dto.UpdateProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	input := profile.UpdateProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
	}

	user, err := h.profileService.UpdateProfile(r.Context(), claims.UserID, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ошибка обновления профиля")
		return
	}

	respondJSON(w, http.StatusOK, user)

}

// UploadAvatar PUT /api/users/avatar
// @Summary      Загрузить аватар
// @Description  Принимает multipart/form-data с полем "avatar" (image/*). Лимит 5MB.
// @Tags         profile
// @Security     BearerAuth
// @Accept       mpfd
// @Produce      json
// @Param        avatar formData file true "Файл аватара"
// @Success      200 {object} dto.UserOutput
// @Failure      400 {string} string "bad request"
// @Failure      401 {string} string "unauthorized"
// @Failure      413 {string} string "file too large"
// @Router       /api/users/avatar [put]
func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	const maxAvatar = 5 << 20 // 5 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatar)

	// ParseMultipartForm с memory=1MB, остальное уйдёт в temp-файлы
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		// MaxBytesReader возвращает *http.MaxBytesError
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			respondError(w, http.StatusRequestEntityTooLarge, "file too large (max 5MB)")
			return
		}
		respondError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		respondError(w, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	// Проверка MIME
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		respondError(w, http.StatusBadRequest, "avatar must be an image")
		return
	}

	// Разрешённые расширения
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		// ok
	default:
		respondError(w, http.StatusBadRequest, "unsupported image format")
		return
	}

	objectName := fmt.Sprintf("avatars/%d/%s%s", claims.UserID, uuid.NewString(), ext)
	// 8. Грузим в MinIO
	url, err := h.profileService.UploadAvatar(r.Context(), objectName, file, header.Size)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось загрузить файл")
		return
	}

	// 9. Сохраняем URL в БД и возвращаем обновлённый DTO
	out, err := h.profileService.SetAvatar(r.Context(), claims.UserID, url)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось сохранить аватар")
		return
	}

	respondJSON(w, http.StatusOK, out)
}
