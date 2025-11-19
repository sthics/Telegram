package database

import (
	"time"
)

// User represents a registered user
type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
}

// Chat represents a chat room (direct or group)
type Chat struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Type      int16     `gorm:"not null;check:type IN (1,2)" json:"type"` // 1=direct, 2=group
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
}

// ChatMember represents membership in a chat
// Role represents a user's role in a chat
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// ChatMember represents membership in a chat
type ChatMember struct {
	ChatID        int64     `gorm:"primaryKey" json:"chat_id"`
	UserID        int64     `gorm:"primaryKey" json:"user_id"`
	Role          Role      `gorm:"default:'member'" json:"role"`
	LastReadMsgID int64     `gorm:"default:0" json:"last_read_msg_id"`
	JoinedAt      time.Time `gorm:"default:now()" json:"joined_at"`
}

// Message represents a chat message
type Message struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	ChatID    int64     `gorm:"not null;index:idx_messages_chat_created" json:"chat_id"`
	UserID    int64     `gorm:"not null" json:"user_id"`
	Body      string    `gorm:"not null" json:"body"`
	MediaURL  string    `json:"media_url,omitempty"`
	ReplyToID *int64    `json:"reply_to_id,omitempty"`
	Reactions []byte    `gorm:"type:jsonb;default:'{}'" json:"reactions"`
	CreatedAt time.Time `gorm:"default:now();index:idx_messages_chat_created" json:"created_at"`
}

// Receipt represents message delivery/read status
type Receipt struct {
	MsgID  int64     `gorm:"primaryKey" json:"msg_id"`
	UserID int64     `gorm:"primaryKey" json:"user_id"`
	Status int16     `gorm:"not null;check:status IN (1,2,3)" json:"status"` // 1=sent, 2=delivered, 3=read
	Ts     time.Time `gorm:"default:now()" json:"ts"`
}

// Receipt status constants
const (
	ReceiptStatusSent      = 1
	ReceiptStatusDelivered = 2
	ReceiptStatusRead      = 3
)

// Chat type constants
const (
	ChatTypeDirect = 1
	ChatTypeGroup  = 2
)
