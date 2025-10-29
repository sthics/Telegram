package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "ValidPass123",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "minimum length",
			password: "12345678",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				// Verify the hash works
				err = VerifyPassword(tt.password, hash)
				assert.NoError(t, err)

				// Verify wrong password fails
				err = VerifyPassword("wrongpassword", hash)
				assert.Error(t, err)
			}
		})
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	// Generate a test private key
	privateKey, err := GeneratePrivateKey()
	require.NoError(t, err)

	service := NewService(privateKey)

	// Generate access token
	userID := int64(12345)
	token, err := service.GenerateAccessToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.NotNil(t, claims)

	// Extract user ID
	extractedUserID, err := ExtractUserID(claims)
	require.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestValidateToken_Invalid(t *testing.T) {
	privateKey, err := GeneratePrivateKey()
	require.NoError(t, err)

	service := NewService(privateKey)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "not.a.valid.token",
		},
		{
			name:  "random string",
			token: "abcdefghijklmnop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	privateKey, err := GeneratePrivateKey()
	require.NoError(t, err)

	service := NewService(privateKey)

	userID := int64(67890)
	refreshToken, err := service.GenerateRefreshToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Validate refresh token
	claims, err := service.ValidateToken(refreshToken)
	require.NoError(t, err)

	extractedUserID, err := ExtractUserID(claims)
	require.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}
