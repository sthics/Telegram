package domain

import (
	"context"
)

// MessageBroker defines the interface for messaging operations
type MessageBroker interface {
	PublishToChatQueue(ctx context.Context, chatID int64, payload []byte) error
	PublishToDeliveryExchange(ctx context.Context, chatID int64, payload []byte) error
	PublishReadReceipt(ctx context.Context, payload []byte) error
	PublishTypingEvent(ctx context.Context, chatID int64, payload []byte) error
	PublishPresenceEvent(ctx context.Context, payload []byte) error
	
	BindDeliveryQueue(queueName string, chatID int64) error
}
