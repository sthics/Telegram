# Security Checklist

## Threat Model (STRIDE)
| Threat | Attack | Mitigation |
|--------|--------|------------|
| Spoofing | Stolen JWT | Short exp 15 min, refresh rotation |
| Tampering | MitM on WS | TLS 1.3 only, HSTS |
| Repudiation | Deny sending msg | Message signature (E2EE) |
| Info disclosure | Read others' chats | Row-level security (RLS) in DB |
| DoS | Socket flood | Rate limit 20 conn/min/IP |
| Elevation | Admin escalation | RBAC, least-privilege K8s SA |

## Input Validation
- gin validator tags `binding:"required,min=1,max=10000"` on all structs
- SQL via GORM placeholders → no string concat

## Auth
- JWT ES256, 15 min exp, rotate refresh token
- bcrypt cost 12

## Rate Limit
- `/login` 5 req/min IP → `github.com/ulule/limiter`
- `/ws` 20 conn/min per IP (flood protection)

## Secrets
- JWT keys in K8s secret, mounted read-only
- Use sealed-secrets for GitOps

## Transport
- TLS 1.3 only, HSTS, 0-RTT disabled for idempotent only
- WebSocket `wss://` only
