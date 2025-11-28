package media

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ambarg/mini-telegram/internal/domain"
	"github.com/google/uuid"
)

type Service struct {
	repo domain.MediaRepository
}

func NewService(repo domain.MediaRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUploadURL(ctx context.Context, userID int64, filename string, contentType string) (string, string, error) {
	// Generate unique object name: uploads/{userID}/{uuid}{ext}
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", "", fmt.Errorf("filename must have an extension")
	}

	objectName := fmt.Sprintf("uploads/%d/%s%s", userID, uuid.New().String(), ext)

	// Generate presigned URL (valid for 15 minutes)
	url, err := s.repo.GeneratePresignedURL(ctx, objectName, contentType, 900)
	if err != nil {
		return "", "", err
	}

	return url, objectName, nil
}
