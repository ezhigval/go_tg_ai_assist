package middleware

import (
	"net/http"
	"time"

	"tg_bot_asist/internal/logger"
)

// LoggingMiddleware логирует все HTTP запросы.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Обёртка для ResponseWriter для получения статуса
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logger.Info(
			"HTTP " + r.Method + " " + r.URL.Path +
				" - " + http.StatusText(wrapped.statusCode) +
				" - " + duration.String(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
