package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ambarg/mini-telegram/internal/auth"
	"github.com/ambarg/mini-telegram/internal/config"
	"github.com/ambarg/mini-telegram/internal/database"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/redis"
	"github.com/ambarg/mini-telegram/internal/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// GatewayServer handles HTTP and WebSocket requests
type GatewayServer struct {
	cfg         *config.Config
	authService *auth.Service
	db          *database.DB
	redis       *redis.Client
	rabbitmq    *rabbitmq.Client
	hub         *websocket.Hub
	upgrader    gorillaws.Upgrader
	metrics     *Metrics
	deliveryQ   string
}

// Metrics holds Prometheus metrics
type Metrics struct {
	msgSent  *prometheus.CounterVec
	durHist  *prometheus.HistogramVec
	wsConns  prometheus.Gauge
}

// NewGatewayServer creates a new gateway server
func NewGatewayServer(
	cfg *config.Config,
	authService *auth.Service,
	db *database.DB,
	redisClient *redis.Client,
	rmqClient *rabbitmq.Client,
	hub *websocket.Hub,
) *GatewayServer {
	metrics := &Metrics{
		msgSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "gateway_msg_sent_total"},
			[]string{"chat_type"},
		),
		durHist: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_delivery_duration_seconds",
				Buckets: []float64{.01, .05, .1, .2},
			},
			[]string{"status"},
		),
		wsConns: prometheus.NewGauge(
			prometheus.GaugeOpts{Name: "gateway_ws_connections"},
		),
	}

	prometheus.MustRegister(metrics.msgSent, metrics.durHist, metrics.wsConns)

	allowedOrigins := make(map[string]bool)
	for _, origin := range cfg.AllowedOrigins {
		allowedOrigins[origin] = true
	}

	return &GatewayServer{
		cfg:         cfg,
		authService: authService,
		db:          db,
		redis:       redisClient,
		rabbitmq:    rmqClient,
		hub:         hub,
		upgrader: gorillaws.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // Allow non-browser clients? Or strictly enforce? For now, allow empty origin (e.g. mobile apps/curl potentially)
				}
				return allowedOrigins[origin]
			},
		},
		metrics: metrics,
	}
}

// Router creates the Gin router with all routes
func (s *GatewayServer) Router() *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("gateway"))

	// CORS Config
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = s.cfg.AllowedOrigins
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// Health check
	r.GET("/v1/health", s.healthHandler)

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Rate limiter for auth endpoints
	loginRate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  int64(s.cfg.LoginRateLimit),
	}
	loginStore := memory.NewStore()
	loginLimiter := limiter.New(loginStore, loginRate)
	loginMiddleware := ginlimiter.NewMiddleware(loginLimiter)

	// Auth routes
	authGroup := r.Group("/v1/auth")
	{
		authGroup.POST("/register", loginMiddleware, s.registerHandler)
		authGroup.POST("/login", loginMiddleware, s.loginHandler)
		authGroup.POST("/refresh", s.refreshHandler)
	}

	// Protected routes
	protected := r.Group("/v1")
	protected.Use(s.authService.JWTMiddleware())
	{
		// Chat routes
		protected.GET("/chats", s.getChatsHandler)
		protected.POST("/chats", s.createChatHandler)
		protected.POST("/chats/:id/invite", s.inviteToChatHandler)
		protected.DELETE("/chats/:id/members/:userId", s.kickMemberHandler)
		protected.DELETE("/chats/:id/members", s.leaveChatHandler)
		protected.POST("/devices", s.registerDeviceHandler)

		// WebSocket
		wsRate := limiter.Rate{
			Period: 1 * time.Minute,
			Limit:  int64(s.cfg.WSRateLimit),
		}
		wsStore := memory.NewStore()
		wsLimiter := limiter.New(wsStore, wsRate)
		wsMiddleware := ginlimiter.NewMiddleware(wsLimiter)

		protected.GET("/ws", wsMiddleware, s.websocketHandler)
	}

	return r
}

// healthHandler handles health check requests
func (s *GatewayServer) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// registerHandler handles user registration
func (s *GatewayServer) registerHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	// Create user
	user := &database.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.cfg.PostgresTimeout)
	defer cancel()

	if err := s.db.CreateUser(ctx, user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists or database error"})
		return
	}

	// Generate tokens
	accessToken, err := s.authService.GenerateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	refreshToken, err := s.authService.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	// Set refresh token as HttpOnly cookie
	c.SetCookie("refreshToken", refreshToken, int(auth.RefreshTokenLifetime.Seconds()), "/", "", true, true)

	c.JSON(http.StatusCreated, gin.H{
		"userId":       user.ID,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// loginHandler handles user login
func (s *GatewayServer) loginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.cfg.PostgresTimeout)
	defer cancel()

	// Get user by email
	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Verify password
	if err := auth.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, err := s.authService.GenerateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	refreshToken, err := s.authService.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	// Set refresh token as HttpOnly cookie
	c.SetCookie("refreshToken", refreshToken, int(auth.RefreshTokenLifetime.Seconds()), "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"userId":       user.ID,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// refreshHandler handles token refresh
func (s *GatewayServer) refreshHandler(c *gin.Context) {
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	// Validate refresh token
	claims, err := s.authService.ValidateToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	userID, err := auth.ExtractUserID(claims)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	// Generate new tokens
	accessToken, err := s.authService.GenerateAccessToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	newRefreshToken, err := s.authService.GenerateRefreshToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	// Set new refresh token
	c.SetCookie("refreshToken", newRefreshToken, int(auth.RefreshTokenLifetime.Seconds()), "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
	})
}

// getChatsHandler retrieves user's chats
func (s *GatewayServer) getChatsHandler(c *gin.Context) {
	userID, _ := auth.GetUserID(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.cfg.PostgresTimeout)
	defer cancel()

	chats, err := s.db.GetUserChats(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chats"})
		return
	}

	c.JSON(http.StatusOK, chats)
}

// createChatHandler creates a new chat
func (s *GatewayServer) createChatHandler(c *gin.Context) {
	userID, _ := auth.GetUserID(c)

	var req struct {
		Type      int16   `json:"type" binding:"required,oneof=1 2"`
		MemberIDs []int64 `json:"memberIds" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.cfg.PostgresTimeout)
	defer cancel()

	// Create chat
	chat := &database.Chat{Type: req.Type}
	if err := s.db.CreateChat(ctx, chat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create chat"})
		return
	}

	// Add creator as member
	allMembers := append([]int64{userID}, req.MemberIDs...)

	for _, memberID := range allMembers {
		role := database.RoleMember
		if memberID == userID {
			role = database.RoleAdmin
		}
		
		member := &database.ChatMember{
			ChatID: chat.ID,
			UserID: memberID,
			Role:   role,
		}
		if err := s.db.AddChatMember(ctx, member); err != nil {
			log.Error().Err(err).Int64("chat_id", chat.ID).Int64("user_id", memberID).Msg("failed to add chat member")
		}
	}

	// Cache members in Redis
	if err := s.redis.AddGroupMembers(ctx, chat.ID, allMembers); err != nil {
		log.Error().Err(err).Msg("failed to cache group members")
	}

	// Note: No need to declare per-chat queues anymore
	// We use a single shared queue (chat.messages) for all chats
	// This is best practice for scalable messaging systems

	c.JSON(http.StatusCreated, gin.H{"chatId": chat.ID})
}

// inviteToChatHandler invites a user to a chat
func (s *GatewayServer) inviteToChatHandler(c *gin.Context) {
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat ID"})
		return
	}

	var req struct {
		UserID int64 `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.cfg.PostgresTimeout)
	defer cancel()

	// Add member
	member := &database.ChatMember{
		ChatID: chatID,
		UserID: req.UserID,
	}

	if err := s.db.AddChatMember(ctx, member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add member"})
		return
	}

	// Update Redis cache
	if err := s.redis.AddGroupMembers(ctx, chatID, []int64{req.UserID}); err != nil {
		log.Error().Err(err).Msg("failed to update group members cache")
	}

	c.Status(http.StatusNoContent)
}

// kickMemberHandler removes a member from a chat (admin only)
func (s *GatewayServer) kickMemberHandler(c *gin.Context) {
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat ID"})
		return
	}

	targetUserID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	requesterID, _ := auth.GetUserID(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.cfg.PostgresTimeout)
	defer cancel()

	// Check if requester is admin
	requesterMember, err := s.db.GetChatMember(ctx, chatID, requesterID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this chat"})
		return
	}

	if requesterMember.Role != database.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can kick members"})
		return
	}

	// Remove member
	if err := s.db.RemoveChatMember(ctx, chatID, targetUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member"})
		return
	}

	// Update Redis cache (remove from group members)
	// Note: This is a simplification. Ideally we should fetch current members, remove one, and set back.
	// For now, we'll just invalidate the cache or let it expire.
	// A better approach for Redis set would be SREM.
	// Assuming AddGroupMembers uses SADD, we should implement RemoveGroupMember using SREM.
	if err := s.redis.RemoveGroupMember(ctx, chatID, targetUserID); err != nil {
		log.Error().Err(err).Msg("failed to update group members cache")
	}

	c.Status(http.StatusNoContent)
}

// leaveChatHandler allows a user to leave a chat
func (s *GatewayServer) leaveChatHandler(c *gin.Context) {
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat ID"})
		return
	}

	userID, _ := auth.GetUserID(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.cfg.PostgresTimeout)
	defer cancel()

	// Remove self
	if err := s.db.RemoveChatMember(ctx, chatID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to leave chat"})
		return
	}

	// Update Redis cache
	if err := s.redis.RemoveGroupMember(ctx, chatID, userID); err != nil {
		log.Error().Err(err).Msg("failed to update group members cache")
	}

	c.Status(http.StatusNoContent)
}

// registerDeviceHandler registers a device token for push notifications
func (s *GatewayServer) registerDeviceHandler(c *gin.Context) {
	userID, _ := auth.GetUserID(c)

	var req struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform" binding:"required,oneof=ios android web"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.cfg.PostgresTimeout)
	defer cancel()

	token := &database.DeviceToken{
		UserID:   userID,
		Token:    req.Token,
		Platform: req.Platform,
	}

	if err := s.db.AddDeviceToken(ctx, token); err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("failed to register device token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register device"})
		return
	}

	c.Status(http.StatusCreated)
}

// websocketHandler handles WebSocket upgrade and connection
func (s *GatewayServer) websocketHandler(c *gin.Context) {
	userID, _ := auth.GetUserID(c)

	// Get device ID from query param (default to "web")
	device := c.DefaultQuery("device", "web")

	// Upgrade connection
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade connection")
		return
	}
	log.Info().Int64("user_id", userID).Msg("websocket connection upgraded")

	// Create handler
	handler := websocket.NewHandler(conn, userID, device, log.Logger)

	// Register connection
	s.hub.Register(handler)
	s.metrics.wsConns.Set(float64(s.hub.Count()))
	log.Info().Int64("user_id", userID).Msg("websocket handler registered")

	// Subscribe to chats and bind delivery queue
	// This ensures we receive messages for this user's chats
	userChats, err := s.db.GetUserChats(context.Background(), userID)
	if err == nil {
		for _, chat := range userChats {
			s.hub.Subscribe(userID, chat.ID)
			if s.deliveryQ != "" {
				if err := s.rabbitmq.BindDeliveryQueue(s.deliveryQ, chat.ID); err != nil {
					log.Error().Err(err).Int64("chat_id", chat.ID).Msg("failed to bind delivery queue")
				} else {
					log.Info().Int64("chat_id", chat.ID).Str("queue", s.deliveryQ).Msg("bound delivery queue to chat")
				}
			}
		}
	} else {
		log.Error().Err(err).Msg("failed to get user chats for subscription")
	}

	// Register in Redis
	ctx := context.Background()
	podIP := os.Getenv("POD_IP")
	if podIP == "" {
		podIP = "localhost"
	}

	if err := s.redis.RegisterConnection(ctx, userID, device, podIP, s.cfg.ConnTTL); err != nil {
		log.Error().Err(err).Msg("failed to register connection in Redis")
	}

	// Set presence to online
	if err := s.redis.SetPresence(ctx, userID, true, s.cfg.ConnTTL+5*time.Second); err != nil {
		log.Error().Err(err).Msg("failed to set presence")
	}

	// Publish presence event (user online)
	presencePayload, _ := json.Marshal(map[string]any{
		"type":     "Presence",
		"userId":   userID,
		"online":   true,
		"lastSeen": time.Now().Unix(),
	})
	if err := s.rabbitmq.PublishPresenceEvent(ctx, presencePayload); err != nil {
		log.Warn().Err(err).Msg("failed to publish presence event")
	}

	// Start read/write pumps
	go handler.WritePump(s.cfg.PingInterval)
	go handler.ReadPump(func(message []byte) error {
		return s.handleWebSocketMessage(handler, message)
	})

	// Wait for connection to close
	<-handler.Context().Done()

	// Cleanup
	s.hub.Unregister(userID, device)
	// Unsubscribe from chats
	if userChats, err := s.db.GetUserChats(context.Background(), userID); err == nil {
		for _, chat := range userChats {
			s.hub.Unsubscribe(userID, chat.ID)
		}
	}
	s.metrics.wsConns.Set(float64(s.hub.Count()))

	if err := s.redis.UnregisterConnection(ctx, userID, device); err != nil {
		log.Error().Err(err).Msg("failed to unregister connection from Redis")
	}

	if err := s.redis.SetPresence(ctx, userID, false, 0); err != nil {
		log.Error().Err(err).Msg("failed to update presence")
	}

	// Publish presence event (user offline)
	offlinePayload, _ := json.Marshal(map[string]any{
		"type":     "Presence",
		"userId":   userID,
		"online":   false,
		"lastSeen": time.Now().Unix(),
	})
	if err := s.rabbitmq.PublishPresenceEvent(context.Background(), offlinePayload); err != nil {
		log.Warn().Err(err).Msg("failed to publish offline presence event")
	}
}

// handleWebSocketMessage processes incoming WebSocket messages
func (s *GatewayServer) handleWebSocketMessage(handler *websocket.Handler, message []byte) error {
	// Parse message (expecting JSON for now, protobuf can be added later)
	var msg map[string]any
	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return fmt.Errorf("missing message type")
	}

	switch msgType {
	case "SendMessage":
		return s.handleSendMessage(handler, msg)
	case "Read":
		return s.handleReadReceipt(handler, msg)
	case "Typing":
		return s.handleTypingEvent(handler, msg)
	case "Ping":
		return handler.SendJSON(map[string]any{"type": "Pong", "timestamp": time.Now().Unix()})
	default:
		return fmt.Errorf("unknown message type: %s", msgType)
	}
}

// handleSendMessage processes SendMessage events
func (s *GatewayServer) handleSendMessage(handler *websocket.Handler, msg map[string]any) error {
	uuid, _ := msg["uuid"].(string)
	chatID, _ := msg["chatId"].(float64)
	body, _ := msg["body"].(string)

	if uuid == "" || chatID == 0 || body == "" {
		return fmt.Errorf("invalid SendMessage fields")
	}

	// Publish to RabbitMQ
	payload, _ := json.Marshal(map[string]any{
		"uuid":   uuid,
		"chatId": int64(chatID),
		"userId": handler.UserID(),
		"body":   body,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.rabbitmq.PublishToChatQueue(ctx, int64(chatID), payload); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	s.metrics.msgSent.WithLabelValues("message").Inc()
	return nil
}

// handleReadReceipt processes read receipt events
func (s *GatewayServer) handleReadReceipt(handler *websocket.Handler, msg map[string]any) error {
	chatID, _ := msg["chatId"].(float64)
	msgID, _ := msg["msgId"].(float64)

	if chatID == 0 || msgID == 0 {
		return fmt.Errorf("invalid read receipt fields")
	}

	// Publish to read receipt queue for batch processing
	receiptPayload, _ := json.Marshal(map[string]any{
		"chatId": int64(chatID),
		"userId": handler.UserID(),
		"msgId":  int64(msgID),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.rabbitmq.PublishReadReceipt(ctx, receiptPayload); err != nil {
		log.Error().Err(err).Msg("failed to publish read receipt")
		return err
	}

	log.Info().
		Int64("chat_id", int64(chatID)).
		Int64("msg_id", int64(msgID)).
		Int64("user_id", handler.UserID()).
		Msg("read receipt published")

	return nil
}

// handleTypingEvent processes Typing events
func (s *GatewayServer) handleTypingEvent(handler *websocket.Handler, msg map[string]any) error {
	chatID, _ := msg["chatId"].(float64)
	if chatID == 0 {
		return fmt.Errorf("invalid typing event fields")
	}

	// Publish ephemeral event
	payload, _ := json.Marshal(map[string]any{
		"type":   "Typing",
		"chatId": int64(chatID),
		"userId": handler.UserID(),
	})

	if err := s.rabbitmq.PublishTypingEvent(context.Background(), int64(chatID), payload); err != nil {
		return fmt.Errorf("failed to publish typing event: %w", err)
	}

	return nil
}

// StartConsumers starts consuming from RabbitMQ
func (s *GatewayServer) StartConsumers(ctx context.Context) error {
	// Declare a unique delivery queue for this pod
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = fmt.Sprintf("gateway-%d", time.Now().UnixNano())
	}
	
	// We start with no bindings, they will be added as users connect
	queueName, err := s.rabbitmq.DeclareDeliveryQueue(podName, nil)
	if err != nil {
		return fmt.Errorf("failed to declare delivery queue: %w", err)
	}
	s.deliveryQ = queueName
	
	log.Info().Str("queue", queueName).Msg("started delivery consumer")

	msgs, err := s.rabbitmq.ConsumeDeliveryQueue(queueName, podName)
	if err != nil {
		return fmt.Errorf("failed to start consuming delivery queue: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				
				// Broadcast to local subscribers
				var payload map[string]any
				if err := json.Unmarshal(d.Body, &payload); err == nil {
					if chatID, ok := payload["chatId"].(float64); ok {
						s.hub.BroadcastToChat(int64(chatID), d.Body)
					}
				}
				
				d.Ack(false)
			}
		}
	}()

	return nil
}
