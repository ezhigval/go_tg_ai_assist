package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"tg_bot_asist/internal/api/middleware"
	"tg_bot_asist/internal/api/websocket"
	"tg_bot_asist/internal/finance"
	"tg_bot_asist/internal/logger"
)

type FinanceHandler struct {
	service *finance.Service
	hub     *websocket.Hub
}

func NewFinanceHandler(service *finance.Service, hub *websocket.Hub) *FinanceHandler {
	return &FinanceHandler{service: service, hub: hub}
}

// List возвращает список финансовых операций.
func (h *FinanceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	entries, err := h.service.ListEntries(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to list finance entries: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

type AddFinanceRequest struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Type     string  `json:"type"` // "income" or "expense"
	Note     string  `json:"note"`
}

// Add создаёт новую финансовую операцию.
func (h *FinanceHandler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddFinanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Type != "income" && req.Type != "expense" {
		http.Error(w, "Type must be 'income' or 'expense'", http.StatusBadRequest)
		return
	}

	entry := &finance.FinanceEntry{
		UserID:    userID,
		Amount:    req.Amount,
		Category:  req.Category,
		Type:      req.Type,
		Note:      req.Note,
		CreatedAt: time.Now(),
	}

	err := h.service.AddEntry(r.Context(), entry)
	if err != nil {
		logger.Error("Failed to add finance entry: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Stats возвращает статистику по финансам.
func (h *FinanceHandler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	entries, err := h.service.ListEntries(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to get finance stats: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var totalIncome, totalExpense float64
	for _, e := range entries {
		if e.Type == "income" {
			totalIncome += e.Amount
		} else {
			totalExpense += e.Amount
		}
	}

	stats := map[string]interface{}{
		"total_income":  totalIncome,
		"total_expense": totalExpense,
		"balance":       totalIncome - totalExpense,
		"transactions":  len(entries),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

type AddRecurringRequest struct {
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Period      string    `json:"period"` // "daily", "weekly", "monthly", "yearly"
	NextPayment time.Time `json:"next_payment"`
}

// AddRecurring создаёт новый регулярный платёж.
func (h *FinanceHandler) AddRecurring(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddRecurringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payment := &finance.RecurringPayment{
		UserID:      userID,
		Title:       req.Title,
		Amount:      req.Amount,
		Category:    req.Category,
		Period:      req.Period,
		NextPayment: req.NextPayment,
		CreatedAt:   time.Now(),
	}

	id, err := h.service.AddRecurring(r.Context(), payment)
	if err != nil {
		logger.Error("Failed to add recurring payment: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "status": "ok"})
}
