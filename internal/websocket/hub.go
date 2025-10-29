package websocket

import (
	"sync"

	"github.com/rs/zerolog"
)

// Hub manages active WebSocket connections
type Hub struct {
	connections map[int64]map[string]*Handler // userID -> device -> handler
	mu          sync.RWMutex
	logger      zerolog.Logger
}

// NewHub creates a new WebSocket hub
func NewHub(logger zerolog.Logger) *Hub {
	return &Hub{
		connections: make(map[int64]map[string]*Handler),
		logger:      logger,
	}
}

// Register adds a connection to the hub
func (h *Hub) Register(handler *Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userID := handler.UserID()
	device := handler.Device()

	if h.connections[userID] == nil {
		h.connections[userID] = make(map[string]*Handler)
	}

	// Close existing connection for same device
	if existing, ok := h.connections[userID][device]; ok {
		existing.Close()
	}

	h.connections[userID][device] = handler
	h.logger.Info().
		Int64("user_id", userID).
		Str("device", device).
		Int("total_connections", h.Count()).
		Msg("connection registered")
}

// Unregister removes a connection from the hub
func (h *Hub) Unregister(userID int64, device string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if devices, ok := h.connections[userID]; ok {
		if handler, ok := devices[device]; ok {
			handler.Close()
			delete(devices, device)

			if len(devices) == 0 {
				delete(h.connections, userID)
			}

			h.logger.Info().
				Int64("user_id", userID).
				Str("device", device).
				Int("total_connections", h.Count()).
				Msg("connection unregistered")
		}
	}
}

// Get retrieves a handler for a user's device
func (h *Hub) Get(userID int64, device string) (*Handler, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if devices, ok := h.connections[userID]; ok {
		handler, ok := devices[device]
		return handler, ok
	}
	return nil, false
}

// GetAllForUser retrieves all handlers for a user
func (h *Hub) GetAllForUser(userID int64) []*Handler {
	h.mu.RLock()
	defer h.mu.RUnlock()

	devices, ok := h.connections[userID]
	if !ok {
		return nil
	}

	handlers := make([]*Handler, 0, len(devices))
	for _, handler := range devices {
		handlers = append(handlers, handler)
	}
	return handlers
}

// Count returns the total number of active connections
func (h *Hub) Count() int {
	count := 0
	for _, devices := range h.connections {
		count += len(devices)
	}
	return count
}

// SendToUser sends a message to all devices of a user
func (h *Hub) SendToUser(userID int64, message []byte) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	devices, ok := h.connections[userID]
	if !ok {
		return 0
	}

	sent := 0
	for _, handler := range devices {
		if err := handler.Send(message); err == nil {
			sent++
		}
	}

	return sent
}

// Broadcast sends a message to multiple users
func (h *Hub) Broadcast(userIDs []int64, message []byte) int {
	sent := 0
	for _, userID := range userIDs {
		sent += h.SendToUser(userID, message)
	}
	return sent
}

// GetConnectedUserIDs returns all connected user IDs
func (h *Hub) GetConnectedUserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]int64, 0, len(h.connections))
	for userID := range h.connections {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
