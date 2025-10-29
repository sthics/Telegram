# Deployment

## Dependency Versions
go 1.22  
gin 1.9.1  
gorilla/websocket 1.5.1  
postgres 15-alpine  
rabbitmq 3.12-alpine  

## .env.example (commit this)
```
GIN_MODE=release
DSN=postgres://user:pass@localhost:5432/myapp?sslmode=disable
AMQP_URL=amqp://guest:guest@localhost:5672/
REDIS_ADDR=localhost:6379
JWT_PRIVATE_KEY_PATH=/secrets/es256.key
```

## Docker Multi-stage
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum . && go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o gateway ./cmd/gateway

FROM scratch
COPY --from=builder /src/gateway /gateway
ENTRYPOINT ["/gateway"]
```

## Rollback Procedure
1. List history: `kubectl rollout history deployment/gateway`
2. Rollback: `kubectl rollout undo deployment/gateway --to-revision=2`
3. Verify: `curl http://gateway/health`
4. DB: if migration was applied, run `goose down 1` (manual)

## GitHub Actions Pipeline
File: `.github/workflows/deploy.yml`
```yaml
name: CI-CD
on:
  push:
    branches: [main]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with: {go-version: '1.22'}
      - run: go test ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: docker/setup-buildx-action@v2
      - uses: docker/login-action@v2
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v4
        with:
          push: true
          tags: ghcr.io/${{ github.repository }}:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: deploy to k3s
        uses: appleboy/ssh-action@v0.1.5
        with:
          host: ${{ secrets.K3S_HOST }}
          username: ubuntu
          key: ${{ secrets.K3S_KEY }}
          script: |
            kubectl set image deployment/gateway \
              gateway=ghcr.io/${{ github.repository }}:${{ github.sha }}
```

## Local docker-compose
```yaml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: pgpass
  rabbitmq:
    image: rabbitmq:3.12-management-alpine
    ports: ["5672:5672","15672:15672"]
  gateway:
    build: .
    env_file: .env
    ports: ["8080:8080"]
```

## Zero-Downtime
Rolling update with `maxUnavailable: 1`, `maxSurge: 1`  
Readiness probe: `/health` returns 200
