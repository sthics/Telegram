package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// Handler manages a WebSocket connection
type Handler struct {
	conn      *websocket.Conn
	userID    int64
	device    string
	send      chan []byte
	logger    zerolog.Logger
	mu        sync.Mutex
	pingTimer *time.Timer
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewHandler creates a new WebSocket handler
func NewHandler(conn *websocket.Conn, userID int64, device string, logger zerolog.Logger) *Handler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Handler{
		conn:   conn,
		userID: userID,
		device: device,
		send:   make(chan []byte, 256),
		logger: logger.With().
			Int64("user_id", userID).
			Str("device", device).
			Logger(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// ReadPump reads messages from the WebSocket connection
func (h *Handler) ReadPump(onMessage func([]byte) error) {
	defer func() {
		h.cancel()
		h.conn.Close()
	}()

	h.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	h.conn.SetPongHandler(func(string) error {
		h.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, message, err := h.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.logger.Error().Err(err).Msg("websocket read error")
			}
			break
		}

		if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
			if err := onMessage(message); err != nil {
				h.logger.Error().Err(err).Msg("failed to handle message")
			}
		}
	}
}

// WritePump sends messages to the WebSocket connection
func (h *Handler) WritePump(pingInterval time.Duration) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		h.conn.Close()
	}()

	for {
		select {
		case message, ok := <-h.send:
			h.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				h.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := h.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				h.logger.Error().Err(err).Msg("failed to write message")
				return
			}

		case <-ticker.C:
			h.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := h.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Error().Err(err).Msg("failed to write ping")
				return
			}

		case <-h.ctx.Done():
			return
		}
	}
}

// Send queues a message to be sent to the client
func (h *Handler) Send(message []byte) error {
	select {
	case h.send <- message:
		return nil
	case <-h.ctx.Done():
		return fmt.Errorf("connection closed")
	default:
		return fmt.Errorf("send buffer full")
	}
}

// SendJSON sends a JSON message to the client
func (h *Handler) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return h.Send(data)
}

// Close closes the WebSocket connection
func (h *Handler) Close() error {
	h.cancel()
	close(h.send)
	return h.conn.Close()
}

// UserID returns the user ID
func (h *Handler) UserID() int64 {
	return h.userID
}

// Device returns the device ID
func (h *Handler) Device() string {
	return h.device
}

// Context returns the handler context
func (h *Handler) Context() context.Context {
	return h.ctx
}
