package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ambarg/mini-telegram/internal/config"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/repository/postgres"
	"github.com/ambarg/mini-telegram/internal/repository/redis"
	"github.com/ambarg/mini-telegram/internal/service/push"
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
	shutdown, err := telemetry.InitTracer("push-svc", cfg.OtelCollectorURL)
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

	// Declare shared chat queue (idempotent)
	if err := rmqClient.DeclareSharedChatQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare shared chat queue")
	}

	// Initialize Repositories
	chatRepo := postgres.NewChatRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)

	// Initialize Service
	svc := push.NewService(chatRepo, cacheRepo)

	// Start consumer
	msgs, err := rmqClient.ConsumeSharedChatQueue("push-svc")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start consuming")
	}

	log.Info().Msg("push-svc started")

	// Process messages
	go func() {
		for d := range msgs {
			ctx := context.Background()
			if err := svc.ProcessPushNotification(ctx, d.Body); err != nil {
				log.Error().Err(err).Msg("failed to process push notification")
				d.Ack(false) // Ack anyway to prevent loop for now, or Nack if retryable
			} else {
				d.Ack(false)
			}
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down push-svc...")
	
	// Give workers time to finish
	time.Sleep(2 * time.Second)
	log.Info().Msg("push-svc exited")
}
