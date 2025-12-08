package postgres

import (
	"time"

	"github.com/ambarg/mini-telegram/internal/domain"
)

// UserDAO represents a registered user in the database
type UserDAO struct {
	ID           int64     `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"default:now()"`
}

func (u *UserDAO) ToDomain() *domain.User {
	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}

func FromDomainUser(u *domain.User) *UserDAO {
	return &UserDAO{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}

// ChatDAO represents a chat room
type ChatDAO struct {
	ID        int64     `gorm:"primaryKey"`
	Type      int16     `gorm:"not null;check:type IN (1,2)"`
	Title     string    `gorm:"size:255"`
	CreatedAt time.Time `gorm:"default:now()"`
	UnreadCount int64   `gorm:"->;column:unread_count"`
}

func (c *ChatDAO) ToDomain() *domain.Chat {
	return &domain.Chat{
		ID:          c.ID,
		Type:        c.Type,
		Title:       c.Title,
		CreatedAt:   c.CreatedAt,
		UnreadCount: c.UnreadCount,
	}
}

func FromDomainChat(c *domain.Chat) *ChatDAO {
	return &ChatDAO{
		ID:        c.ID,
		Type:      c.Type,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
	}
}

// ChatMemberDAO represents membership in a chat
type ChatMemberDAO struct {
	ChatID        int64     `gorm:"primaryKey"`
	UserID        int64     `gorm:"primaryKey"`
	Role          string    `gorm:"default:'member'"`
	LastReadMsgID int64     `gorm:"default:0"`
	JoinedAt      time.Time `gorm:"default:now()"`
	User          UserDAO   `gorm:"foreignKey:UserID"`
}

func (m *ChatMemberDAO) ToDomain() *domain.ChatMember {
	dm := &domain.ChatMember{
		ChatID:        m.ChatID,
		UserID:        m.UserID,
		Role:          domain.Role(m.Role),
		LastReadMsgID: m.LastReadMsgID,
		JoinedAt:      m.JoinedAt,
	}
	if m.User.ID != 0 {
		dm.User = m.User.ToDomain()
	}
	return dm
}

func FromDomainChatMember(m *domain.ChatMember) *ChatMemberDAO {
	return &ChatMemberDAO{
		ChatID:        m.ChatID,
		UserID:        m.UserID,
		Role:          string(m.Role),
		LastReadMsgID: m.LastReadMsgID,
		JoinedAt:      m.JoinedAt,
	}
}

// MessageDAO represents a chat message
type MessageDAO struct {
	ID        int64     `gorm:"primaryKey"`
	ChatID    int64     `gorm:"not null;index:idx_messages_chat_created"`
	UserID    int64     `gorm:"not null"`
	Body      string    `gorm:"not null"`
	MediaURL  string    ``
	ReplyToID *int64    ``
	Reactions []byte    `gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time `gorm:"default:now();index:idx_messages_chat_created"`
}

func (m *MessageDAO) ToDomain() *domain.Message {
	return &domain.Message{
		ID:        m.ID,
		ChatID:    m.ChatID,
		UserID:    m.UserID,
		Body:      m.Body,
		MediaURL:  m.MediaURL,
		ReplyToID: m.ReplyToID,
		Reactions: m.Reactions,
		CreatedAt: m.CreatedAt,
	}
}

func FromDomainMessage(m *domain.Message) *MessageDAO {
	return &MessageDAO{
		ID:        m.ID,
		ChatID:    m.ChatID,
		UserID:    m.UserID,
		Body:      m.Body,
		MediaURL:  m.MediaURL,
		ReplyToID: m.ReplyToID,
		Reactions: m.Reactions,
		CreatedAt: m.CreatedAt,
	}
}

// ReceiptDAO represents message delivery/read status
type ReceiptDAO struct {
	MsgID  int64     `gorm:"primaryKey"`
	UserID int64     `gorm:"primaryKey"`
	Status int16     `gorm:"not null;check:status IN (1,2,3)"`
	CreatedAt     time.Time `gorm:"default:now()"`
}

func (r *ReceiptDAO) ToDomain() *domain.Receipt {
	return &domain.Receipt{
		MsgID:  r.MsgID,
		UserID: r.UserID,
		Status: r.Status,
		Ts:     r.CreatedAt,
	}
}

func FromDomainReceipt(r *domain.Receipt) *ReceiptDAO {
	return &ReceiptDAO{
		MsgID:  r.MsgID,
		UserID: r.UserID,
		Status: r.Status,
		CreatedAt:     r.Ts,
	}
}

// DeviceTokenDAO represents a user's device for push notifications
type DeviceTokenDAO struct {
	UserID    int64     `gorm:"primaryKey"`
	Token     string    `gorm:"primaryKey"`
	Platform  string    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"default:now()"`
}

func (d *DeviceTokenDAO) ToDomain() *domain.DeviceToken {
	return &domain.DeviceToken{
		UserID:    d.UserID,
		Token:     d.Token,
		Platform:  d.Platform,
		UpdatedAt: d.UpdatedAt,
	}
}

func FromDomainDeviceToken(d *domain.DeviceToken) *DeviceTokenDAO {
	return &DeviceTokenDAO{
		UserID:    d.UserID,
		Token:     d.Token,
		Platform:  d.Platform,
		UpdatedAt: d.UpdatedAt,
	}
}

// TableName overrides
func (UserDAO) TableName() string        { return "users" }
func (ChatDAO) TableName() string        { return "chats" }
func (ChatMemberDAO) TableName() string  { return "chat_members" }
func (MessageDAO) TableName() string     { return "messages" }
func (ReceiptDAO) TableName() string     { return "receipts" }
func (DeviceTokenDAO) TableName() string { return "device_tokens" }
