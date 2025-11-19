package main

import (
	"testing"

	"github.com/ambarg/mini-telegram/internal/database"
	"github.com/ambarg/mini-telegram/internal/rabbitmq"
	"github.com/ambarg/mini-telegram/internal/redis"
	"github.com/stretchr/testify/assert"
)

func TestNewChatService(t *testing.T) {
	// This is a basic test to ensure the service can be instantiated
	// In a real scenario, we would mock the dependencies
	db := &database.DB{}
	redisClient := &redis.Client{}
	rmqClient := &rabbitmq.Client{}

	svc := NewChatService(db, redisClient, rmqClient)
	assert.NotNil(t, svc)
	assert.Equal(t, db, svc.db)
	assert.Equal(t, redisClient, svc.redis)
	assert.Equal(t, rmqClient, svc.rabbitmq)
}
