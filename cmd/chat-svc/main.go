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
	"github.com/ambarg/mini-telegram/internal/database"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/redis"
	amqp "github.com/rabbitmq/amqp091-go"
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

	// Declare exchanges
	if err := rmqClient.DeclareExchanges(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare exchanges")
	}

	// Declare shared chat queue (best practice for scalable systems)
	if err := rmqClient.DeclareSharedChatQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare shared chat queue")
	}

	// Create chat service
	svc := NewChatService(db, redisClient, rmqClient)

	log.Info().Msg("chat service started, waiting for messages...")

	// Create a channel to listen for consumer messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a worker pool - all workers compete for messages from the shared queue
	// This provides automatic load balancing and scales horizontally
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		go svc.Worker(ctx, i)
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

// ChatService handles chat message processing
type ChatService struct {
	db       *database.DB
	redis    *redis.Client
	rabbitmq *rabbitmq.Client
}

// NewChatService creates a new chat service
func NewChatService(db *database.DB, redisClient *redis.Client, rmqClient *rabbitmq.Client) *ChatService {
	return &ChatService{
		db:       db,
		redis:    redisClient,
		rabbitmq: rmqClient,
	}
}

// Worker processes messages from the shared chat queue
// All workers compete for messages, providing automatic load balancing
func (s *ChatService) Worker(ctx context.Context, workerID int) {
	logger := log.With().Int("worker_id", workerID).Logger()
	logger.Info().Msg("worker started")

	consumerTag := fmt.Sprintf("chat-worker-%d", workerID)

	// Start consuming from the shared chat queue
	// This is best practice: all workers compete for messages from a single queue
	msgs, err := s.rabbitmq.ConsumeSharedChatQueue(consumerTag)
	if err != nil {
		logger.Error().Err(err).Msg("failed to start consuming from shared queue")
		return
	}

	logger.Info().Str("queue", "chat.messages").Msg("consuming messages from shared queue")

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
			// Process the message
			if err := s.ProcessMessage(ctx, delivery); err != nil {
				logger.Error().Err(err).Msg("failed to process message")
			}
		}
	}
}

// ProcessMessage processes a single chat message
func (s *ChatService) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	start := time.Now()

	// Parse message payload
	var payload struct {
		UUID   string `json:"uuid"`
		ChatID int64  `json:"chatId"`
		UserID int64  `json:"userId"`
		Body   string `json:"body"`
	}

	if err := json.Unmarshal(delivery.Body, &payload); err != nil {
		log.Error().Err(err).Msg("failed to parse message payload")
		delivery.Nack(false, false) // Don't requeue invalid messages
		return err
	}

	logger := log.With().
		Str("uuid", payload.UUID).
		Int64("chat_id", payload.ChatID).
		Int64("user_id", payload.UserID).
		Logger()

	logger.Info().Msg("processing message")

	// 1. Persist message to Postgres
	msg := &database.Message{
		ChatID: payload.ChatID,
		UserID: payload.UserID,
		Body:   payload.Body,
	}

	if err := s.db.CreateMessage(ctx, msg); err != nil {
		logger.Error().Err(err).Msg("failed to persist message")
		delivery.Nack(false, true) // Requeue for retry
		return err
	}

	logger.Info().Int64("msg_id", msg.ID).Msg("message persisted")

	// 2. Read member list from Redis (with DB fallback)
	members, err := s.redis.GetGroupMembers(ctx, payload.ChatID)
	if err != nil || len(members) == 0 {
		// Fallback to database
		members, err = s.db.GetChatMembers(ctx, payload.ChatID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get chat members")
			delivery.Nack(false, true)
			return err
		}

		// Cache in Redis for next time
		if err := s.redis.AddGroupMembers(ctx, payload.ChatID, members); err != nil {
			logger.Warn().Err(err).Msg("failed to cache group members")
		}
	}

	// 3. Create receipts for all members
	for _, memberID := range members {
		receipt := &database.Receipt{
			MsgID:  msg.ID,
			UserID: memberID,
			Status: database.ReceiptStatusSent,
		}
		if err := s.db.CreateReceipt(ctx, receipt); err != nil {
			logger.Warn().Err(err).Int64("member_id", memberID).Msg("failed to create receipt")
		}
	}

	// 4. Publish delivery event to delivery exchange
	deliveryPayload, _ := json.Marshal(map[string]interface{}{
		"type":      "Message",
		"msgId":     msg.ID,
		"chatId":    payload.ChatID,
		"userId":    payload.UserID,
		"body":      payload.Body,
		"createdAt": msg.CreatedAt.Unix(),
	})

	if err := s.rabbitmq.PublishToDeliveryExchange(ctx, payload.ChatID, deliveryPayload); err != nil {
		logger.Error().Err(err).Msg("failed to publish delivery event")
		delivery.Nack(false, true)
		return err
	}

	// 5. Send delivered acknowledgment back to sender
	deliveredPayload, _ := json.Marshal(map[string]interface{}{
		"type":  "Delivered",
		"uuid":  payload.UUID,
		"msgId": msg.ID,
	})

	if err := s.rabbitmq.PublishToDeliveryExchange(ctx, payload.ChatID, deliveredPayload); err != nil {
		logger.Warn().Err(err).Msg("failed to publish delivered ack")
	}

	// ACK the message
	if err := delivery.Ack(false); err != nil {
		logger.Error().Err(err).Msg("failed to ack message")
		return err
	}

	duration := time.Since(start)
	logger.Info().
		Dur("duration_ms", duration).
		Msg("message processed successfully")

	return nil
}
