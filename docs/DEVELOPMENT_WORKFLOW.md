# Development Workflow

## Git Branching
main ← feature/foo ← hotfix/bar  
No direct push to main; PR required.

## Commit Message
```
feat: add voice note protobuf
fix: redis leak on ws close
docs: update deployment steps
```

## PR Checklist
- [ ] go test ./... passes
- [ ] docker build succeeds
- [ ] swagger spec updated
- [ ] k6 smoke test ≥5 k conn
- [ ] metrics documented

## Local Setup
1. `git clone` && `cd mini-telegram`
2. `docker-compose up -d postgres rabbitmq redis`
3. `air` (live-reload) inside gateway folder
4. `go run cmd/chat-svc/main.go`

## Generate Protobuf Code
Install buf:
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
```
Generate:
```bash
buf generate
```
Output: `internal/pb/*.pb.go`

buf.gen.yaml:
```yaml
version: v1
plugins:
  - plugin: buf.build/protocolbuffers/go
    out: internal/pb
    opt: paths=source_relative
```

## Debugging
- `dlv debug` remote attach port 2345
- RabbitMQ UI http://localhost:15672
- Prometheus http://localhost:9090
