package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"tg_bot_asist/internal/config"
)

var jwtSecret []byte

// InitJWT инициализирует секретный ключ для JWT.
func InitJWT() {
	// Используем значение по умолчанию без предупреждения для development
	secret := config.GetWithDefault("JWT_SECRET", "default-secret-key-change-in-production")
	jwtSecret = []byte(secret)
}

// Claims представляет JWT claims.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT создаёт JWT токен для пользователя.
func GenerateJWT(userID int64) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 дней

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT проверяет и парсит JWT токен.
func ValidateJWT(tokenString string) (*Claims, error) {
	if jwtSecret == nil {
		InitJWT()
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
