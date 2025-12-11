package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"tg_bot_asist/internal/logger"
)

// VerifyTelegramInitData проверяет подлинность initData от Telegram WebApp.
// Использует HMAC-SHA256 для проверки подписи.
func VerifyTelegramInitData(initData string, botToken string) (map[string]string, error) {
	// Парсим initData
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("invalid initData format: %w", err)
	}

	// Извлекаем hash
	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("hash not found in initData")
	}

	// Удаляем hash из значений для проверки
	values.Del("hash")

	// Сортируем ключи и формируем data-check-string
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheckParts []string
	for _, k := range keys {
		dataCheckParts = append(dataCheckParts, fmt.Sprintf("%s=%s", k, values.Get(k)))
	}
	dataCheckString := strings.Join(dataCheckParts, "\n")

	// Вычисляем секретный ключ
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secretKeyBytes := secretKey.Sum(nil)

	// Вычисляем HMAC
	h := hmac.New(sha256.New, secretKeyBytes)
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	// Проверяем подпись
	if calculatedHash != hash {
		logger.Warn("Telegram initData hash mismatch")
		return nil, fmt.Errorf("invalid hash")
	}

	// Проверяем время (auth_date не должен быть старше 24 часов)
	authDateStr := values.Get("auth_date")
	if authDateStr != "" {
		var authDate int64
		if _, err := fmt.Sscanf(authDateStr, "%d", &authDate); err == nil {
			authTime := time.Unix(authDate, 0)
			if time.Since(authTime) > 24*time.Hour {
				return nil, fmt.Errorf("auth_date expired")
			}
		}
	}

	// Возвращаем данные пользователя
	result := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}

	return result, nil
}

// ExtractUserFromInitData извлекает user_id из проверенных данных.
func ExtractUserFromInitData(data map[string]string) (int64, error) {
	userStr := data["user"]
	if userStr == "" {
		return 0, fmt.Errorf("user not found in initData")
	}

	// Простой парсинг JSON (можно улучшить с помощью json.Unmarshal)
	// Ищем "id":число в строке user
	var userID int64
	if _, err := fmt.Sscanf(userStr, `{"id":%d`, &userID); err != nil {
		return 0, fmt.Errorf("failed to parse user ID: %w", err)
	}

	return userID, nil
}
