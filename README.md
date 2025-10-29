# Mini-Telegram

A high-performance, real-time messaging service built with Go, supporting 10k concurrent WebSocket connections with <200ms p99 delivery latency.

## 🚀 Quick Start (One Command!)

```bash
# Check what you have installed
./scripts/check-dependencies.sh

# Install everything you need automatically
./scripts/install-dependencies.sh

# Set up and start the project
./scripts/setup.sh
```

That's it! The scripts will handle everything: Go, Docker, all tools, and configuration.

## Features

- Real-time WebSocket messaging
- JWT authentication (ES256)
- End-to-end encryption support (optional)
- Direct and group chats
- Message persistence and history
- Read receipts and presence
- Horizontal scalability
- Zero-downtime deployments

## Architecture

### Core Components

1. **gateway** - Stateless WebSocket gateway (Gin + gorilla/websocket)
2. **chat-svc** - Stateless chat message processor
3. **presence-svc** - Stateless presence and receipt processor
4. **PostgreSQL** - Message and user persistence
5. **Redis** - Connection registry, presence cache, group members
6. **RabbitMQ** - Message queue for async processing

### Tech Stack

- Go 1.22
- Gin (HTTP router)
- gorilla/websocket
- PostgreSQL 15
- Redis 7
- RabbitMQ 3.12
- GORM (ORM)
- Prometheus + Grafana (observability)

## Performance Targets

| Metric | Target |
|--------|--------|
| p50 latency | ≤50 ms |
| p99 latency | ≤200 ms |
| Concurrent sockets | 10,000 |
| Gateway RAM | ≤40 MB / 5k sockets |

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- (Optional) buf CLI for protobuf generation

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/ambarg/mini-telegram.git
cd mini-telegram
```

2. Generate JWT keys:
```bash
mkdir -p secrets
openssl ecparam -genkey -name prime256v1 -noout -out secrets/es256.key
```

3. Copy environment file:
```bash
cp .env.example .env
```

4. Start services:
```bash
docker-compose up -d postgres redis rabbitmq
```

5. Run database migrations:
```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
goose -dir db/migrations postgres "postgres://user:pgpass@localhost:5432/minitelegram?sslmode=disable" up
```

6. Run the gateway:
```bash
go run cmd/gateway/main.go
```

7. Run workers:
```bash
# Terminal 1
go run cmd/chat-svc/main.go

# Terminal 2
go run cmd/presence-svc/main.go
```

### Docker Compose (All Services)

```bash
docker-compose up --build
```

Access:
- Gateway API: http://localhost:8080
- RabbitMQ Management: http://localhost:15672 (guest/guest)

## API Endpoints

### Authentication

**Register:**
```bash
POST /v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Login:**
```bash
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Refresh Token:**
```bash
POST /v1/auth/refresh
Cookie: refreshToken=<token>
```

### Chats

**Get Chats:**
```bash
GET /v1/chats
Authorization: Bearer <accessToken>
```

**Create Chat:**
```bash
POST /v1/chats
Authorization: Bearer <accessToken>
Content-Type: application/json

{
  "type": 2,
  "memberIds": [2, 3, 4]
}
```

### WebSocket

Connect to WebSocket:
```
wss://localhost:8080/v1/ws?token=<accessToken>&device=web
```

Send message:
```json
{
  "type": "SendMessage",
  "uuid": "client-generated-uuid",
  "chatId": 1,
  "body": "Hello, World!"
}
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

```bash
go test -tags=integration ./...
```

### Load Testing (k6)

```bash
# Install k6
brew install k6  # macOS
# or download from https://k6.io/

# Run load test
k6 run load/k6-ws.js
```

### Coverage

```bash
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out -o cover.html
```

## Database Migrations

```bash
# Up
goose -dir db/migrations postgres "$DSN" up

# Down
goose -dir db/migrations postgres "$DSN" down

# Status
goose -dir db/migrations postgres "$DSN" status
```

## Observability

### Metrics

Prometheus metrics available at `/metrics`:
- `gateway_msg_sent_total` - Total messages sent
- `gateway_delivery_duration_seconds` - Message delivery latency histogram
- `gateway_ws_connections` - Active WebSocket connections
- `go_goroutines` - Active goroutines

### Logs

Structured JSON logs using zerolog:
```json
{
  "level": "info",
  "time": 1234567890,
  "service": "gateway",
  "user_id": 123,
  "chat_id": 456,
  "latency": 45,
  "message": "message sent"
}
```

## Deployment

### Docker

Build images:
```bash
docker build -f Dockerfile.gateway -t gateway:latest .
docker build -f Dockerfile.chat-svc -t chat-svc:latest .
docker build -f Dockerfile.presence-svc -t presence-svc:latest .
```

### Kubernetes (k3s)

See deployment configurations in `/build` directory.

### CI/CD

GitHub Actions workflow in `.github/workflows/deploy.yml`:
1. Run tests
2. Build Docker images
3. Push to GitHub Container Registry
4. Deploy to k3s cluster

## Security

- JWT tokens with ES256 (ECDSA P-256)
- Access token: 15 minutes
- Refresh token: 7 days (HttpOnly cookie)
- bcrypt password hashing (cost 12)
- Rate limiting on auth endpoints (5 req/min)
- Rate limiting on WebSocket connections (20 conn/min)
- TLS 1.3 only in production
- Input validation on all endpoints

## Performance Optimization

- Connection pooling (PostgreSQL, Redis)
- RabbitMQ lazy queues
- Redis caching for group members and presence
- Batch processing for read receipts (50ms window)
- Binary protobuf frames (5× smaller than JSON)
- Stateless services for horizontal scaling

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## Documentation

See `/docs` directory for detailed documentation:
- Architecture overview
- API specification
- Database schema
- WebSocket protocol
- Deployment guide
- Testing strategy
- Troubleshooting guide
