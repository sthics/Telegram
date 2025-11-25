package push

import (
	"context"
	"encoding/json"

	"github.com/ambarg/mini-telegram/internal/domain"
	"github.com/rs/zerolog/log"
)

// Service handles push notifications
type Service struct {
	chatRepo  domain.ChatRepository
	cacheRepo domain.CacheRepository
}

// NewService creates a new push service
func NewService(chatRepo domain.ChatRepository, cacheRepo domain.CacheRepository) *Service {
	return &Service{
		chatRepo:  chatRepo,
		cacheRepo: cacheRepo,
	}
}

// ProcessPushNotification handles a push notification request
func (s *Service) ProcessPushNotification(ctx context.Context, payload []byte) error {
	var msg map[string]any
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}

	chatID, _ := msg["chatId"].(float64)
	senderID, _ := msg["userId"].(float64)
	body, _ := msg["body"].(string)

	// Get chat members
	members, err := s.chatRepo.GetMembers(ctx, int64(chatID))
	if err != nil {
		return err
	}

	log.Info().Int64("chat_id", int64(chatID)).Msg("Processing message for push")

	for _, memberID := range members {
		// Skip sender
		if memberID == int64(senderID) {
			continue
		}

		// Check presence
		online, _, err := s.cacheRepo.GetPresence(ctx, memberID)
		if err != nil {
			log.Error().Err(err).Int64("user_id", memberID).Msg("failed to check presence")
			continue
		}

		log.Info().Int64("user_id", memberID).Bool("online", online).Msg("User presence check")

		if !online {
			// User is offline, send push
			tokens, err := s.chatRepo.GetDeviceTokens(ctx, memberID)
			if err != nil {
				log.Error().Err(err).Int64("user_id", memberID).Msg("failed to get device tokens")
				continue
			}

			log.Info().Int64("user_id", memberID).Int("token_count", len(tokens)).Msg("Found device tokens")

			for _, token := range tokens {
				// In a real implementation, we would call APNS/FCM here
				log.Info().
					Int64("user_id", memberID).
					Str("token", token).
					Str("body", body).
					Msg("Sending push notification")
			}
		}
	}

	return nil
}
