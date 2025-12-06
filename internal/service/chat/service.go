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

func (s *Service) CreateChat(ctx context.Context, creatorID int64, reqType int16, memberIDs []int64, title string) (*domain.Chat, error) {
	// If private chat, check if exists
	if reqType == domain.ChatTypeDirect && len(memberIDs) == 1 {
		existing, err := s.chatRepo.GetPrivateChatBetweenUsers(ctx, creatorID, memberIDs[0])
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	chat := &domain.Chat{Type: reqType, Title: title}
	var err error
	chat, err = s.chatRepo.CreateChat(ctx, chat, memberIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	// Add creator as admin
	allMembers := append([]int64{creatorID}, memberIDs...)
	
	for _, memberID := range allMembers {
		role := domain.RoleMember
		if memberID == creatorID {
			role = domain.RoleAdmin
		}
		
		if err := s.chatRepo.AddMember(ctx, chat.ID, memberID, role); err != nil {

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
	chats, err := s.chatRepo.GetUserChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range chats {
		if chats[i].Type == domain.ChatTypeGroup {
			chats[i].Name = chats[i].Title
		} else {
			// Private chat: Find other member
			members, err := s.chatRepo.GetChatMembers(ctx, chats[i].ID)
			if err == nil {
				for _, m := range members {
					if m.UserID != userID && m.User != nil {
						chats[i].Name = m.User.Email
						// Check presence
						online, _, _ := s.cacheRepo.GetPresence(ctx, m.UserID)
						chats[i].Online = online
						break
					}
				}
				// If no other member found (e.g. self chat or other left), default to "Saved Messages" or similar?
				// Or leave empty, frontend handles "Unknown".
				if chats[i].Name == "" {
					// Fallback to searching specifically for the other ID if GetChatMembers didn't load User?
					// But we updated GetChatMembers to Preload User.
					// If strictly self-chat, it might be empty.
				}
			}
		}
	}
	return chats, nil
}

func (s *Service) GetMessages(ctx context.Context, chatID, userID int64, limit int) ([]domain.Message, error) {
	// Check membership
	isMember, err := s.chatRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("permission denied: user is not a member of this chat")
	}

	messages, err := s.chatRepo.GetMessageHistory(ctx, chatID, limit)
	if err != nil {
		return nil, err
	}

	// Calculate status
	// Get all chat members to check LastReadMsgID
	members, err := s.chatRepo.GetChatMembers(ctx, chatID)
	if err == nil {
		var maxReadID int64
		for _, m := range members {
			if m.UserID != userID && m.LastReadMsgID > maxReadID {
				maxReadID = m.LastReadMsgID
			}
		}

		for i := range messages {
			if messages[i].UserID == userID { // Only for my messages
				if messages[i].ID <= maxReadID {
					messages[i].Status = 3 // Read
				} else {
					messages[i].Status = 1 // Sent
				}
			}
		}
	}

	return messages, nil
}

func (s *Service) AddMember(ctx context.Context, chatID, userID int64) error {
	if err := s.chatRepo.AddMember(ctx, chatID, userID, domain.RoleMember); err != nil {
		return err
	}
	
	// Update cache
	return s.cacheRepo.AddGroupMembers(ctx, chatID, []int64{userID})
}

func (s *Service) RemoveMember(ctx context.Context, chatID, userID int64) error {
	// TODO: Add permission check if caller is not userID (i.e. kick vs leave)
	
	if err := s.chatRepo.RemoveMember(ctx, chatID, userID); err != nil {
		return err
	}
	
	// Update cache
	return s.cacheRepo.RemoveGroupMember(ctx, chatID, userID)
}

func (s *Service) UpdateGroupInfo(ctx context.Context, chatID, actorID int64, title string) error {
	isAdmin, err := s.isAdmin(ctx, chatID, actorID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return fmt.Errorf("permission denied: only admins can update group info")
	}

	chat, err := s.chatRepo.GetChat(ctx, chatID)
	if err != nil {
		return err
	}

	chat.Title = title
	return s.chatRepo.UpdateChat(ctx, chat)
}

func (s *Service) PromoteMember(ctx context.Context, chatID, actorID, targetID int64) error {
	isAdmin, err := s.isAdmin(ctx, chatID, actorID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return fmt.Errorf("permission denied: only admins can promote members")
	}

	return s.chatRepo.UpdateMemberRole(ctx, chatID, targetID, domain.RoleAdmin)
}

func (s *Service) DemoteMember(ctx context.Context, chatID, actorID, targetID int64) error {
	isAdmin, err := s.isAdmin(ctx, chatID, actorID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return fmt.Errorf("permission denied: only admins can demote members")
	}

	// Prevent demoting self? Or allow it? Allowing it for now.
	return s.chatRepo.UpdateMemberRole(ctx, chatID, targetID, domain.RoleMember)
}

func (s *Service) MarkChatRead(ctx context.Context, chatID, userID, msgID int64) error {
	// Update last_read_msg_id
	if err := s.chatRepo.UpdateLastReadMessage(ctx, chatID, userID, msgID); err != nil {
		return err
	}
	
	// Broadcast Read Event to chat so senders can update ticks?
	// For now, simpler to just update DB. Real-time ticks require broadcasting event.
	// Let's broadcast "ReadReceipt" event
	payload, _ := json.Marshal(map[string]interface{}{
		"type":       "Read",
		"chat_id":    chatID,
		"user_id":    userID,
		"max_id":     msgID,
	})
	return s.broker.PublishToDeliveryExchange(ctx, chatID, payload)
}

func (s *Service) isAdmin(ctx context.Context, chatID, userID int64) (bool, error) {
	role, err := s.chatRepo.GetMemberRole(ctx, chatID, userID)
	if err != nil {
		return false, err
	}
	return role == domain.RoleAdmin, nil
}

func (s *Service) ProcessMessage(ctx context.Context, msg *domain.Message, clientUUID string) error {
	// 1. Persist message
	if err := s.chatRepo.CreateMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to persist message: %w", err)
	}

	// 2. Get members (from cache or DB)
	members, err := s.cacheRepo.GetGroupMembers(ctx, msg.ChatID)
	if err != nil || len(members) == 0 {
		chatMembers, err := s.chatRepo.GetChatMembers(ctx, msg.ChatID)
		if err != nil {
			return fmt.Errorf("failed to get chat members: %w", err)
		}
		
		members = make([]int64, len(chatMembers))
		for i, m := range chatMembers {
			members[i] = m.UserID
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
		"type":       "Message",
		"id":         msg.ID,
		"chat_id":    msg.ChatID,
		"user_id":    msg.UserID,
		"body":       msg.Body,
		"created_at": msg.CreatedAt, // Serializes to ISO string by default
	})

	if err := s.broker.PublishToDeliveryExchange(ctx, msg.ChatID, deliveryPayload); err != nil {
		return fmt.Errorf("failed to publish delivery event: %w", err)
	}

	// 5. Send delivered acknowledgment back to sender
	if clientUUID != "" {
		deliveredPayload, _ := json.Marshal(map[string]interface{}{
			"type":   "Delivered",
			"uuid":   clientUUID,
			"msg_id": msg.ID,
		})

		if err := s.broker.PublishToDeliveryExchange(ctx, msg.ChatID, deliveredPayload); err != nil {

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
func (s *Service) GetChatMembers(ctx context.Context, chatID, userID int64) ([]domain.ChatMember, error) {
	// Check membership
	isMember, err := s.chatRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("permission denied: user is not a member of this chat")
	}

	return s.chatRepo.GetChatMembers(ctx, chatID)
}

func (s *Service) IsMember(ctx context.Context, chatID, userID int64) (bool, error) {
	return s.chatRepo.IsMember(ctx, chatID, userID)
}
