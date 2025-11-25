package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/ambarg/mini-telegram/internal/auth"
	"github.com/ambarg/mini-telegram/internal/domain"
)

// Service handles authentication logic
type Service struct {
	userRepo    domain.UserRepository
	authService *auth.Service // Utility service for JWT/Hashing
}

func NewService(userRepo domain.UserRepository, authService *auth.Service) *Service {
	return &Service{
		userRepo:    userRepo,
		authService: authService,
	}
}

type RegisterInput struct {
	Email    string
	Password string
}

type TokenResponse struct {
	UserID       int64
	AccessToken  string
	RefreshToken string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*TokenResponse, error) {
	// Hash password
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		Email:        input.Email,
		PasswordHash: passwordHash,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	return s.generateTokens(user.ID)
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := auth.VerifyPassword(password, user.PasswordHash); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.generateTokens(user.ID)
}

func (s *Service) RefreshToken(refreshToken string) (string, error) {
	claims, err := s.authService.ValidateToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	userID, err := auth.ExtractUserID(claims)
	if err != nil {
		return "", errors.New("invalid user ID")
	}

	accessToken, err := s.authService.GenerateAccessToken(userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

func (s *Service) generateTokens(userID int64) (*TokenResponse, error) {
	accessToken, err := s.authService.GenerateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.authService.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
