package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/TheTempest7/social_network/app-go/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

// RequireAuth — проверяет JWT из Cookie ИЛИ Authorization header
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// 1. Пробуем Cookie
		if cookie, err := r.Cookie("access_token"); err == nil {
			tokenString = cookie.Value
		}

		// 2. Если нет — пробуем Authorization: Bearer
		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// 3. Нет токена — 401
		if tokenString == "" {
			http.Error(w, `{"error":"не авторизован"}`, http.StatusUnauthorized)
			return
		}

		// 4. Проверяем JWT
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"невалидный токен"}`, http.StatusUnauthorized)
			return
		}

		// 5. Кладём claims в context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
