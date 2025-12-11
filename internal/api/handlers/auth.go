package handlers

import (
	"encoding/json"
	"net/http"

	"tg_bot_asist/internal/api/auth"
	"tg_bot_asist/internal/api/middleware"
	"tg_bot_asist/internal/config"
	"tg_bot_asist/internal/logger"
	"tg_bot_asist/internal/storage"
)

type AuthHandler struct {
	userRepo *storage.UserRepo
}

func NewAuthHandler(userRepo *storage.UserRepo) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

type AuthRequest struct {
	InitData string `json:"initData"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	UserID int64  `json:"user_id"`
}

// Auth обрабатывает авторизацию через Telegram initData.
func (h *AuthHandler) Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем initData
	botToken := config.GetRequired("TELEGRAM_BOT_API")
	if botToken == "" {
		logger.Error("TELEGRAM_BOT_API not configured")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	data, err := auth.VerifyTelegramInitData(req.InitData, botToken)
	if err != nil {
		logger.Warn("Telegram initData verification failed: " + err.Error())
		http.Error(w, "Invalid initData", http.StatusUnauthorized)
		return
	}

	// Извлекаем user_id
	userID, err := auth.ExtractUserFromInitData(data)
	if err != nil {
		logger.Warn("Failed to extract user ID: " + err.Error())
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	// Регистрируем пользователя (если ещё не зарегистрирован)
	ctx := r.Context()
	if err := h.userRepo.RegisterUser(ctx, userID, userID, ""); err != nil {
		logger.Error("Failed to register user: " + err.Error())
		// Продолжаем даже если регистрация не удалась
	}

	// Генерируем JWT
	token, err := auth.GenerateJWT(userID)
	if err != nil {
		logger.Error("Failed to generate JWT: " + err.Error())
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Возвращаем токен
	response := AuthResponse{
		Token:  token,
		UserID: userID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Me возвращает информацию о текущем пользователе.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"user_id": userID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
