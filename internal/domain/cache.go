package domain

import (
	"context"
	"time"
)

// CacheRepository defines the interface for caching and ephemeral data
type CacheRepository interface {
	// Presence
	SetPresence(ctx context.Context, userID int64, online bool, ttl time.Duration) error
	GetPresence(ctx context.Context, userID int64) (online bool, lastSeen int64, err error)

	// Group Members Caching
	AddGroupMembers(ctx context.Context, chatID int64, userIDs []int64) error
	GetGroupMembers(ctx context.Context, chatID int64) ([]int64, error)
	RemoveGroupMember(ctx context.Context, chatID, userID int64) error

	// Connection Tracking (Gateway)
	RegisterConnection(ctx context.Context, userID int64, device, gwPodIP string, ttl time.Duration) error
	UnregisterConnection(ctx context.Context, userID int64, device string) error
	GetConnection(ctx context.Context, userID int64, device string) (string, error)
}
