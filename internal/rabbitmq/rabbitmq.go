package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client wraps RabbitMQ connection and channel
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// Config holds RabbitMQ configuration
type Config struct {
	URL string
}

// New creates a new RabbitMQ client
func New(cfg Config) (*Client, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set prefetch count for fair dispatch
	if err := channel.Qos(20, 0, false); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: channel,
	}, nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// DeclareExchanges declares the required exchanges
func (c *Client) DeclareExchanges() error {
	// Declare chat.topic exchange
	if err := c.channel.ExchangeDeclare(
		"chat.topic",    // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	); err != nil {
		return fmt.Errorf("failed to declare chat.topic exchange: %w", err)
	}

	// Declare delivery.topic exchange
	if err := c.channel.ExchangeDeclare(
		"delivery.topic", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	); err != nil {
		return fmt.Errorf("failed to declare delivery.topic exchange: %w", err)
	}

	// Declare presence.fanout exchange for broadcasting presence updates
	if err := c.channel.ExchangeDeclare(
		"presence.fanout", // name
		"fanout",          // type - fanout broadcasts to all bound queues
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	); err != nil {
		return fmt.Errorf("failed to declare presence.fanout exchange: %w", err)
	}

	return nil
}

// DeclareSharedChatQueue declares a single shared queue for all chat messages
// This follows best practices for scalable message processing systems
func (c *Client) DeclareSharedChatQueue() error {
	queueName := "chat.messages"

	// Declare queue with lazy mode and TTL
	args := amqp.Table{
		"x-queue-mode":   "lazy",
		"x-message-ttl":  86400000, // 24 hours in milliseconds
		"x-max-priority": 3,        // Support message priorities
	}

	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare shared chat queue: %w", err)
	}

	// Bind queue to exchange with wildcard routing key to capture all chat messages
	if err := c.channel.QueueBind(
		queueName,    // queue name
		"*",          // routing key (wildcard to match all chat IDs)
		"chat.topic", // exchange
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("failed to bind shared chat queue: %w", err)
	}

	return nil
}

// DeclareChatQueue declares a queue for a specific chat
func (c *Client) DeclareChatQueue(chatID int64) error {
	queueName := fmt.Sprintf("chat.%d", chatID)
	routingKey := fmt.Sprintf("%d", chatID)

	// Declare queue with lazy mode and TTL
	args := amqp.Table{
		"x-queue-mode": "lazy",
		"x-message-ttl": 86400000, // 24 hours in milliseconds
	}

	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	if err := c.channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		"chat.topic", // exchange
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

// PublishToChatQueue publishes a message to a chat queue
func (c *Client) PublishToChatQueue(ctx context.Context, chatID int64, body []byte) error {
	routingKey := fmt.Sprintf("%d", chatID)

	// Enable publisher confirms
	if err := c.channel.Confirm(false); err != nil {
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	confirms := c.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	err := c.channel.PublishWithContext(
		ctx,
		"chat.topic", // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         body,
			DeliveryMode: amqp.Persistent, // persistent messages
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation with timeout
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("publish not acknowledged")
		}
	case <-time.After(500 * time.Millisecond):
		return fmt.Errorf("publish confirmation timeout")
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// PublishToDeliveryExchange publishes a delivery event
func (c *Client) PublishToDeliveryExchange(ctx context.Context, chatID int64, body []byte) error {
	routingKey := fmt.Sprintf("%d", chatID)

	err := c.channel.PublishWithContext(
		ctx,
		"delivery.topic", // exchange
		routingKey,       // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish delivery event: %w", err)
	}

	return nil
}

// ConsumeChatQueue starts consuming messages from a chat queue
// Deprecated: Use ConsumeSharedChatQueue for better scalability
func (c *Client) ConsumeChatQueue(chatID int64, consumerTag string) (<-chan amqp.Delivery, error) {
	queueName := fmt.Sprintf("chat.%d", chatID)

	msgs, err := c.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer tag
		false,       // auto-ack (we'll manually ack)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	return msgs, nil
}

// ConsumeSharedChatQueue starts consuming from the shared chat messages queue
// This is the recommended approach for scalable message processing
func (c *Client) ConsumeSharedChatQueue(consumerTag string) (<-chan amqp.Delivery, error) {
	queueName := "chat.messages"

	msgs, err := c.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer tag
		false,       // auto-ack (we'll manually ack for reliability)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming from shared queue: %w", err)
	}

	return msgs, nil
}

// DeclarePresenceQueue declares a shared queue for presence events
func (c *Client) DeclarePresenceQueue() error {
	queueName := "presence.events"

	// Declare queue
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare presence queue: %w", err)
	}

	return nil
}

// DeclareReadReceiptQueue declares a shared queue for read receipts
func (c *Client) DeclareReadReceiptQueue() error {
	queueName := "read.receipts"

	// Declare queue for batching read receipts
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare read receipt queue: %w", err)
	}

	return nil
}

// PublishPresenceEvent publishes a presence update
func (c *Client) PublishPresenceEvent(ctx context.Context, body []byte) error {
	err := c.channel.PublishWithContext(
		ctx,
		"presence.fanout", // exchange
		"",                // routing key (ignored for fanout)
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Transient, // Don't persist presence events
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish presence event: %w", err)
	}

	return nil
}

// PublishTypingEvent publishes a typing indicator event
func (c *Client) PublishTypingEvent(ctx context.Context, chatID int64, body []byte) error {
	routingKey := fmt.Sprintf("%d", chatID)

	err := c.channel.PublishWithContext(
		ctx,
		"delivery.topic", // exchange
		routingKey,       // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Transient, // Transient for ephemeral events
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish typing event: %w", err)
	}

	return nil
}

// PublishReadReceipt publishes a read receipt to the queue
func (c *Client) PublishReadReceipt(ctx context.Context, body []byte) error {
	err := c.channel.PublishWithContext(
		ctx,
		"",              // exchange (empty = default)
		"read.receipts", // routing key (queue name)
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish read receipt: %w", err)
	}

	return nil
}

// ConsumePresenceQueue starts consuming from the presence queue
func (c *Client) ConsumePresenceQueue(consumerTag string) (<-chan amqp.Delivery, error) {
	queueName := "presence.events"

	msgs, err := c.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer tag
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume presence queue: %w", err)
	}

	return msgs, nil
}

// ConsumeReadReceiptQueue starts consuming from the read receipt queue
func (c *Client) ConsumeReadReceiptQueue(consumerTag string) (<-chan amqp.Delivery, error) {
	queueName := "read.receipts"

	msgs, err := c.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer tag
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume read receipt queue: %w", err)
	}

	return msgs, nil
}

// DeclareDeliveryQueue declares a delivery queue for a gateway pod
func (c *Client) DeclareDeliveryQueue(podID string, chatIDs []int64) (string, error) {
	queueName := fmt.Sprintf("delivery.%s", podID)

	_, err := c.channel.QueueDeclare(
		queueName, // name
		false,     // durable (transient queue per pod)
		true,      // delete when unused
		true,      // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return "", fmt.Errorf("failed to declare delivery queue: %w", err)
	}

	// Bind to all chat IDs
	for _, chatID := range chatIDs {
		routingKey := fmt.Sprintf("%d", chatID)
		if err := c.channel.QueueBind(
			queueName,        // queue name
			routingKey,       // routing key
			"delivery.topic", // exchange
			false,            // no-wait
			nil,              // arguments
		); err != nil {
			return "", fmt.Errorf("failed to bind delivery queue: %w", err)
		}
	}

	return queueName, nil
}

// BindDeliveryQueue binds a delivery queue to a chat ID
func (c *Client) BindDeliveryQueue(queueName string, chatID int64) error {
	routingKey := fmt.Sprintf("%d", chatID)
	if err := c.channel.QueueBind(
		queueName,        // queue name
		routingKey,       // routing key
		"delivery.topic", // exchange
		false,            // no-wait
		nil,              // arguments
	); err != nil {
		return fmt.Errorf("failed to bind delivery queue: %w", err)
	}
	return nil
}

// ConsumeDeliveryQueue starts consuming from a delivery queue
func (c *Client) ConsumeDeliveryQueue(queueName, consumerTag string) (<-chan amqp.Delivery, error) {
	msgs, err := c.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer tag
		false,       // auto-ack
		true,        // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume delivery queue: %w", err)
	}

	return msgs, nil
}
