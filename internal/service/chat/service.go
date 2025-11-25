package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ambarg/mini-telegram/internal/domain"
)

// Service handles chat business logic
type Service struct {
	chatRepo  domain.ChatRepository
	cacheRepo domain.CacheRepository
	broker    domain.MessageBroker
}

func NewService(chatRepo domain.ChatRepository, cacheRepo domain.CacheRepository, broker domain.MessageBroker) *Service {
	return &Service{
		chatRepo:  chatRepo,
		cacheRepo: cacheRepo,
		broker:    broker,
	}
}

func (s *Service) CreateChat(ctx context.Context, creatorID int64, reqType int16, memberIDs []int64) (*domain.Chat, error) {
	chat := &domain.Chat{Type: reqType}
	if err := s.chatRepo.CreateChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	// Add creator as admin
	allMembers := append([]int64{creatorID}, memberIDs...)
	
	for _, memberID := range allMembers {
		role := domain.RoleMember
		if memberID == creatorID {
			role = domain.RoleAdmin
		}
		
		member := &domain.ChatMember{
			ChatID: chat.ID,
			UserID: memberID,
			Role:   role,
		}
		if err := s.chatRepo.AddMember(ctx, member); err != nil {
			// Log error but continue? Or fail?
			// For now, fail
			return nil, fmt.Errorf("failed to add member %d: %w", memberID, err)
		}
	}

	// Cache members
	if err := s.cacheRepo.AddGroupMembers(ctx, chat.ID, allMembers); err != nil {
		// Log error
	}

	return chat, nil
}

func (s *Service) GetUserChats(ctx context.Context, userID int64) ([]domain.Chat, error) {
	return s.chatRepo.GetUserChats(ctx, userID)
}

func (s *Service) AddMember(ctx context.Context, chatID, userID int64) error {
	member := &domain.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   domain.RoleMember,
	}
	if err := s.chatRepo.AddMember(ctx, member); err != nil {
		return err
	}
	
	// Update cache
	return s.cacheRepo.AddGroupMembers(ctx, chatID, []int64{userID})
}

func (s *Service) RemoveMember(ctx context.Context, chatID, userID int64) error {
	if err := s.chatRepo.RemoveMember(ctx, chatID, userID); err != nil {
		return err
	}
	
	// Update cache
	return s.cacheRepo.RemoveGroupMember(ctx, chatID, userID)
}

func (s *Service) ProcessMessage(ctx context.Context, msg *domain.Message, clientUUID string) error {
	// 1. Persist message
	if err := s.chatRepo.CreateMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to persist message: %w", err)
	}

	// 2. Get members (from cache or DB)
	members, err := s.cacheRepo.GetGroupMembers(ctx, msg.ChatID)
	if err != nil || len(members) == 0 {
		members, err = s.chatRepo.GetMembers(ctx, msg.ChatID)
		if err != nil {
			return fmt.Errorf("failed to get chat members: %w", err)
		}
		// Cache them
		_ = s.cacheRepo.AddGroupMembers(ctx, msg.ChatID, members)
	}

	// 3. Create receipts
	for _, memberID := range members {
		receipt := &domain.Receipt{
			MsgID:  msg.ID,
			UserID: memberID,
			Status: domain.ReceiptStatusSent,
		}
		_ = s.chatRepo.CreateReceipt(ctx, receipt)
	}

	// 4. Publish delivery event
	deliveryPayload, _ := json.Marshal(map[string]interface{}{
		"type":      "Message",
		"msgId":     msg.ID,
		"chatId":    msg.ChatID,
		"userId":    msg.UserID,
		"body":      msg.Body,
		"createdAt": msg.CreatedAt.Unix(),
	})

	if err := s.broker.PublishToDeliveryExchange(ctx, msg.ChatID, deliveryPayload); err != nil {
		return fmt.Errorf("failed to publish delivery event: %w", err)
	}

	// 5. Send delivered acknowledgment back to sender
	if clientUUID != "" {
		deliveredPayload, _ := json.Marshal(map[string]interface{}{
			"type":  "Delivered",
			"uuid":  clientUUID,
			"msgId": msg.ID,
		})

		if err := s.broker.PublishToDeliveryExchange(ctx, msg.ChatID, deliveredPayload); err != nil {
			// Log warning but don't fail the whole process?
			// For now, return error to be safe
			return fmt.Errorf("failed to publish delivered ack: %w", err)
		}
	}

	return nil
}

func (s *Service) RegisterDevice(ctx context.Context, userID int64, token, platform string) error {
	deviceToken := &domain.DeviceToken{
		UserID:   userID,
		Token:    token,
		Platform: platform,
	}
	return s.chatRepo.AddDeviceToken(ctx, deviceToken)
}
