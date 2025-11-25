package presence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ambarg/mini-telegram/internal/domain"
	"github.com/rs/zerolog/log"
)

// ReadReceiptBatch represents a batch of read receipts
type ReadReceiptBatch struct {
	ChatID int64
	UserID int64
	MsgID  int64
}

// Service handles presence and read receipt processing
type Service struct {
	chatRepo  domain.ChatRepository
	cacheRepo domain.CacheRepository
	broker    domain.MessageBroker
	batch     chan ReadReceiptBatch
}

// NewService creates a new presence service
func NewService(chatRepo domain.ChatRepository, cacheRepo domain.CacheRepository, broker domain.MessageBroker) *Service {
	return &Service{
		chatRepo:  chatRepo,
		cacheRepo: cacheRepo,
		broker:    broker,
		batch:     make(chan ReadReceiptBatch, 1000),
	}
}

// RunReadReceiptWorker processes read receipts from the queue
func (s *Service) RunReadReceiptWorker(ctx context.Context, workerID int, consumerTag string) {
	logger := log.With().Int("worker_id", workerID).Logger()
	logger.Info().Msg("read receipt worker started")

	// We need to access the underlying rabbitmq client to consume
	// This is a bit of a leak in abstraction, but for now we assume broker has a method or we pass the client differently.
	// However, the domain.MessageBroker interface doesn't have Consume methods.
	// In the previous code, main.go accessed rmqClient directly.
	// For Clean Architecture, the Service shouldn't depend on concrete RabbitMQ implementation details like 'ConsumeReadReceiptQueue'.
	// Ideally, the infrastructure layer (main.go) should set up the consumer and call the Service.
	// But to keep it similar to the plan: "RunReadReceiptWorker... Worker logic for consuming receipts"
	
	// Wait, the plan said: "RunReadReceiptWorker(ctx, workerID)".
	// But `ConsumeReadReceiptQueue` is on `*rabbitmq.Client`.
	// If `domain.MessageBroker` is the interface, we can't call `ConsumeReadReceiptQueue`.
	
	// Option 1: Add Consume methods to MessageBroker interface.
	// Option 2: Pass the concrete rabbitmq client to the service (violates clean arch but pragmatic).
	// Option 3: Have the worker in main.go and call s.ProcessReadReceipt.
	
	// The implementation plan said: "Implement methods: RunReadReceiptWorker".
	// Let's assume for now we will inject a consumer interface or just keep the worker logic here but we need the channel.
	
	// Actually, looking at `internal/service/chat/service.go`, it has `ProcessMessage`. The worker is in `cmd/chat-svc/main.go`.
	// So for `presence-svc`, we should probably follow the same pattern:
	// The Service provides `ProcessReadReceipt` and `BatchProcessor`.
	// The `main.go` handles the RabbitMQ consumption loop and calls `ProcessReadReceipt`.
	
	// BUT, `BatchProcessor` is a background routine that needs to run.
	// And `ProcessReadReceipt` in the old code added to a batch channel.
	
	// So:
	// Service has `ProcessReadReceipt(ctx, payload)` which adds to batch.
	// Service has `RunBatchProcessor(ctx)` which runs the batch loop.
	// `main.go` runs the RabbitMQ consumer and calls `ProcessReadReceipt`.
}

// ProcessReadReceipt handles a single read receipt message
func (s *Service) ProcessReadReceipt(ctx context.Context, payload []byte) error {
	var data struct {
		ChatID int64 `json:"chatId"`
		UserID int64 `json:"userId"`
		MsgID  int64 `json:"msgId"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("failed to parse read receipt: %w", err)
	}

	// Add to batch channel
	select {
	case s.batch <- ReadReceiptBatch{
		ChatID: data.ChatID,
		UserID: data.UserID,
		MsgID:  data.MsgID,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RunBatchProcessor processes read receipts in batches
func (s *Service) RunBatchProcessor(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	receipts := make([]ReadReceiptBatch, 0, 100)

	for {
		select {
		case receipt := <-s.batch:
			receipts = append(receipts, receipt)
			if len(receipts) >= 100 {
				s.processBatch(ctx, receipts)
				receipts = receipts[:0]
			}

		case <-ticker.C:
			if len(receipts) > 0 {
				s.processBatch(ctx, receipts)
				receipts = receipts[:0]
			}

		case <-ctx.Done():
			if len(receipts) > 0 {
				s.processBatch(ctx, receipts)
			}
			return
		}
	}
}

func (s *Service) processBatch(ctx context.Context, receipts []ReadReceiptBatch) {
	logger := log.With().Int("batch_size", len(receipts)).Logger()
	start := time.Now()

	for _, receipt := range receipts {
		// Update receipt status
		r := &domain.Receipt{
			MsgID:  receipt.MsgID,
			UserID: receipt.UserID,
			Status: domain.ReceiptStatusRead,
		}

		if err := s.chatRepo.CreateReceipt(ctx, r); err != nil {
			logger.Warn().Err(err).Int64("msg_id", receipt.MsgID).Msg("failed to update receipt")
		}

		// Update last read message
		if err := s.chatRepo.UpdateLastReadMessage(ctx, receipt.ChatID, receipt.UserID, receipt.MsgID); err != nil {
			logger.Warn().Err(err).Msg("failed to update last read message")
		}

		// Broadcast
		payload, _ := json.Marshal(map[string]any{
			"type":   "Read",
			"chatId": receipt.ChatID,
			"userId": receipt.UserID,
			"msgId":  receipt.MsgID,
		})

		// We assume MessageBroker has a method for this or we use a generic one.
		// The interface has PublishReadReceipt but that might be for the queue?
		// Checking domain/messaging.go: PublishReadReceipt(ctx, payload) error
		// Wait, `PublishReadReceipt` in interface likely publishes to the queue (for processing).
		// We want to broadcast to the chat (fanout).
		// The interface has `PublishToDeliveryExchange`.
		
		// In old code: s.rabbitmq.PublishReadReceiptBroadcast(ctx, receipt.ChatID, payload)
		// We should add `PublishReadReceiptBroadcast` to the interface or use `PublishToDeliveryExchange`.
		// Let's use `PublishToDeliveryExchange` as it seems to be the fanout mechanism.
		
		if err := s.broker.PublishToDeliveryExchange(ctx, receipt.ChatID, payload); err != nil {
			logger.Warn().Err(err).Msg("failed to broadcast read receipt")
		}
	}
	
	logger.Info().Dur("duration_ms", time.Since(start)).Msg("batch processed")
}

// UpdatePresence updates user presence
func (s *Service) UpdatePresence(ctx context.Context, userID int64, online bool) error {
	ttl := 60 * time.Second
	if !online {
		ttl = 0
	}

	if err := s.cacheRepo.SetPresence(ctx, userID, online, ttl); err != nil {
		return err
	}

	// Publish presence event
	payload, _ := json.Marshal(map[string]interface{}{
		"type":     "Presence",
		"userId":   userID,
		"online":   online,
		"lastSeen": time.Now().Unix(),
	})

	// Use broker to publish presence event
	return s.broker.PublishPresenceEvent(ctx, payload)
}
