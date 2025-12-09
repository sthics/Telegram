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
func (r *UserRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]domain.User, error) {
	if query == "" {
		return []domain.User{}, nil
	}

	var daos []UserDAO
	// Search by email (partial match)
	err := r.db.WithContext(ctx).
		Where("email LIKE ?", "%"+query+"%").
		Limit(limit).
		Offset(offset).
		Find(&daos).Error
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, len(daos))
	for i, dao := range daos {
		users[i] = *dao.ToDomain()
	}
	return users, nil
}

// ChatRepository implementation
type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *DB) *ChatRepository {
	return &ChatRepository{db: db.DB}
}

func (r *ChatRepository) CreateChat(ctx context.Context, chat *domain.Chat, memberIDs []int64) (*domain.Chat, error) {
	dao := FromDomainChat(chat)
	if err := r.db.WithContext(ctx).Create(dao).Error; err != nil {
		return nil, err
	}
	chat.ID = dao.ID
	chat.CreatedAt = dao.CreatedAt
	return chat, nil
}

func (r *ChatRepository) UpdateChat(ctx context.Context, chat *domain.Chat) error {
	dao := FromDomainChat(chat)
	return r.db.WithContext(ctx).Model(dao).Updates(dao).Error
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
		Table("chats").
		Select("chats.*, (SELECT COUNT(*) FROM messages WHERE messages.chat_id = chats.id AND messages.id > chat_members.last_read_msg_id AND messages.user_id != chat_members.user_id) as unread_count").
		Joins("JOIN chat_members ON chat_members.chat_id = chats.id").
		Where("chat_members.user_id = ?", userID).
		Find(&daos).Error; err != nil {
		return nil, err
	}
	
	chats := make([]domain.Chat, len(daos))
	for i, dao := range daos {
		chats[i] = *dao.ToDomain()
		
		// Fetch last message
		var msgDAO MessageDAO
		if err := r.db.WithContext(ctx).
			Where("chat_id = ?", dao.ID).
			Order("created_at DESC").
			Limit(1).
			Find(&msgDAO).Error; err == nil && msgDAO.ID != 0 {
			chats[i].LastMessage = msgDAO.ToDomain()
			// Populate User for LastMessage if needed? 
			// Frontend uses `lastMessage.body` and `created_at`. User not strictly needed for preview unless we show "Name: Body".
			// Current UI just shows Body.
		}
	}
	return chats, nil
}

func (r *ChatRepository) AddMember(ctx context.Context, chatID, userID int64, role domain.Role) error {
	dao := &ChatMemberDAO{
		ChatID: chatID,
		UserID: userID,
		Role:   string(role),
	}
	return r.db.WithContext(ctx).Create(dao).Error
}

func (r *ChatRepository) UpdateMemberRole(ctx context.Context, chatID, userID int64, role domain.Role) error {
	return r.db.WithContext(ctx).
		Model(&ChatMemberDAO{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Update("role", string(role)).Error
}

func (r *ChatRepository) RemoveMember(ctx context.Context, chatID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Delete(&ChatMemberDAO{}).Error
}

func (r *ChatRepository) GetChatMembers(ctx context.Context, chatID int64) ([]domain.ChatMember, error) {
	var daos []ChatMemberDAO
	if err := r.db.WithContext(ctx).Preload("User").Where("chat_id = ?", chatID).Find(&daos).Error; err != nil {
		return nil, err
	}

	members := make([]domain.ChatMember, len(daos))
	for i, dao := range daos {
		members[i] = *dao.ToDomain()
	}
	return members, nil
}

func (r *ChatRepository) IsMember(ctx context.Context, chatID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ChatMemberDAO{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ChatRepository) GetMemberRole(ctx context.Context, chatID, userID int64) (domain.Role, error) {
	var role string
	err := r.db.WithContext(ctx).
		Model(&ChatMemberDAO{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Pluck("role", &role).Error
	if err != nil {
		return "", err
	}
	return domain.Role(role), nil
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
		Where("chat_id = ? AND user_id = ? AND last_read_msg_id < ?", chatID, userID, msgID).
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

func (r *ChatRepository) GetPrivateChatBetweenUsers(ctx context.Context, userA, userB int64) (*domain.Chat, error) {
	var dao ChatDAO
	// Find a chat of type 1 (Direct) that has both members
	// This query assumes only 2 members in Type 1 chat, or at least matching these two
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT c.* FROM chats c
			JOIN chat_members cm1 ON c.id = cm1.chat_id
			JOIN chat_members cm2 ON c.id = cm2.chat_id
			WHERE c.type = ? AND cm1.user_id = ? AND cm2.user_id = ?
			LIMIT 1
		`, domain.ChatTypeDirect, userA, userB).
		Scan(&dao).Error

	if err != nil {
		return nil, err
	}
	if dao.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return dao.ToDomain(), nil
}
