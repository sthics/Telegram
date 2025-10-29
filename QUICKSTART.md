# Quick Start Guide

Get Mini-Telegram running in 5 minutes with automatic dependency installation!

## Prerequisites Check

First, see what you already have:

```bash
cd /path/to/mini-telegram
./scripts/check-dependencies.sh
```

This will show you:
- ✓ What's already installed
- ✗ What's missing
- ⚠ Optional tools that would be helpful

## Automatic Installation

If anything is missing, run the automatic installer:

```bash
./scripts/install-dependencies.sh
```

This script will automatically install:

### Core Requirements
- **Go 1.22+** - Programming language
- **Docker** - Container runtime
- **Docker Compose** - Multi-container orchestration
- **OpenSSL** - For generating JWT keys

### Development Tools
- **goose** - Database migrations
- **buf** - Protobuf code generation
- **golangci-lint** - Code linting
- **k6** - Load testing

### Platform Support
- ✅ **macOS** (Intel & Apple Silicon) - Uses Homebrew
- ✅ **Linux** (Ubuntu/Debian) - Uses apt-get
- ⚠️ **Windows** - Manual installation required

## Complete Setup

After dependencies are installed, run the setup script:

```bash
./scripts/setup.sh
```

This will:
1. Generate JWT ES256 encryption keys
2. Create `.env` configuration file
3. Start PostgreSQL, Redis, and RabbitMQ
4. Run database migrations
5. Verify everything is ready

## Start Development

### Option 1: Docker Compose (Easiest)

Start all services with one command:

```bash
docker-compose up --build
```

Access:
- API Gateway: http://localhost:8080
- RabbitMQ Dashboard: http://localhost:15672 (guest/guest)

### Option 2: Local Development

Run each service separately for development:

```bash
# Terminal 1 - Infrastructure
docker-compose up postgres redis rabbitmq

# Terminal 2 - Gateway
make run-gateway

# Terminal 3 - Chat Service
make run-chat

# Terminal 4 - Presence Service
make run-presence
```

## Verify Installation

Test the API:

```bash
# Health check
curl http://localhost:8080/v1/health

# Register a user
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123"}'
```

## Troubleshooting

### Docker Not Running

**macOS:**
```bash
open -a Docker
# Wait 30 seconds, then try again
```

**Linux:**
```bash
sudo systemctl start docker
sudo systemctl enable docker
```

### Permission Denied (Linux)

Add your user to the docker group:
```bash
sudo usermod -aG docker $USER
# Log out and back in
```

### Go Command Not Found

Update your shell PATH:

**macOS (zsh):**
```bash
source ~/.zshrc
```

**Linux (bash):**
```bash
source ~/.bashrc
```

### Port Already in Use

Check what's using the port:
```bash
# macOS
lsof -i :8080

# Linux
sudo netstat -tlnp | grep 8080
```

Kill the process or change the port in `.env`:
```bash
PORT=9090
```

## Next Steps

1. **Read the docs** - Check `/docs` for detailed documentation
2. **Run tests** - `make test`
3. **Load test** - `k6 run load/k6-ws.js`
4. **View metrics** - Import `grafana/dashboard.json`
5. **Explore the API** - See README.md for all endpoints

## Development Workflow

```bash
# Install dependencies
make deps

# Build all services
make build

# Run tests
make test

# Generate coverage report
make coverage

# Format code
make fmt

# Lint code
make lint

# Clean build artifacts
make clean
```

## Environment Variables

Key settings in `.env`:

```bash
# Database
DSN=postgres://user:pgpass@localhost:5432/minitelegram?sslmode=disable

# Redis
REDIS_ADDR=localhost:6379

# RabbitMQ
AMQP_URL=amqp://guest:guest@localhost:5672/

# JWT (generate with 'make generate-keys')
JWT_PRIVATE_KEY_PATH=./secrets/es256.key

# Performance
CONN_TTL=35s
PING_INTERVAL=30s
```

## Getting Help

- **Documentation**: See `/docs` folder
- **Issues**: Check existing issues on GitHub
- **Logs**: `docker-compose logs -f [service]`
- **Health**: `curl http://localhost:8080/v1/health`

## What Gets Installed

| Tool | Version | Purpose | Size |
|------|---------|---------|------|
| Go | 1.22+ | Programming language | ~100MB |
| Docker Desktop | Latest | Containers | ~500MB |
| Docker Compose | 2.x | Orchestration | Included |
| OpenSSL | Latest | Key generation | ~5MB |
| goose | Latest | Database migrations | ~10MB |
| buf | Latest | Protobuf generation | ~20MB |
| golangci-lint | Latest | Code linting | ~50MB |
| k6 | Latest | Load testing | ~50MB |

**Total**: ~735MB (most is Docker Desktop)

## Uninstall

To remove Mini-Telegram:

```bash
# Stop and remove containers
docker-compose down -v

# Remove Docker images
docker rmi $(docker images 'mini-telegram/*' -q)

# Remove project directory
cd ..
rm -rf mini-telegram
```

To uninstall tools (optional):
```bash
# macOS
brew uninstall go docker docker-compose k6 golangci-lint

# Linux - Docker
sudo apt-get remove docker-ce docker-ce-cli containerd.io
```

---

**Ready to build something amazing? Let's go! 🚀**
