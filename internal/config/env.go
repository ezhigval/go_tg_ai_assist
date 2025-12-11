package config

import (
	"os"

	"tg_bot_asist/internal/logger"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env not found — using system environment variables")
		return err
	}

	logger.Info(".env loaded successfully")
	return nil
}

// Get возвращает значение переменной окружения.
// Если переменная не найдена, возвращает пустую строку без предупреждения.
func Get(key string) string {
	return os.Getenv(key)
}

// GetWithDefault возвращает значение переменной окружения или значение по умолчанию.
// Предупреждение не выводится, если используется значение по умолчанию.
func GetWithDefault(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

// GetRequired возвращает значение переменной окружения или логирует предупреждение.
// Используется для критичных переменных, которые должны быть установлены.
func GetRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		logger.Warn("Missing ENV variable: " + key)
	}
	return v
}
