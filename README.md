# Mini-Telegram

A high-performance, real-time messaging service built with Go (Backend) and React (Frontend), supporting 10k concurrent WebSocket connections with <200ms p99 delivery latency.

## 🚀 Quick Start (One Command!)

```bash
# Check what you have installed
./scripts/check-dependencies.sh

# Install everything you need automatically
./scripts/install-dependencies.sh

# Set up and start the project
./scripts/setup.sh
```

The setup script currently handles the backend. For the frontend:

```bash
cd web
npm install
npm run dev
```

## Features

- **Real-time WebSocket messaging** (JSON based)
- **Modern Web Client** (React 19, Tailwind, Vite)
- **JWT authentication** (ES256)
- **Direct and Group Chats**
- **Group Administration** (Roles, Titles, Permissions)
- **Media Support** (Image/File uploads via S3/MinIO)
- **Message Persistence** (PostgreSQL)
- **Read Receipts** (Ticks/Double Ticks)
- **Presence Status** (Online/Offline)
- **Push Notifications** (FCM infrastructure ready)
- **Horizontal Scalability** (RabbitMQ + Redis)

## Architecture

### Core Components

1.  **Web Client** - Single Page Application (React + Vite)
2.  **Gateway** - WebSocket & HTTP Gateway (Gin + gorilla/websocket)
3.  **Chat Service** - Core business logic and persistence
4.  **Presence Service** - Status management
5.  **PostgreSQL** - Primary data store (Users, Chats, Messages)
6.  **Redis** - Hot cache (Sessions, Presence, Websocket Routing)
7.  **RabbitMQ** - Message broker for async events and fan-out
8.  **MinIO** - Object storage for media

### Tech Stack

**Backend**
-   Go 1.22
-   Gin (HTTP), Gorilla (WS)
-   PostgreSQL 15, Redis 7, RabbitMQ 3.12
-   MinIO, AWS SDK v2
-   GORM, Zerolog, OpenTelemetry

**Frontend**
-   React 19
-   TypeScript
-   Vite
-   Tailwind CSS
-   Zustand (State)
-   TanStack Query (Data Fetching)

## Quick Start (Manual)

### Prerequisites

-   Go 1.22+
-   Node.js 18+
-   Docker & Docker Compose

### 1. Start Infrastructure

```bash
docker-compose up -d postgres redis rabbitmq minio createbuckets
```

### 2. Backend Setup

```bash
# Generate Keys
mkdir -p keys
openssl ecparam -genkey -name prime256v1 -noout -out keys/private.pem
openssl ec -in keys/private.pem -pubout -out keys/public.pem

# Migrations
go install github.com/pressly/goose/v3/cmd/goose@latest
export DSN="host=localhost user=user password=pgpass dbname=minitelegram port=5432 sslmode=disable"
goose -dir db/migrations postgres "$DSN" up

# Run Gateway
export REDIS_ADDR="localhost:6379"
export AMQP_URL="amqp://guest:guest@localhost:5672/"
export JWT_PRIVATE_KEY_PATH="keys/private.pem"
go run cmd/gateway/main.go
```

### 3. Frontend Setup

```bash
cd web
npm install
npm run dev
```

Access the web client at `http://localhost:5173`.

## API Documentation

The backend exposes a Swagger UI at `http://localhost:8080/swagger/index.html`.

Key Endpoints:
-   `POST /v1/auth/register` - Create account
-   `POST /v1/auth/login` - Get JWT
-   `GET /v1/chats` - List chats
-   `POST /v1/chats` - Create chat
-   `GET /v1/ws` - WebSocket connection

## License

MIT License.
