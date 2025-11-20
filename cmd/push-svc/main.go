package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/ambarg/mini-telegram/internal/config"
	"github.com/ambarg/mini-telegram/internal/database"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/redis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Load configuration
	cfg := config.MustLoad()

	// Initialize services
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

	// Declare shared chat queue (idempotent)
	if err := rmqClient.DeclareSharedChatQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare shared chat queue")
	}

	// Start consumer
	msgs, err := rmqClient.ConsumeSharedChatQueue("push-svc")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start consuming")
	}

	log.Info().Msg("push-svc started")

	// Process messages
	go func() {
		for d := range msgs {
			var msg map[string]any
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal message")
				d.Ack(false)
				continue
			}

			chatID, _ := msg["chatId"].(float64)
			senderID, _ := msg["userId"].(float64)
			body, _ := msg["body"].(string)

			ctx := context.Background()

			// Get chat members
			members, err := db.GetChatMembers(ctx, int64(chatID))
			if err != nil {
				log.Error().Err(err).Msg("failed to get chat members")
				d.Ack(false)
				continue
			}

			log.Info().Int64("chat_id", int64(chatID)).Msg("Processing message for push")

			for _, memberID := range members {
				// Skip sender
				if memberID == int64(senderID) {
					continue
				}

				// Check presence
				online, _, err := redisClient.GetPresence(ctx, memberID)
				if err != nil {
					log.Error().Err(err).Int64("user_id", memberID).Msg("failed to check presence")
					continue
				}

				log.Info().Int64("user_id", memberID).Bool("online", online).Msg("User presence check")

				if !online {
					// User is offline, send push
					tokens, err := db.GetDeviceTokens(ctx, memberID)
					if err != nil {
						log.Error().Err(err).Int64("user_id", memberID).Msg("failed to get device tokens")
						continue
					}
					
					log.Info().Int64("user_id", memberID).Int("token_count", len(tokens)).Msg("Found device tokens")

					for _, token := range tokens {
						log.Info().
							Int64("user_id", memberID).
							Str("token", token).
							Str("body", body).
							Msg("Sending push notification")
					}
				}
			}

			d.Ack(false)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down push-svc...")
}
