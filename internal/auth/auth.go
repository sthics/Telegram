package auth

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	// AccessTokenLifetime is 15 minutes
	AccessTokenLifetime = 15 * time.Minute
	// RefreshTokenLifetime is 7 days
	RefreshTokenLifetime = 7 * 24 * time.Hour
	// BcryptCost is the bcrypt cost factor (≈250ms on 2GHz core)
	BcryptCost = 12
	// Issuer is the JWT issuer
	Issuer = "minitelegram"
)

// Claims represents JWT claims
type Claims struct {
	jwt.RegisteredClaims
}

// Service handles authentication
type Service struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

// NewService creates a new authentication service
func NewService(privateKey *ecdsa.PrivateKey) *Service {
	return &Service{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateAccessToken generates a JWT access token
func (s *Service) GenerateAccessToken(userID int64) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenLifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a JWT refresh token
func (s *Service) GenerateRefreshToken(userID int64) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenLifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
}

// ExtractUserID extracts user ID from claims
func ExtractUserID(claims *Claims) (int64, error) {
	var userID int64
	_, err := fmt.Sscanf(claims.Subject, "%d", &userID)
	if err != nil {
		return 0, fmt.Errorf("failed to parse user ID from subject: %w", err)
	}
	return userID, nil
}
