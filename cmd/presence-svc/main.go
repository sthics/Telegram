package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ambarg/mini-telegram/internal/config"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/repository/postgres"
	"github.com/ambarg/mini-telegram/internal/repository/redis"
	"github.com/ambarg/mini-telegram/internal/service/presence"
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
	shutdown, err := telemetry.InitTracer("presence-svc", cfg.OtelCollectorURL)
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

	// Declare presence and read receipt queues
	if err := rmqClient.DeclarePresenceQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare presence queue")
	}

	if err := rmqClient.DeclareReadReceiptQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare read receipt queue")
	}

	// Initialize Repositories
	chatRepo := postgres.NewChatRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)

	// Initialize Service
	svc := presence.NewService(chatRepo, cacheRepo, rmqClient)

	// Start workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start read receipt workers
	numReadReceiptWorkers := 3
	for i := 0; i < numReadReceiptWorkers; i++ {
		go runReadReceiptWorker(ctx, i, svc, rmqClient)
	}

	// Start batch processor
	go svc.RunBatchProcessor(ctx)

	log.Info().Msg("presence service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down presence service...")
	cancel()

	// Give workers time to finish
	time.Sleep(2 * time.Second)
	log.Info().Msg("presence service exited")
}

func runReadReceiptWorker(ctx context.Context, workerID int, svc *presence.Service, rmqClient *rabbitmq.Client) {
	logger := log.With().Int("worker_id", workerID).Logger()
	logger.Info().Msg("read receipt worker started")

	consumerTag := fmt.Sprintf("receipt-worker-%d", workerID)

	msgs, err := rmqClient.ConsumeReadReceiptQueue(consumerTag)
	if err != nil {
		logger.Error().Err(err).Msg("failed to start consuming read receipts")
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
			
			if err := svc.ProcessReadReceipt(ctx, delivery.Body); err != nil {
				logger.Error().Err(err).Msg("failed to process read receipt")
				delivery.Nack(false, false) // Retry? Or drop? For now retry
			} else {
				delivery.Ack(false)
			}
		}
	}
}
