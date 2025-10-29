# Troubleshooting Guide

## WS 401 on connect
- Check JWT exp, clock skew
- Verify ES256 public key matches private

## High RabbitMQ memory
- Enable lazy queues: `queue-declare-args: x-queue-mode=lazy`
- Add TTL or max-length policy

## Postgres "too many connections"
- Use pgBouncer in transaction pool mode
- Increase `max_connections` or add read-replica

## Redis OOM
- Lower TTL (35 s → 30 s)
- Eviction policy `allkeys-lru`

## Gateway memory leak
- pprof heap: `go tool pprof http://:6060/debug/pprof/heap`
- Look for goroutine leak (unclosed `range` on channel)

## Delivery latency spikes
- Check RabbitMQ queue depth panel
- Enable publisher confirms, watch for nacks
- Check Postgres `pg_stat_statements` for slow INSERT
