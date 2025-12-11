package middleware

import (
	"net/http"

	"tg_bot_asist/internal/logger"
)

// RecoverMiddleware перехватывает panic и возвращает 500 ошибку.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered: " + err.(string))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
