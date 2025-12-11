package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"tg_bot_asist/internal/api/middleware"
	"tg_bot_asist/internal/api/websocket"
	"tg_bot_asist/internal/logger"
	"tg_bot_asist/internal/todo"
)

type TodoHandler struct {
	service *todo.Service
	hub     *websocket.Hub
}

func NewTodoHandler(service *todo.Service, hub *websocket.Hub) *TodoHandler {
	return &TodoHandler{service: service, hub: hub}
}

// List возвращает список задач пользователя.
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.service.List(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to list todos: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

type AddTodoRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	DueDate     *string `json:"due_date,omitempty"`
}

// Add создаёт новую задачу.
func (h *TodoHandler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			dueDate = &parsed
		}
	}

	err := h.service.Add(r.Context(), userID, req.Title, req.Description, dueDate)
	if err != nil {
		logger.Error("Failed to add todo: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Отправляем событие через WebSocket
	if h.hub != nil {
		h.hub.Broadcast(websocket.Event{
			Type:      "todo_added",
			UserID:    userID,
			Data:      map[string]interface{}{"title": req.Title},
			Timestamp: time.Now().Unix(),
		})
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Delete удаляет задачу.
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	err := h.service.Delete(r.Context(), userID, req.ID)
	if err != nil {
		logger.Error("Failed to delete todo: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
