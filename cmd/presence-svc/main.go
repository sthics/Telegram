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

	// Declare presence and read receipt queues
	if err := rmqClient.DeclarePresenceQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare presence queue")
	}

	if err := rmqClient.DeclareReadReceiptQueue(); err != nil {
		log.Fatal().Err(err).Msg("failed to declare read receipt queue")
	}

	// Create presence service
	svc := NewPresenceService(db, redisClient, rmqClient)

	// Start workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start read receipt workers (3 workers for read receipts)
	numReadReceiptWorkers := 3
	for i := 0; i < numReadReceiptWorkers; i++ {
		go svc.ReadReceiptWorker(ctx, i)
	}

	// Start batch processor for read receipts
	go svc.BatchProcessor(ctx)

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

// PresenceService handles presence and read receipt processing
type PresenceService struct {
	db       *database.DB
	redis    *redis.Client
	rabbitmq *rabbitmq.Client
	batch    chan ReadReceiptBatch
}

// ReadReceiptBatch represents a batch of read receipts
type ReadReceiptBatch struct {
	ChatID int64
	UserID int64
	MsgID  int64
}

// NewPresenceService creates a new presence service
func NewPresenceService(db *database.DB, redisClient *redis.Client, rmqClient *rabbitmq.Client) *PresenceService {
	return &PresenceService{
		db:       db,
		redis:    redisClient,
		rabbitmq: rmqClient,
		batch:    make(chan ReadReceiptBatch, 1000),
	}
}

// ReadReceiptWorker processes read receipts from the queue
func (s *PresenceService) ReadReceiptWorker(ctx context.Context, workerID int) {
	logger := log.With().Int("worker_id", workerID).Logger()
	logger.Info().Msg("read receipt worker started")

	consumerTag := fmt.Sprintf("receipt-worker-%d", workerID)

	// Start consuming from read receipt queue
	msgs, err := s.rabbitmq.ConsumeReadReceiptQueue(consumerTag)
	if err != nil {
		logger.Error().Err(err).Msg("failed to start consuming read receipts")
		return
	}

	logger.Info().Str("queue", "read.receipts").Msg("consuming read receipts")

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
			// Process the read receipt
			if err := s.ProcessReadReceipt(ctx, delivery); err != nil {
				logger.Error().Err(err).Msg("failed to process read receipt")
			}
		}
	}
}

// ProcessReadReceipt processes a read receipt message
func (s *PresenceService) ProcessReadReceipt(ctx context.Context, delivery amqp.Delivery) error {
	var payload struct {
		ChatID int64 `json:"chatId"`
		UserID int64 `json:"userId"`
		MsgID  int64 `json:"msgId"`
	}

	if err := json.Unmarshal(delivery.Body, &payload); err != nil {
		log.Error().Err(err).Msg("failed to parse read receipt")
		delivery.Nack(false, false)
		return err
	}

	logger := log.With().
		Int64("chat_id", payload.ChatID).
		Int64("user_id", payload.UserID).
		Int64("msg_id", payload.MsgID).
		Logger()

	logger.Info().Msg("processing read receipt")

	// Add to batch channel
	select {
	case s.batch <- ReadReceiptBatch{
		ChatID: payload.ChatID,
		UserID: payload.UserID,
		MsgID:  payload.MsgID,
	}:
		delivery.Ack(false)
	case <-ctx.Done():
		delivery.Nack(false, true)
		return ctx.Err()
	}

	return nil
}

// BatchProcessor processes read receipts in batches
func (s *PresenceService) BatchProcessor(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	receipts := make([]ReadReceiptBatch, 0, 100)

	for {
		select {
		case receipt := <-s.batch:
			receipts = append(receipts, receipt)

			// Process when batch is full
			if len(receipts) >= 100 {
				s.processBatch(ctx, receipts)
				receipts = receipts[:0]
			}

		case <-ticker.C:
			// Process any pending receipts
			if len(receipts) > 0 {
				s.processBatch(ctx, receipts)
				receipts = receipts[:0]
			}

		case <-ctx.Done():
			// Process remaining receipts before exit
			if len(receipts) > 0 {
				s.processBatch(ctx, receipts)
			}
			return
		}
	}
}

// processBatch processes a batch of read receipts
func (s *PresenceService) processBatch(ctx context.Context, receipts []ReadReceiptBatch) {
	logger := log.With().Int("batch_size", len(receipts)).Logger()
	logger.Info().Msg("processing read receipt batch")

	start := time.Now()

	// Update receipts in database
	for _, receipt := range receipts {
		// Update receipt status to read
		dbReceipt := &database.Receipt{
			MsgID:  receipt.MsgID,
			UserID: receipt.UserID,
			Status: database.ReceiptStatusRead,
		}

		if err := s.db.CreateReceipt(ctx, dbReceipt); err != nil {
			logger.Warn().Err(err).
				Int64("msg_id", receipt.MsgID).
				Int64("user_id", receipt.UserID).
				Msg("failed to update receipt")
		}

		// Update last read message
		if err := s.db.UpdateLastReadMessage(ctx, receipt.ChatID, receipt.UserID, receipt.MsgID); err != nil {
			logger.Warn().Err(err).Msg("failed to update last read message")
		}

		// Broadcast read receipt to chat members
		payload, _ := json.Marshal(map[string]any{
			"type":   "Read",
			"chatId": receipt.ChatID,
			"userId": receipt.UserID,
			"msgId":  receipt.MsgID,
		})

		if err := s.rabbitmq.PublishReadReceiptBroadcast(ctx, receipt.ChatID, payload); err != nil {
			logger.Warn().Err(err).Msg("failed to broadcast read receipt")
		}
	}

	duration := time.Since(start)
	logger.Info().
		Dur("duration_ms", duration).
		Msg("batch processed")
}

// UpdatePresence updates user presence
func (s *PresenceService) UpdatePresence(ctx context.Context, userID int64, online bool) error {
	ttl := 60 * time.Second
	if !online {
		ttl = 0
	}

	if err := s.redis.SetPresence(ctx, userID, online, ttl); err != nil {
		return err
	}

	// Optionally publish presence event
	payload, _ := json.Marshal(map[string]interface{}{
		"type":     "Presence",
		"userId":   userID,
		"online":   online,
		"lastSeen": time.Now().Unix(),
	})

	// Publish to users who need to see this presence update
	// (This would require tracking which chats the user is in)
	_ = payload // Placeholder

	return nil
}
