package middleware

import (
	"context"
	"net/http"
	"strings"

	"tg_bot_asist/internal/api/auth"
	"tg_bot_asist/internal/logger"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// JWTAuthMiddleware проверяет JWT токен и добавляет user_id в контекст.
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Формат: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := auth.ValidateJWT(token)
		if err != nil {
			logger.Warn("JWT validation failed: " + err.Error())
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Добавляем user_id в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID извлекает user_id из контекста.
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
