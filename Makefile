.PHONY: help build test run clean docker migrate proto

# Include environment variables from .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build all binaries
	@echo "Building gateway..."
	go build -o bin/gateway ./cmd/gateway
	@echo "Building chat-svc..."
	go build -o bin/chat-svc ./cmd/chat-svc
	@echo "Building presence-svc..."
	go build -o bin/presence-svc ./cmd/presence-svc

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

coverage: test ## Generate coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run-gateway: ## Run gateway service
	go run ./cmd/gateway

run-chat: ## Run chat service
	go run ./cmd/chat-svc

run-presence: ## Run presence service
	go run ./cmd/presence-svc

docker-build: ## Build Docker images
	docker build -f Dockerfile.gateway -t mini-telegram/gateway:latest .
	docker build -f Dockerfile.chat-svc -t mini-telegram/chat-svc:latest .
	docker build -f Dockerfile.presence-svc -t mini-telegram/presence-svc:latest .

docker-up: ## Start all services with docker-compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## Show docker-compose logs
	docker-compose logs -f

migrate-up: ## Run database migrations up
	goose -dir db/migrations postgres "$(DSN)" up

migrate-down: ## Run database migrations down
	goose -dir db/migrations postgres "$(DSN)" down

migrate-status: ## Show migration status
	goose -dir db/migrations postgres "$(DSN)" status

proto: ## Generate protobuf code
	buf generate

generate-keys: ## Generate JWT ES256 keys
	@mkdir -p secrets
	openssl ecparam -genkey -name prime256v1 -noout -out secrets/es256.key
	@echo "Private key generated: secrets/es256.key"

install-tools: ## Install required tools
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	@echo "Tools installed successfully"

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...

deps: ## Download dependencies
	go mod download
	go mod verify

tidy: ## Tidy go.mod
	go mod tidy

all: clean deps build test ## Clean, download deps, build, and test
