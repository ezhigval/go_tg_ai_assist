package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"tg_bot_asist/internal/api/middleware"
	"tg_bot_asist/internal/api/websocket"
	"tg_bot_asist/internal/credits"
	"tg_bot_asist/internal/logger"
)

type CreditsHandler struct {
	service *credits.Service
	hub     *websocket.Hub
}

func NewCreditsHandler(service *credits.Service, hub *websocket.Hub) *CreditsHandler {
	return &CreditsHandler{service: service, hub: hub}
}

// List возвращает список кредитов пользователя.
func (h *CreditsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	list, err := h.service.List(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to list credits: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type AddCreditRequest struct {
	Title     string  `json:"title"`
	Principal float64 `json:"principal"`
	Rate      float64 `json:"rate"`
	Months    int     `json:"months"`
}

// Add создаёт новый кредит.
func (h *CreditsHandler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	credit := &credits.Credit{
		UserID:    userID,
		Title:     req.Title,
		Principal: req.Principal,
		Rate:      req.Rate,
		Months:    req.Months,
	}

	id, err := h.service.Add(r.Context(), credit)
	if err != nil {
		logger.Error("Failed to add credit: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "status": "ok"})
}

// Close удаляет кредит.
func (h *CreditsHandler) Close(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.Delete(r.Context(), req.ID, userID)
	if err != nil {
		logger.Error("Failed to close credit: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Schedule возвращает график платежей по кредиту.
func (h *CreditsHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	creditIDStr := r.URL.Query().Get("id")
	if creditIDStr == "" {
		http.Error(w, "Credit ID required", http.StatusBadRequest)
		return
	}

	var creditID int
	if _, err := fmt.Sscanf(creditIDStr, "%d", &creditID); err != nil {
		http.Error(w, "Invalid credit ID", http.StatusBadRequest)
		return
	}

	// Получаем кредит
	credit, err := h.service.GetByID(r.Context(), creditID, userID)
	if err != nil {
		logger.Error("Failed to get credit: " + err.Error())
		http.Error(w, "Credit not found", http.StatusNotFound)
		return
	}

	// Вычисляем график (упрощённо, используем функцию из credit_impl.go)
	// В реальном проекте вынесите это в сервис
	schedule := []map[string]interface{}{} // TODO: использовать calculator

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"credit":   credit,
		"schedule": schedule,
	})
}
