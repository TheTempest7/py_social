package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/TheTempest7/social_network/app-go/internal/auth"
	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/middleware"
	"github.com/TheTempest7/social_network/app-go/internal/services/profile"
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
	}

	user, err := h.profileService.UpdateProfile(r.Context(), claims.UserID, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ошибка обновления профиля")
		return
	}

	respondJSON(w, http.StatusOK, user)

}
