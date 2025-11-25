package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ambarg/mini-telegram/internal/config"
	"github.com/ambarg/mini-telegram/internal/domain"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/repository/postgres"
	"github.com/ambarg/mini-telegram/internal/repository/redis"
	chatService "github.com/ambarg/mini-telegram/internal/service/chat"
	"github.com/ambarg/mini-telegram/internal/telemetry"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Load configuration
	cfg := config.MustLoad()

	// Initialize Tracer
	shutdown, err := telemetry.InitTracer("chat-svc", cfg.OtelCollectorURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize tracer")
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed to shutdown tracer")
		}
	}()

	// Initialize Infrastructure
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
	chatRepo := postgres.NewChatRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)

	// Initialize Service
	svc := chatService.NewService(chatRepo, cacheRepo, rmqClient)

	log.Info().Msg("chat service started, waiting for messages...")

	// Create a channel to listen for consumer messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a worker pool
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		go runWorker(ctx, i, svc, rmqClient)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down chat service...")
	cancel()

	// Give workers time to finish
	time.Sleep(2 * time.Second)
	log.Info().Msg("chat service exited")
}

func runWorker(ctx context.Context, workerID int, svc *chatService.Service, rmqClient *rabbitmq.Client) {
	logger := log.With().Int("worker_id", workerID).Logger()
	logger.Info().Msg("worker started")

	consumerTag := fmt.Sprintf("chat-worker-%d", workerID)

	msgs, err := rmqClient.ConsumeSharedChatQueue(consumerTag)
	if err != nil {
		logger.Error().Err(err).Msg("failed to start consuming from shared queue")
		return
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("worker stopped")
			return
		case delivery, ok := <-msgs:
			if !ok {
				logger.Warn().Msg("message channel closed")
				return
			}
			
			// Process message
			var payload struct {
				UUID   string `json:"uuid"`
				ChatID int64  `json:"chatId"`
				UserID int64  `json:"userId"`
				Body   string `json:"body"`
			}

			if err := json.Unmarshal(delivery.Body, &payload); err != nil {
				logger.Error().Err(err).Msg("failed to parse message payload")
				delivery.Nack(false, false)
				continue
			}

			msg := &domain.Message{
				ChatID: payload.ChatID,
				UserID: payload.UserID,
				Body:   payload.Body,
			}

			if err := svc.ProcessMessage(ctx, msg, payload.UUID); err != nil {
				logger.Error().Err(err).Msg("failed to process message")
				delivery.Nack(false, true)
				continue
			}

			delivery.Ack(false)
		}
	}
}
