package api

import (
	"net/http"
	"time"

	"tg_bot_asist/internal/api/auth"
	"tg_bot_asist/internal/api/handlers"
	"tg_bot_asist/internal/api/middleware"
	"tg_bot_asist/internal/api/websocket"
	"tg_bot_asist/internal/credits"
	"tg_bot_asist/internal/finance"
	"tg_bot_asist/internal/logger"
	"tg_bot_asist/internal/storage"
	"tg_bot_asist/internal/todo"

	ws "golang.org/x/net/websocket"
)

type Router struct {
	authHandler    *handlers.AuthHandler
	todoHandler    *handlers.TodoHandler
	financeHandler *handlers.FinanceHandler
	creditsHandler *handlers.CreditsHandler
	hub            *websocket.Hub
}

// NewRouter создаёт новый роутер API.
func NewRouter(
	userRepo *storage.UserRepo,
	todoService *todo.Service,
	financeService *finance.Service,
	creditsService *credits.Service,
	hub *websocket.Hub,
) *Router {
	return &Router{
		authHandler:    handlers.NewAuthHandler(userRepo),
		todoHandler:    handlers.NewTodoHandler(todoService, hub),
		financeHandler: handlers.NewFinanceHandler(financeService, hub),
		creditsHandler: handlers.NewCreditsHandler(creditsService, hub),
		hub:            hub,
	}
}

// SetupRoutes настраивает все маршруты API.
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Middleware chain
	chain := middleware.RecoverMiddleware(
		middleware.LoggingMiddleware(
			middleware.CORSMiddleware(
				middleware.TimeoutMiddleware(30 * time.Second)(
					mux,
				),
			),
		),
	)

	// Auth routes (без JWT)
	mux.HandleFunc("/api/auth", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			r.authHandler.Auth(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.Handle("/api/user/me", middleware.JWTAuthMiddleware(http.HandlerFunc(r.authHandler.Me)))

	// TODO routes
	mux.Handle("/api/todo/list", middleware.JWTAuthMiddleware(http.HandlerFunc(r.todoHandler.List)))
	mux.Handle("/api/todo/add", middleware.JWTAuthMiddleware(http.HandlerFunc(r.todoHandler.Add)))
	mux.Handle("/api/todo/delete", middleware.JWTAuthMiddleware(http.HandlerFunc(r.todoHandler.Delete)))

	// Finance routes
	mux.Handle("/api/finance/list", middleware.JWTAuthMiddleware(http.HandlerFunc(r.financeHandler.List)))
	mux.Handle("/api/finance/add", middleware.JWTAuthMiddleware(http.HandlerFunc(r.financeHandler.Add)))
	mux.Handle("/api/finance/stats", middleware.JWTAuthMiddleware(http.HandlerFunc(r.financeHandler.Stats)))
	mux.Handle("/api/finance/recurring/add", middleware.JWTAuthMiddleware(http.HandlerFunc(r.financeHandler.AddRecurring)))

	// Credits routes
	mux.Handle("/api/credits/list", middleware.JWTAuthMiddleware(http.HandlerFunc(r.creditsHandler.List)))
	mux.Handle("/api/credits/add", middleware.JWTAuthMiddleware(http.HandlerFunc(r.creditsHandler.Add)))
	mux.Handle("/api/credits/close", middleware.JWTAuthMiddleware(http.HandlerFunc(r.creditsHandler.Close)))
	mux.Handle("/api/credits/schedule", middleware.JWTAuthMiddleware(http.HandlerFunc(r.creditsHandler.Schedule)))

	// WebSocket
	mux.HandleFunc("/ws", r.handleWebSocket)

	return chain
}

// handleWebSocket обрабатывает WebSocket соединения.
func (r *Router) handleWebSocket(w http.ResponseWriter, req *http.Request) {
	// Извлекаем токен из query параметров
	token := req.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token required", http.StatusUnauthorized)
		return
	}

	// Проверяем JWT токен
	claims, err := auth.ValidateJWT(token)
	if err != nil {
		logger.Warn("WebSocket JWT validation failed: " + err.Error())
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	ws.Handler(func(conn *ws.Conn) {
		r.hub.HandleWebSocket(conn, claims.UserID)
	}).ServeHTTP(w, req)
}
