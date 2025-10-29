# RabbitMQ Patterns

## Exchange Topology
exchange: `chat.topic`  type: topic  durable: true  
exchange: `delivery.topic`  type: topic  durable: true

## Queues
- `chat.<chatId>` (auto-delete false, lazy, TTL 24 h)  
  bound to `chat.topic` with rk `<chatId>`

- `delivery` (per gateway pod)  
  bound to `delivery.topic` with rk list `<chatId1>, <chatId2>...`

## Routing Key Strategy
- chat-svc publishes to `chat.topic` rk = `<chatId>`  
- presence-svc publishes to `delivery.topic` rk = `<chatId>`

## Consumer Patterns (Go)
Use `github.com/rabbitmq/amqp091-go`  
Prefetch = 20, auto-ack = false, multiple-ack = true every 50 ms.

## Delivery Guarantees
- Publisher confirms enabled (wait ≤500 ms)  
- Persistent messages + durable queues  
- Consumer ACK only after Postgres commit

## Connection Pooling
Single long-lived connection per pod, 1 channel per goroutine (channels are cheap).
