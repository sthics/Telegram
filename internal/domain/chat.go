package domain

import (
	"context"
	"time"
)

// Chat types
const (
	ChatTypeDirect = 1
	ChatTypeGroup  = 2
)

// Chat roles
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// Chat represents a chat room
type Chat struct {
	ID        int64     `json:"id"`
	Type      int16     `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatMember represents a user in a chat
type ChatMember struct {
	ChatID        int64     `json:"chat_id"`
	UserID        int64     `json:"user_id"`
	Role          Role      `json:"role"`
	LastReadMsgID int64     `json:"last_read_msg_id"`
	JoinedAt      time.Time `json:"joined_at"`
}

// Message represents a chat message
type Message struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	UserID    int64     `json:"user_id"`
	Body      string    `json:"body"`
	MediaURL  string    `json:"media_url,omitempty"`
	ReplyToID *int64    `json:"reply_to_id,omitempty"`
	Reactions []byte    `json:"reactions"` // JSONB
	CreatedAt time.Time `json:"created_at"`
}

// Receipt status
const (
	ReceiptStatusSent      = 1
	ReceiptStatusDelivered = 2
	ReceiptStatusRead      = 3
)

// Receipt represents message delivery status
type Receipt struct {
	MsgID  int64     `json:"msg_id"`
	UserID int64     `json:"user_id"`
	Status int16     `json:"status"`
	Ts     time.Time `json:"ts"`
}

// DeviceToken represents a push token
type DeviceToken struct {
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatRepository defines the interface for chat data access
type ChatRepository interface {
	CreateChat(ctx context.Context, chat *Chat) error
	GetChat(ctx context.Context, id int64) (*Chat, error)
	GetUserChats(ctx context.Context, userID int64) ([]Chat, error)
	
	AddMember(ctx context.Context, member *ChatMember) error
	RemoveMember(ctx context.Context, chatID, userID int64) error
	GetMember(ctx context.Context, chatID, userID int64) (*ChatMember, error)
	GetMembers(ctx context.Context, chatID int64) ([]int64, error)
	
	CreateMessage(ctx context.Context, msg *Message) error
	GetMessageHistory(ctx context.Context, chatID int64, limit int) ([]Message, error)
	
	CreateReceipt(ctx context.Context, receipt *Receipt) error
	UpdateLastReadMessage(ctx context.Context, chatID, userID, msgID int64) error
	
	AddDeviceToken(ctx context.Context, token *DeviceToken) error
	GetDeviceTokens(ctx context.Context, userID int64) ([]string, error)
}
