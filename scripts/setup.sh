#!/bin/bash
set -e

echo "=== Mini-Telegram Setup Script ==="
echo ""

# Check for required tools
echo "Checking for required tools..."

if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.22 or later."
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed. Please install Docker."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "Error: Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

echo "✓ All required tools are installed"
echo ""

# Generate JWT keys if they don't exist
if [ ! -f "secrets/es256.key" ]; then
    echo "Generating JWT ES256 keys..."
    mkdir -p secrets
    openssl ecparam -genkey -name prime256v1 -noout -out secrets/es256.key
    chmod 400 secrets/es256.key
    echo "✓ JWT keys generated"
else
    echo "✓ JWT keys already exist"
fi
echo ""

# Copy .env.example to .env if it doesn't exist
if [ ! -f ".env" ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo "✓ .env file created"
    echo "⚠️  Please review and update .env with your configuration"
else
    echo "✓ .env file already exists"
fi
echo ""

# Install Go dependencies
echo "Installing Go dependencies..."
go mod download
go mod verify
echo "✓ Dependencies installed"
echo ""

# Install development tools
echo "Installing development tools..."
go install github.com/pressly/goose/v3/cmd/goose@latest
echo "✓ Development tools installed"
echo ""

# Start infrastructure services
echo "Starting infrastructure services (PostgreSQL, Redis, RabbitMQ)..."
docker-compose up -d postgres redis rabbitmq
echo "Waiting for services to be healthy..."
sleep 10
echo "✓ Infrastructure services started"
echo ""

# Run database migrations
echo "Running database migrations..."
export DSN="postgres://user:pgpass@localhost:5432/minitelegram?sslmode=disable"
goose -dir db/migrations postgres "$DSN" up
echo "✓ Database migrations completed"
echo ""

echo "==================================="
echo "✓ Setup completed successfully!"
echo ""
echo "To start the application:"
echo "  1. Gateway:  make run-gateway"
echo "  2. Chat:     make run-chat"
echo "  3. Presence: make run-presence"
echo ""
echo "Or use Docker Compose:"
echo "  docker-compose up --build"
echo ""
echo "Access:"
echo "  - Gateway API:        http://localhost:8080"
echo "  - RabbitMQ Dashboard: http://localhost:15672 (guest/guest)"
echo "==================================="
