package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ambarg/mini-telegram/internal/auth"
	"github.com/ambarg/mini-telegram/internal/config"
	"github.com/ambarg/mini-telegram/internal/database"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/redis"
	"github.com/ambarg/mini-telegram/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

	// Initialize services
	authService := auth.NewService(privateKey)

	db, err := database.New(database.Config{
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

	// Declare shared chat queue (best practice for scalable systems)
	if err := rmqClient.DeclareSharedChatQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare shared chat queue")
	}

	// Create WebSocket hub
	hub := websocket.NewHub(log.Logger)

	// Initialize gateway server
	server := NewGatewayServer(cfg, authService, db, redisClient, rmqClient, hub)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: server.Router(),
	}

	// Start server
	go func() {
		log.Info().Int("port", cfg.Port).Msg("starting gateway server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// Graceful shutdown with 15 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited")
}
