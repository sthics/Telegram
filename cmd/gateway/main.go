package main

import (
	"context"
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
	_ = websocket.NewHub(log.Logger) // Hub unused for now

	// Let's create the router here directly
	r := gin.Default()
	r.Use(otelgin.Middleware("gateway"))

	// Health check
	r.GET("/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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
