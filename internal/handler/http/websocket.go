package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ambarg/mini-telegram/internal/auth"
	"github.com/ambarg/mini-telegram/internal/domain"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/service/chat"
	ws "github.com/ambarg/mini-telegram/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type WebSocketHandler struct {
	hub       *ws.Hub
	chatSvc   *chat.Service
	authSvc   *auth.Service
	rmqClient *rabbitmq.Client
	queueName string // Gateway's delivery queue name
}

func NewWebSocketHandler(hub *ws.Hub, chatSvc *chat.Service, authSvc *auth.Service, rmqClient *rabbitmq.Client, queueName string) *WebSocketHandler {
	return &WebSocketHandler{
		hub:       hub,
		chatSvc:   chatSvc,
		authSvc:   authSvc,
		rmqClient: rmqClient,
		queueName: queueName,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WebSocketHandler) HandleWS(c *gin.Context) {
	// 1. Authenticate
	// Try to get token from query param or header
	token := c.Query("token")
	if token == "" {
		// Try Authorization header
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := h.authSvc.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	userID, err := auth.ExtractUserID(claims)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token subject"})
		return
	}

	// 2. Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade websocket")
		return
	}

	// 3. Create Handler
	// Use a default device ID for now, or extract from query/header
	device := c.Query("device")
	if device == "" {
		device = "web"
	}

	wsHandler := ws.NewHandler(conn, userID, device, log.Logger)
	h.hub.Register(wsHandler)

	// 4. Subscribe to user's chats
	// We need to get user's chats and bind the gateway queue to them
	// This should ideally be done async or optimized, but for now do it here
	ctx := c.Request.Context()
	chats, err := h.chatSvc.GetUserChats(ctx, userID)
	if err == nil {
		for _, chat := range chats {
			h.hub.Subscribe(userID, chat.ID)
			// Bind gateway queue to this chat
			// Note: This might be redundant if already bound, but RabbitMQ handles idempotency
			if err := h.rmqClient.BindDeliveryQueue(h.queueName, chat.ID); err != nil {
				log.Error().Err(err).Int64("chat_id", chat.ID).Msg("failed to bind delivery queue")
			}
		}
	}

	// 5. Start Pumps
	go wsHandler.WritePump(50 * time.Second)
	go wsHandler.ReadPump(func(msg []byte) error {
		return h.handleMessage(userID, msg)
	})
}

func (h *WebSocketHandler) handleMessage(userID int64, payload []byte) error {
	var msg map[string]any
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}

	// Inject UserID if missing
	msg["userId"] = userID
	// Re-marshal payload
	newPayload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	msgType, _ := msg["type"].(string)
	ctx := context.Background()

	switch msgType {
	case "SendMessage":
		chatID, _ := msg["chatId"].(float64)
		body, _ := msg["body"].(string)
		uuid, _ := msg["uuid"].(string)

		domainMsg := &domain.Message{
			ChatID:    int64(chatID),
			UserID:    userID,
			Body:      body,
			CreatedAt: time.Now(),
		}

		return h.chatSvc.ProcessMessage(ctx, domainMsg, uuid)

	case "Typing":
		chatID, _ := msg["chatId"].(float64)
		// Publish typing event
		return h.rmqClient.PublishTypingEvent(ctx, int64(chatID), newPayload)

	case "Read":
		// Publish read receipt
		return h.rmqClient.PublishReadReceipt(ctx, newPayload)
	}

	return nil
}
