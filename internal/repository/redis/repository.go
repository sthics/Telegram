package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client
type Client struct {
	*redis.Client
}

// Config holds Redis configuration
type Config struct {
	Addr     string
	Password string
	DB       int
}

// New creates a new Redis client
func New(cfg Config) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Enable tracing
	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, fmt.Errorf("failed to instrument redis with tracing: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{client}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.Client.Close()
}

// CacheRepository implementation
type CacheRepository struct {
	client *Client
}

func NewCacheRepository(client *Client) *CacheRepository {
	return &CacheRepository{client: client}
}

// RegisterConnection registers a WebSocket connection in Redis
func (r *CacheRepository) RegisterConnection(ctx context.Context, userID int64, device, gwPodIP string, ttl time.Duration) error {
	key := fmt.Sprintf("conn:%d:%s", userID, device)
	if err := r.client.Set(ctx, key, gwPodIP, ttl).Err(); err != nil {
		return fmt.Errorf("failed to register connection: %w", err)
	}
	return nil
}

// UnregisterConnection removes a WebSocket connection from Redis
func (r *CacheRepository) UnregisterConnection(ctx context.Context, userID int64, device string) error {
	key := fmt.Sprintf("conn:%d:%s", userID, device)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to unregister connection: %w", err)
	}
	return nil
}

// GetConnection retrieves the gateway pod IP for a connection
func (r *CacheRepository) GetConnection(ctx context.Context, userID int64, device string) (string, error) {
	key := fmt.Sprintf("conn:%d:%s", userID, device)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("connection not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get connection: %w", err)
	}
	return val, nil
}

// SetPresence sets user presence.
// If online is true, it stores the current timestamp.
// If online is false, it stores the current timestamp as a negative value (explicit offline).
func (r *CacheRepository) SetPresence(ctx context.Context, userID int64, online bool, ttl time.Duration) error {
	key := fmt.Sprintf("pres:%d", userID)
	timestamp := time.Now().Unix()
	if !online {
		timestamp = -timestamp
	}
	
	value := fmt.Sprintf("%d", timestamp)

	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set presence: %w", err)
	}
	return nil
}

// GetPresence retrieves user presence
func (r *CacheRepository) GetPresence(ctx context.Context, userID int64) (online bool, lastSeen int64, err error) {
	key := fmt.Sprintf("pres:%d", userID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("failed to get presence: %w", err)
	}

	if val == "0" {
		return false, 0, nil
	}

	var timestamp int64
	_, err = fmt.Sscanf(val, "%d", &timestamp)
	if err != nil {
		return false, 0, fmt.Errorf("failed to parse presence timestamp: %w", err)
	}

	// If negative, it means explicit offline
	if timestamp < 0 {
		return false, -timestamp, nil
	}

	// Consider online if last seen within 60 seconds
	if time.Since(time.Unix(timestamp, 0)) < 60*time.Second {
		return true, timestamp, nil
	}

	return false, timestamp, nil
}

// AddGroupMembers adds members to a group cache
func (r *CacheRepository) AddGroupMembers(ctx context.Context, chatID int64, userIDs []int64) error {
	key := fmt.Sprintf("grp:%d", chatID)
	members := make([]interface{}, len(userIDs))
	for i, uid := range userIDs {
		members[i] = uid
	}

	if err := r.client.SAdd(ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("failed to add group members: %w", err)
	}
	return nil
}

// GetGroupMembers retrieves all members of a group
func (r *CacheRepository) GetGroupMembers(ctx context.Context, chatID int64) ([]int64, error) {
	key := fmt.Sprintf("grp:%d", chatID)
	vals, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	userIDs := make([]int64, 0, len(vals))
	for _, val := range vals {
		var uid int64
		_, err := fmt.Sscanf(val, "%d", &uid)
		if err == nil {
			userIDs = append(userIDs, uid)
		}
	}

	return userIDs, nil
}

// RemoveGroupMember removes a member from a group cache
func (r *CacheRepository) RemoveGroupMember(ctx context.Context, chatID, userID int64) error {
	key := fmt.Sprintf("grp:%d", chatID)
	if err := r.client.SRem(ctx, key, userID).Err(); err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}
	return nil
}
