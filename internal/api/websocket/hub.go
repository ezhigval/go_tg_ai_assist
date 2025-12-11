package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/websocket"
	
)

// Event представляет событие синхронизации.
type Event struct {
	Type      string      `json:"type"` // "todo_added", "finance_added", "credit_added", etc.
	UserID    int64       `json:"user_id"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// NewEvent создаёт новое событие.
func NewEvent(eventType string, userID int64, data interface{}) Event {
	return Event{
		Type:      eventType,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// Client представляет WebSocket клиента.
type Client struct {
	ID   int64
	Conn *websocket.Conn
	Send chan Event
	Hub  *Hub
}

// Hub управляет всеми WebSocket соединениями.
type Hub struct {
	clients    map[int64]*Client
	broadcast  chan Event
	register   chan *Client
	unregister chan *Client
	stop       chan struct{}
	mu         sync.RWMutex
}

// NewHub создаёт новый Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		broadcast:  make(chan Event, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan struct{}),
	}
}

// Run запускает Hub (должен быть запущен в отдельной горутине).
func (h *Hub) Run() {
	for {
		select {
		case <-h.stop:
			logger.Info("WebSocket hub stopping, closing all connections...")
			h.mu.Lock()
			// Закрываем все соединения
			for _, client := range h.clients {
				close(client.Send)
				client.Conn.Close()
			}
			h.clients = make(map[int64]*Client)
			h.mu.Unlock()
			logger.Info("WebSocket hub stopped")
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			logger.Info(fmt.Sprintf("WebSocket client connected: %d", client.ID))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			logger.Info(fmt.Sprintf("WebSocket client disconnected: %d", client.ID))

		case event := <-h.broadcast:
			h.mu.RLock()
			// Отправляем событие только пользователю, которому оно предназначено
			if client, ok := h.clients[event.UserID]; ok {
				select {
				case client.Send <- event:
				default:
					close(client.Send)
					delete(h.clients, event.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Stop останавливает Hub и закрывает все соединения.
func (h *Hub) Stop() {
	select {
	case h.stop <- struct{}{}:
	default:
		// Уже остановлен
	}
}

// Broadcast отправляет событие пользователю.
func (h *Hub) Broadcast(event Event) {
	select {
	case h.broadcast <- event:
	default:
		logger.Warn("Hub broadcast channel full, dropping event")
	}
}

// HandleWebSocket обрабатывает WebSocket соединение.
func (h *Hub) HandleWebSocket(ws *websocket.Conn, userID int64) {
	client := &Client{
		ID:   userID,
		Conn: ws,
		Send: make(chan Event, 256),
		Hub:  h,
	}

	h.register <- client

	// Запускаем горутины для чтения и записи
	go client.writePump()
	go client.readPump()
}

// readPump читает сообщения от клиента.
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg json.RawMessage
		if err := websocket.JSON.Receive(c.Conn, &msg); err != nil {
			break
		}
		// Обработка входящих сообщений (ping/pong, etc.)
	}
}

// writePump отправляет сообщения клиенту.
func (c *Client) writePump() {
	defer c.Conn.Close()

	for {
		select {
		case event, ok := <-c.Send:
			if !ok {
				return
			}

			if err := websocket.JSON.Send(c.Conn, event); err != nil {
				return
			}
		}
	}
}
