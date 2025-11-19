package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps gorm.DB
type DB struct {
	*gorm.DB
}

// Config holds database configuration
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// New creates a new database connection
func New(cfg Config) (*DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &DB{db}, nil
}

// AutoMigrate runs GORM auto-migration (dev only)
func (db *DB) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&User{},
		&Chat{},
		&ChatMember{},
		&Message{},
		&Receipt{},
	)
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// WithContext returns a new DB with context
func (db *DB) WithContext(ctx context.Context) *DB {
	return &DB{db.DB.WithContext(ctx)}
}

// GetUser retrieves a user by ID
func (db *DB) GetUser(ctx context.Context, userID int64) (*User, error) {
	var user User
	if err := db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// CreateUser creates a new user
func (db *DB) CreateUser(ctx context.Context, user *User) error {
	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// CreateMessage inserts a new message
func (db *DB) CreateMessage(ctx context.Context, msg *Message) error {
	if err := db.WithContext(ctx).Create(msg).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

// GetChatMembers retrieves all members of a chat
func (db *DB) GetChatMembers(ctx context.Context, chatID int64) ([]int64, error) {
	var members []ChatMember
	if err := db.WithContext(ctx).Where("chat_id = ?", chatID).Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to get chat members: %w", err)
	}

	userIDs := make([]int64, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}
	return userIDs, nil
}

// GetMessageHistory retrieves paginated message history
func (db *DB) GetMessageHistory(ctx context.Context, chatID int64, limit int) ([]Message, error) {
	var messages []Message
	if err := db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("id DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}
	return messages, nil
}

// CreateReceipt creates a message receipt
func (db *DB) CreateReceipt(ctx context.Context, receipt *Receipt) error {
	if err := db.WithContext(ctx).Create(receipt).Error; err != nil {
		return fmt.Errorf("failed to create receipt: %w", err)
	}
	return nil
}

// UpdateLastReadMessage updates the last read message for a user in a chat
func (db *DB) UpdateLastReadMessage(ctx context.Context, chatID, userID, msgID int64) error {
	if err := db.WithContext(ctx).
		Model(&ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Update("last_read_msg_id", msgID).Error; err != nil {
		return fmt.Errorf("failed to update last read message: %w", err)
	}
	return nil
}

// CreateChat creates a new chat
func (db *DB) CreateChat(ctx context.Context, chat *Chat) error {
	if err := db.WithContext(ctx).Create(chat).Error; err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}
	return nil
}

// AddChatMember adds a member to a chat
func (db *DB) AddChatMember(ctx context.Context, member *ChatMember) error {
	if err := db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add chat member: %w", err)
	}
	return nil
}

// RemoveChatMember removes a member from a chat
func (db *DB) RemoveChatMember(ctx context.Context, chatID, userID int64) error {
	if err := db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Delete(&ChatMember{}).Error; err != nil {
		return fmt.Errorf("failed to remove chat member: %w", err)
	}
	return nil
}

// GetChatMember retrieves a specific chat member
func (db *DB) GetChatMember(ctx context.Context, chatID, userID int64) (*ChatMember, error) {
	var member ChatMember
	if err := db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to get chat member: %w", err)
	}
	return &member, nil
}

// GetUserChats retrieves all chats for a user
func (db *DB) GetUserChats(ctx context.Context, userID int64) ([]Chat, error) {
	var chats []Chat
	if err := db.WithContext(ctx).
		Joins("JOIN chat_members ON chat_members.chat_id = chats.id").
		Where("chat_members.user_id = ?", userID).
		Find(&chats).Error; err != nil {
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}
	return chats, nil
}
