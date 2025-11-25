package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/ambarg/mini-telegram/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
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

	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("failed to use tracing plugin: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// UserRepository implementation
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db.DB}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	dao := FromDomainUser(user)
	if err := r.db.WithContext(ctx).Create(dao).Error; err != nil {
		return err
	}
	user.ID = dao.ID
	user.CreatedAt = dao.CreatedAt
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var dao UserDAO
	if err := r.db.WithContext(ctx).First(&dao, id).Error; err != nil {
		return nil, err
	}
	return dao.ToDomain(), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var dao UserDAO
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&dao).Error; err != nil {
		return nil, err
	}
	return dao.ToDomain(), nil
}

// ChatRepository implementation
type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *DB) *ChatRepository {
	return &ChatRepository{db: db.DB}
}

func (r *ChatRepository) CreateChat(ctx context.Context, chat *domain.Chat) error {
	dao := FromDomainChat(chat)
	if err := r.db.WithContext(ctx).Create(dao).Error; err != nil {
		return err
	}
	chat.ID = dao.ID
	chat.CreatedAt = dao.CreatedAt
	return nil
}

func (r *ChatRepository) GetChat(ctx context.Context, id int64) (*domain.Chat, error) {
	var dao ChatDAO
	if err := r.db.WithContext(ctx).First(&dao, id).Error; err != nil {
		return nil, err
	}
	return dao.ToDomain(), nil
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userID int64) ([]domain.Chat, error) {
	var daos []ChatDAO
	if err := r.db.WithContext(ctx).
		Joins("JOIN chat_members ON chat_members.chat_id = chats.id").
		Where("chat_members.user_id = ?", userID).
		Find(&daos).Error; err != nil {
		return nil, err
	}
	
	chats := make([]domain.Chat, len(daos))
	for i, dao := range daos {
		chats[i] = *dao.ToDomain()
	}
	return chats, nil
}

func (r *ChatRepository) AddMember(ctx context.Context, member *domain.ChatMember) error {
	dao := FromDomainChatMember(member)
	return r.db.WithContext(ctx).Create(dao).Error
}

func (r *ChatRepository) RemoveMember(ctx context.Context, chatID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Delete(&ChatMemberDAO{}).Error
}

func (r *ChatRepository) GetMember(ctx context.Context, chatID, userID int64) (*domain.ChatMember, error) {
	var dao ChatMemberDAO
	if err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&dao).Error; err != nil {
		return nil, err
	}
	return dao.ToDomain(), nil
}

func (r *ChatRepository) GetMembers(ctx context.Context, chatID int64) ([]int64, error) {
	var members []ChatMemberDAO
	if err := r.db.WithContext(ctx).Where("chat_id = ?", chatID).Find(&members).Error; err != nil {
		return nil, err
	}

	userIDs := make([]int64, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}
	return userIDs, nil
}

func (r *ChatRepository) CreateMessage(ctx context.Context, msg *domain.Message) error {
	dao := FromDomainMessage(msg)
	if err := r.db.WithContext(ctx).Create(dao).Error; err != nil {
		return err
	}
	msg.ID = dao.ID
	msg.CreatedAt = dao.CreatedAt
	return nil
}

func (r *ChatRepository) GetMessageHistory(ctx context.Context, chatID int64, limit int) ([]domain.Message, error) {
	var daos []MessageDAO
	if err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("id DESC").
		Limit(limit).
		Find(&daos).Error; err != nil {
		return nil, err
	}
	
	msgs := make([]domain.Message, len(daos))
	for i, dao := range daos {
		msgs[i] = *dao.ToDomain()
	}
	return msgs, nil
}

func (r *ChatRepository) CreateReceipt(ctx context.Context, receipt *domain.Receipt) error {
	dao := FromDomainReceipt(receipt)
	return r.db.WithContext(ctx).Create(dao).Error
}

func (r *ChatRepository) UpdateLastReadMessage(ctx context.Context, chatID, userID, msgID int64) error {
	return r.db.WithContext(ctx).
		Model(&ChatMemberDAO{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Update("last_read_msg_id", msgID).Error
}

func (r *ChatRepository) AddDeviceToken(ctx context.Context, token *domain.DeviceToken) error {
	dao := FromDomainDeviceToken(token)
	return r.db.WithContext(ctx).Save(dao).Error
}

func (r *ChatRepository) GetDeviceTokens(ctx context.Context, userID int64) ([]string, error) {
	var tokens []string
	err := r.db.WithContext(ctx).
		Model(&DeviceTokenDAO{}).
		Where("user_id = ?", userID).
		Pluck("token", &tokens).Error
	return tokens, err
}
