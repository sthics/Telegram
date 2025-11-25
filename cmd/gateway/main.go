package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ambarg/mini-telegram/internal/auth"
	"github.com/ambarg/mini-telegram/internal/config"
	httpHandler "github.com/ambarg/mini-telegram/internal/handler/http"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/repository/postgres"
	"github.com/ambarg/mini-telegram/internal/repository/redis"
	authService "github.com/ambarg/mini-telegram/internal/service/auth"
	chatService "github.com/ambarg/mini-telegram/internal/service/chat"
	"github.com/ambarg/mini-telegram/internal/telemetry"
	"github.com/ambarg/mini-telegram/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Load configuration
	cfg := config.MustLoad()

	// Initialize Tracer
	shutdown, err := telemetry.InitTracer("gateway", cfg.OtelCollectorURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize tracer")
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed to shutdown tracer")
		}
	}()


	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Load JWT private key
	privateKey, err := auth.LoadPrivateKey(cfg.JWTPrivateKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load JWT private key")
	}

	// Initialize Infrastructure (Repositories)
	db, err := postgres.New(postgres.Config{
		DSN:             cfg.DSN,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	redisClient, err := redis.New(redis.Config{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	defer redisClient.Close()

	rmqClient, err := rabbitmq.New(rabbitmq.Config{
		URL: cfg.AMQPURL,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to RabbitMQ")
	}
	defer rmqClient.Close()

	// Declare exchanges
	if err := rmqClient.DeclareExchanges(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare exchanges")
	}

	// Declare shared chat queue
	if err := rmqClient.DeclareSharedChatQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare shared chat queue")
	}

	// Initialize Repositories
	userRepo := postgres.NewUserRepository(db)
	chatRepo := postgres.NewChatRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)

	// Initialize Services
	authSvc := authService.NewService(userRepo, auth.NewService(privateKey))
	chatSvc := chatService.NewService(chatRepo, cacheRepo, rmqClient)

	// Initialize Handlers
	authHandler := httpHandler.NewAuthHandler(authSvc)
	chatHandler := httpHandler.NewChatHandler(chatSvc)

	// Create WebSocket hub
	hub := websocket.NewHub(log.Logger)

	// Declare Delivery Queue for this Gateway instance
	podID, _ := os.Hostname() // Use hostname as pod ID
	queueName, err := rmqClient.DeclareDeliveryQueue(podID, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to declare delivery queue")
	}

	// Initialize WebSocket Handler
	wsHandler := httpHandler.NewWebSocketHandler(hub, chatSvc, auth.NewService(privateKey), rmqClient, queueName)

	// Start RabbitMQ Consumer for Delivery
	msgs, err := rmqClient.ConsumeDeliveryQueue(queueName, "gateway-"+podID)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start delivery consumer")
	}

	go func() {
		for d := range msgs {
			var msg map[string]any
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal delivery message")
				d.Ack(false)
				continue
			}

			chatID, ok := msg["chatId"].(float64)
			if !ok {
				// Try int64 if float64 fails (though JSON unmarshals numbers to float64)
				// Or maybe it's missing
				d.Ack(false)
				continue
			}

			// Broadcast to chat members connected to this gateway
			hub.BroadcastToChat(int64(chatID), d.Body)
			d.Ack(false)
		}
	}()

	// Let's create the router here directly
	r := gin.Default()
	r.Use(otelgin.Middleware("gateway"))

	// Health check
	r.GET("/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// WebSocket route
	r.GET("/v1/ws", wsHandler.HandleWS)

	// Auth routes
	authGroup := r.Group("/v1/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
	}

	// Protected routes
	protected := r.Group("/v1")
	jwtMiddleware := auth.NewService(privateKey).JWTMiddleware()
	protected.Use(jwtMiddleware)
	{
		// Chat routes
		protected.GET("/chats", chatHandler.GetChats)
		protected.POST("/chats", chatHandler.CreateChat)
		protected.POST("/chats/:id/invite", chatHandler.InviteToChat)
		protected.DELETE("/chats/:id/members/:userId", chatHandler.KickMember)
		protected.DELETE("/chats/:id/members", chatHandler.LeaveChat)
		protected.POST("/devices", chatHandler.RegisterDevice)
	}

	// Start server
	go func() {
		log.Info().Int("port", cfg.Port).Msg("starting gateway server")
		if err := r.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")
}
