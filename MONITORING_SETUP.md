# Monitoring Setup Guide

## Start Everything

```bash
# Start all services including monitoring
docker-compose up -d
```

## Access the Dashboards

Open these URLs in your browser:

1. **Grafana (Main Dashboard)**: http://localhost:3000
   - Username: `admin`
   - Password: `admin`
   - The "Mini-Telegram Dashboard" is auto-loaded

2. **Prometheus (Raw Metrics)**: http://localhost:9090

3. **RabbitMQ Management**: http://localhost:15672
   - Username: `guest`
   - Password: `guest`

## What You'll See in Grafana

The dashboard shows:
- **Message Rate**: How many messages/second are being sent
- **p99 Delivery Latency**: How fast messages are delivered (99th percentile)
- **Active Goroutines**: Number of concurrent operations per service
- **Memory Usage**: Heap memory for each service
- **Redis Metrics**: Connected clients and commands/sec
- **PostgreSQL Metrics**: Active connections and transactions/sec
- **RabbitMQ Queue Metrics**: Messages in queues and publish rate

## First Time Setup

1. Start the services:
```bash
docker-compose up -d
```

2. Wait 30 seconds for everything to start

3. Open http://localhost:3000 in your browser

4. Login with `admin` / `admin`

5. You'll see the dashboard automatically loaded!

## Viewing Metrics

The dashboard auto-refreshes every 5 seconds. You can:
- Change the time range (top-right corner)
- Click on any graph to zoom in
- Hover over lines to see exact values

## Stop Everything

```bash
docker-compose down
```

## Troubleshooting

**No data showing?**
- Make sure your services are running: `docker-compose ps`
- Check Prometheus targets: http://localhost:9090/targets (all should be "UP")
- Wait a minute for metrics to start collecting
