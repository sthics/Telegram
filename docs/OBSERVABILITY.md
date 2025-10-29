# Observability

## Prometheus Metrics (Go)
```go
var msgSent = prometheus.NewCounterVec(
  prometheus.CounterOpts{Name:"gateway_msg_sent_total"},
  []string{"chat_type"})
var durHist = prometheus.NewHistogramVec(
  prometheus.HistogramOpts{Name:"gateway_delivery_duration_seconds",Buckets:[]float64{.01,.05,.1,.2}},
  []string{"status"})
```

## Grafana Dashboard JSON
Save as `grafana/dashboard.json` and import via UI → Upload.
```json
{
  "dashboard": {
    "title": "Mini-Telegram",
    "uid": "mt",
    "panels": [
      {
        "title": "Message Rate",
        "type": "stat",
        "targets": [{"expr": "rate(gateway_msg_sent_total[5m])"}]
      },
      {
        "title": "p99 Delivery",
        "type": "stat",
        "targets": [{"expr": "histogram_quantile(0.99, gateway_delivery_duration_seconds)"}]
      },
      {
        "title": "Goroutines",
        "type": "graph",
        "targets": [{"expr": "go_goroutines"}]
      }
    ]
  }
}
```

## Logging
Structured JSON via `zerolog`  
Fields: `time,level,service,uid,chatId,latency,error`

## Tracing (optional)
OpenTelemetry + Jaeger agent side-car  
Trace parent passed in `X-Trace-ID` header & AMQP header.

## Alerting Rules (Prometheus)
- `rate(gateway_msg_sent_total[5m]) == 0` → page
- `histogram_quantile(0.99, gateway_delivery_duration_seconds) > 0.2` → warn
