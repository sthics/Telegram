# Testing Strategy

## Unit (Go standard)
- Table-driven tests with `testify/assert`
- Mock external deps with `gomock`
- Target coverage ≥70 %

Example:
```go
func TestHashPassword(t *testing.T){
    tests := []struct{
        pass    string
        wantErr bool
    }{
        {"short", true},
        {"ValidPass123", false},
    }
    for _, tt := range tests {
        _, err := HashPassword(tt.pass)
        assert.Equal(t, tt.wantErr, err != nil)
    }
}
```

## Integration (testcontainers)
Full services spun up per test:
```go
func TestChatFlow(t *testing.T) {
    ctx := context.Background()

    // Postgres
    pgContainer, _ := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
        postgres.WithDatabase("test"),
        postgres.WithUsername("user"),
        postgres.WithPassword("pass"),
    )
    defer pgContainer.Terminate(ctx)

    // RabbitMQ
    mqContainer, _ := rabbitmq.RunContainer(ctx,
        testcontainers.WithImage("rabbitmq:3.12-alpine"),
    )
    defer mqContainer.Terminate(ctx)

    // Connect real services
    db := setupDB(mustConnectionString(pgContainer))
    broker := setupRabbitMQ(mustAmqpURL(mqContainer))

    // Test: send message → wait for delivery
    msg := SendMessage{ChatID: 1, Body: "hello"}
    err := gateway.Send(msg)
    require.NoError(t, err)

    delivered := awaitDelivery(ctx, msg.UUID, 2*time.Second)
    assert.True(t, delivered)
}
```

## Load Testing (k6)
Script: `load/k6-ws.js`
```javascript
import ws from 'k6/ws';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 1000 },
    { duration: '3m', target: 5000 },
    { duration: '5m', target: 10000 },
  ],
};

const JWT = __ENV.JWT || 'test-jwt';

export default function () {
  const url = `wss://api.myapp.com/v1/ws?token=${JWT}`;
  const response = ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      socket.send(JSON.stringify({
        type: 'SendMessage',
        chatId: 1,
        body: 'load test',
        uuid: `${__VU}-${__ITER}`,
      }));
    });
    socket.on('message', (data) => {
      const msg = JSON.parse(data);
      check(msg, { 'delivered': (m) => m.type === 'Delivered' });
    });
  });
  check(response, { 'status is 101': (r) => r && r.status === 101 });
}
```
Run:  
```bash
k6 run --env JWT=$(cat test.jwt) load/k6-ws.js
```

## WebSocket Testing
Use gorilla client in Go test; send 1 k messages and await acks to guarantee ordering.

## Coverage
```bash
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out -o cover.html
```
Push `cover.html` as CI artifact.
