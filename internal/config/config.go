package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds application configuration
type Config struct {
	// Server
	GinMode string `envconfig:"GIN_MODE" default:"release"`
	Port    int    `envconfig:"PORT" default:"8080"`

	// Database
	DSN             string        `envconfig:"DSN" required:"true"`
	MaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`

	// Redis
	RedisAddr     string `envconfig:"REDIS_ADDR" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`

	// RabbitMQ
	AMQPURL string `envconfig:"AMQP_URL" required:"true"`

	// JWT
	JWTPrivateKeyPath string `envconfig:"JWT_PRIVATE_KEY_PATH" required:"true"`

	// Timeouts
	RedisTimeout    time.Duration `envconfig:"REDIS_TIMEOUT" default:"2s"`
	PostgresTimeout time.Duration `envconfig:"POSTGRES_TIMEOUT" default:"5s"`

	// Connection Registry
	ConnTTL      time.Duration `envconfig:"CONN_TTL" default:"35s"`
	PingInterval time.Duration `envconfig:"PING_INTERVAL" default:"30s"`

	// Observability
	OtelCollectorURL string `envconfig:"OTEL_COLLECTOR_URL" default:"localhost:4317"`

	// Rate Limiting
	LoginRateLimit int `envconfig:"LOGIN_RATE_LIMIT" default:"5"` // requests per minute per IP
	WSRateLimit    int `envconfig:"WS_RATE_LIMIT" default:"20"`   // connections per minute per IP
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}

// MustLoad loads configuration and panics on error
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
