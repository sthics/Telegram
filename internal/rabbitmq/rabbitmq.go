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
