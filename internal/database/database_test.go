package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	user := &User{
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "hashed_password", user.PasswordHash)
}

func TestChatModel(t *testing.T) {
	chat := &Chat{
		Type: ChatTypeDirect,
	}

	assert.Equal(t, int16(ChatTypeDirect), chat.Type)
}

func TestMessageModel(t *testing.T) {
	replyTo := int64(100)
	msg := &Message{
		ChatID:    1,
		UserID:    2,
		Body:      "Hello, World!",
		MediaURL:  "http://example.com/image.jpg",
		ReplyToID: &replyTo,
		Reactions: []byte(`{"like": 1}`),
	}

	assert.Equal(t, int64(1), msg.ChatID)
	assert.Equal(t, int64(2), msg.UserID)
	assert.Equal(t, "Hello, World!", msg.Body)
	assert.Equal(t, "http://example.com/image.jpg", msg.MediaURL)
	assert.Equal(t, int64(100), *msg.ReplyToID)
	assert.JSONEq(t, `{"like": 1}`, string(msg.Reactions))
}

func TestReceiptConstants(t *testing.T) {
	assert.Equal(t, int16(1), int16(ReceiptStatusSent))
	assert.Equal(t, int16(2), int16(ReceiptStatusDelivered))
	assert.Equal(t, int16(3), int16(ReceiptStatusRead))
}

func TestChatTypeConstants(t *testing.T) {
	assert.Equal(t, int16(1), int16(ChatTypeDirect))
	assert.Equal(t, int16(2), int16(ChatTypeGroup))
}

func TestChatMemberRole(t *testing.T) {
	member := &ChatMember{
		ChatID: 1,
		UserID: 1,
		Role:   RoleAdmin,
	}
	assert.Equal(t, RoleAdmin, member.Role)
	
	member.Role = RoleMember
	assert.Equal(t, RoleMember, member.Role)
}

// Integration tests would go here using testcontainers
// Example structure (requires testcontainers setup):
/*
func TestDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
	)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	// Get connection string
	dsn, err := pgContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Connect to database
	db, err := New(Config{
		DSN:             dsn,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	})
	require.NoError(t, err)
	defer db.Close()

	// Run auto-migration
	err = db.AutoMigrate()
	require.NoError(t, err)

	// Test creating a user
	user := &User{
		Email:        "integration@test.com",
		PasswordHash: "hashed",
	}
	err = db.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)

	// Test retrieving the user
	retrieved, err := db.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, retrieved.Email)
}
*/
