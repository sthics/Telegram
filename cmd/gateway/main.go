package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ambarg/mini-telegram/internal/auth"
	"github.com/ambarg/mini-telegram/internal/config"
	httpHandler "github.com/ambarg/mini-telegram/internal/handler/http"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/repository/postgres"
	"github.com/ambarg/mini-telegram/internal/repository/redis"
	"github.com/ambarg/mini-telegram/internal/repository/s3"
	authService "github.com/ambarg/mini-telegram/internal/service/auth"
	chatService "github.com/ambarg/mini-telegram/internal/service/chat"
	mediaService "github.com/ambarg/mini-telegram/internal/service/media"
	"github.com/ambarg/mini-telegram/internal/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	docs "github.com/ambarg/mini-telegram/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Mini-Telegram API
// @version         1.0
// @description     This is the API server for Mini-Telegram.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Load configuration
	cfg := config.MustLoad()

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
	mediaRepo, err := s3.New(context.Background(), cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize S3 repository")
	}

	// Initialize Services
	authSvc := authService.NewService(userRepo, auth.NewService(privateKey))
	chatSvc := chatService.NewService(chatRepo, cacheRepo, rmqClient)
	mediaSvc := mediaService.NewService(mediaRepo)

	// Initialize Handlers
	authHandler := httpHandler.NewAuthHandler(authSvc)
	chatHandler := httpHandler.NewChatHandler(chatSvc)
	mediaHandler := httpHandler.NewMediaHandler(mediaSvc)
	userHandler := httpHandler.NewUserHandler(cacheRepo, userRepo)

	// Create WebSocket hub
	hub := websocket.NewHub(log.Logger)

	// Declare Delivery Queue for this Gateway instance
	podID, _ := os.Hostname() // Use hostname as pod ID
	queueName, err := rmqClient.DeclareDeliveryQueue(podID, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to declare delivery queue")
	}

	// Initialize WebSocket Handler
	wsHandler := httpHandler.NewWebSocketHandler(hub, chatSvc, auth.NewService(privateKey), cacheRepo, rmqClient, queueName)

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
				d.Ack(false)
				continue
			}

			// Broadcast to chat members connected to this gateway
			hub.BroadcastToChat(int64(chatID), d.Body)
			d.Ack(false)
		}
	}()

	// Setup Router
	r := gin.Default()
	r.Use(otelgin.Middleware("gateway"))

	// CORS Setup
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"}, // Allow local dev and docker web
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger
	docs.SwaggerInfo.BasePath = "/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		protected.PATCH("/chats/:id", chatHandler.UpdateGroupInfo)
		protected.POST("/chats/:id/invite", chatHandler.InviteToChat)
		protected.DELETE("/chats/:id/members/:userId", chatHandler.KickMember)
		protected.DELETE("/chats/:id/members", chatHandler.LeaveChat)
		protected.POST("/chats/:id/members/:userId/promote", chatHandler.PromoteMember)
		protected.POST("/chats/:id/members/:userId/demote", chatHandler.DemoteMember)
		protected.GET("/chats/:id/messages", chatHandler.GetMessages)
		protected.POST("/chats/:id/messages", chatHandler.SendMessage)
		protected.POST("/chats/:id/read", chatHandler.MarkRead) // New route
		protected.GET("/chats/:id/members", chatHandler.GetChatMembers)
		protected.POST("/devices", chatHandler.RegisterDevice)

		// Media routes
		protected.POST("/uploads/presigned", mediaHandler.GetUploadURL)

		// User routes
		protected.GET("/users/me", userHandler.GetProfile)
		protected.PATCH("/users/me", userHandler.UpdateProfile)
		protected.GET("/users/:id/presence", userHandler.GetUserPresence)
		protected.GET("/users", userHandler.SearchUsers)
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
