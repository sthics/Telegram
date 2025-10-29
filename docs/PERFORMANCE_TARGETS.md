# Performance Targets

| Metric | Target | Tool |
|--------|--------|------|
| p50 latency | ≤50 ms | Prometheus hist |
| p99 latency | ≤200 ms | Prometheus |
| WS connect | ≤150 ms | k6 |
| Gateway RAM | ≤40 MB / 5 k sockets | pprof |
| DB query p99 | ≤30 ms | pg_stat_statements |

## Capacity Plan
| Users | Gateway pods (2 vCPU) | Postgres vCPU | RabbitMQ RAM | Redis RAM | Monthly $ |
|-------|----------------------|---------------|--------------|-----------|-----------|
| 1 k   | 1                    | 2             | 512 MB       | 100 MB    | $15       |
| 10 k  | 5                    | 4             | 2 GB         | 500 MB    | $60       |
| 100 k | 50                   | 16            | 8 GB         | 2 GB      | $500      |

Calculation: 1 pod handles 2 k sockets @ 20 MB RSS
