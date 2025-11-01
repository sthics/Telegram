# Kubernetes Deployment Guide

This directory contains Kubernetes manifests for deploying Mini-Telegram to a k3d cluster.

## 📁 File Structure

```
k8s/
├── namespace.yaml       # Namespace for isolating resources
├── configmap.yaml       # Application configuration
├── secret.yaml          # Sensitive data (JWT keys)
├── postgres.yaml        # PostgreSQL database
├── redis.yaml           # Redis cache
├── rabbitmq.yaml        # RabbitMQ message broker
├── gateway.yaml         # Gateway service (HTTP/WebSocket)
├── chat-svc.yaml        # Chat message processor
└── presence-svc.yaml    # Presence & read receipts processor
```

---

## 📋 File Descriptions

### `namespace.yaml`
Creates an isolated namespace called `minitelegram` for all resources.

**Why:** Separates Mini-Telegram from other apps in the cluster.

---

### `configmap.yaml`
Stores **non-sensitive** configuration as key-value pairs.

**Contains:**
- Database connection string (DSN)
- Redis address
- RabbitMQ URL
- JWT key path
- Server port

**Why:** Centralized config that all pods can reference. Easy to update without rebuilding images.

---

### `secret.yaml`
Stores **sensitive** data (JWT private keys).

**Security:** Base64 encoded, restricted access.

**Usage:** Mounted as a file in pods that need it (gateway).

---

### `postgres.yaml`
Deploys PostgreSQL database with persistent storage.

**Components:**
1. **PersistentVolumeClaim** - Requests 5GB storage for database data
2. **Service** - Internal ClusterIP for pods to connect (postgres:5432)
3. **Deployment** - Runs PostgreSQL container

**Data Persistence:** Uses PVC so data survives pod restarts.

---

### `redis.yaml`
Deploys Redis for caching and presence tracking.

**Components:**
1. **Service** - Internal ClusterIP (redis:6379)
2. **Deployment** - Runs Redis with AOF persistence enabled

**Why Single Replica:** Redis cluster not needed for this scale.

---

### `rabbitmq.yaml`
Deploys RabbitMQ message broker.

**Components:**
1. **Service** - Exposes AMQP (5672) and Management UI (15672)
2. **Deployment** - Runs RabbitMQ with management plugin

**Ports:**
- 5672: AMQP protocol (message queuing)
- 15672: Management UI (web interface)

---

### `gateway.yaml`
Deploys the Gateway service (API + WebSocket).

**Components:**
1. **Service** - LoadBalancer type (exposed on port 9090 via k3d)
2. **Deployment** - 2 replicas for high availability

**Features:**
- Handles HTTP REST APIs
- Manages WebSocket connections
- Publishes messages to RabbitMQ
- Requires JWT keys mounted from secret

**Scaling:** 2 replicas for redundancy. Can scale to more.

---

### `chat-svc.yaml`
Deploys Chat Service workers.

**Purpose:**
- Consumes messages from `chat.messages` queue
- Persists messages to PostgreSQL
- Creates delivery receipts
- Publishes delivery events

**Replicas:** 3 workers competing for messages from shared queue.

**Why No Service:** Workers don't need to be accessed; they consume from RabbitMQ.

---

### `presence-svc.yaml`
Deploys Presence Service workers.

**Purpose:**
- Processes read receipts in batches
- Updates user presence
- Handles online/offline events

**Replicas:** 2 workers for read receipt processing.

---

## 🚀 Deployment Order

Deploy in this order (dependencies first):

```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml         # Populate JWT key first!
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f rabbitmq.yaml
kubectl apply -f gateway.yaml
kubectl apply -f chat-svc.yaml
kubectl apply -f presence-svc.yaml
```

**Or all at once:**
```bash
kubectl apply -f k8s/
```

---

## 🔍 Key Concepts Explained

### Services

**ClusterIP:** Internal only, pods can access (postgres, redis, rabbitmq)
```yaml
type: ClusterIP
```

**LoadBalancer:** Externally accessible (gateway)
```yaml
type: LoadBalancer
```

---

### Deployments

**Replicas:** Number of pod copies running
```yaml
replicas: 2  # High availability
```

**Selectors:** How deployment finds its pods
```yaml
selector:
  matchLabels:
    app: gateway
```

**Image Pull Policy:**
```yaml
imagePullPolicy: Always  # Always pull latest from registry
```

---

### Environment Variables

**From ConfigMap:**
```yaml
env:
  - name: DSN
    valueFrom:
      configMapKeyRef:
        name: app-config
        key: DSN
```

**From Secret:**
```yaml
volumeMounts:
  - name: jwt-keys
    mountPath: /secrets
```

---

### Persistent Storage

**PVC (PersistentVolumeClaim):**
```yaml
spec:
  resources:
    requests:
      storage: 5Gi
```

**Why:** Data survives pod deletion/restart.

---

## 🌐 Networking

### Internal Communication (DNS)

Pods communicate using Kubernetes DNS:
```
postgres:5432          # PostgreSQL
redis:6379             # Redis
rabbitmq:5672          # RabbitMQ
gateway:8080           # Gateway (pod-to-pod)
```

### External Access (k3d Port Mapping)

```
localhost:9090   →  gateway:8080   (HTTP/WebSocket)
localhost:5433   →  postgres:5432  (if needed)
```

---

## 📊 Resource Architecture

```
┌─────────────────────────────────────────┐
│         k3d Cluster (minitelegram)      │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │     Namespace: minitelegram     │   │
│  │                                 │   │
│  │  ┌──────────┐  ┌──────────┐    │   │
│  │  │ Gateway  │  │ Gateway  │    │   │  ← LoadBalancer
│  │  │  (Pod)   │  │  (Pod)   │    │   │    :9090
│  │  └─────┬────┘  └─────┬────┘    │   │
│  │        │             │          │   │
│  │  ┌─────┴──────────┬──┴──────┐  │   │
│  │  │                │         │  │   │
│  │  ▼                ▼         ▼  │   │
│  │ PostgreSQL      Redis   RabbitMQ │  │
│  │                                 │   │
│  │  ▲                ▲         ▲  │   │
│  │  │                │         │  │   │
│  │  │   ┌────────────┴─────┬───┘  │   │
│  │  │   │                  │      │   │
│  │  │  Chat-Svc x3    Presence x2 │   │
│  │  └──────────────────────────── │   │
│  │                                 │   │
│  └─────────────────────────────────┘   │
│                                         │
└─────────────────────────────────────────┘
```

---

## 🔧 Useful Commands

### Check Status
```bash
kubectl get all -n minitelegram
kubectl get pods -n minitelegram
kubectl get services -n minitelegram
```

### View Logs
```bash
kubectl logs -n minitelegram deployment/gateway
kubectl logs -n minitelegram deployment/chat-svc
kubectl logs -n minitelegram -l app=gateway --tail=50
```

### Debug a Pod
```bash
kubectl describe pod -n minitelegram <pod-name>
kubectl exec -it -n minitelegram <pod-name> -- /bin/sh
```

### Scale Services
```bash
kubectl scale deployment gateway --replicas=5 -n minitelegram
kubectl scale deployment chat-svc --replicas=10 -n minitelegram
```

### Delete Everything
```bash
kubectl delete namespace minitelegram
```

---

## ⚠️ Before Deploying

1. **Build Docker Images:**
   ```bash
   docker build -f Dockerfile.gateway -t ghcr.io/sthics/telegram/gateway:latest .
   docker build -f Dockerfile.chat-svc -t ghcr.io/sthics/telegram/chat-svc:latest .
   docker build -f Dockerfile.presence-svc -t ghcr.io/sthics/telegram/presence-svc:latest .
   ```

2. **Push to Registry:**
   ```bash
   docker push ghcr.io/sthics/telegram/gateway:latest
   docker push ghcr.io/sthics/telegram/chat-svc:latest
   docker push ghcr.io/sthics/telegram/presence-svc:latest
   ```

3. **Create JWT Secret:**
   ```bash
   kubectl create secret generic jwt-keys \
     --from-file=es256.key=secrets/es256.key \
     -n minitelegram
   ```

---

## 📈 Scaling Strategy

### Horizontal Scaling

**Gateway:** Scale up for more concurrent connections
```bash
kubectl scale deployment gateway --replicas=10 -n minitelegram
```

**Chat-Svc:** Scale up for higher message throughput
```bash
kubectl scale deployment chat-svc --replicas=20 -n minitelegram
```

**Presence-Svc:** Scale based on read receipt volume
```bash
kubectl scale deployment presence-svc --replicas=5 -n minitelegram
```

### Why It Works

All services use **shared queues** (chat.messages, read.receipts):
- Workers automatically compete for messages
- No coordination needed
- Linear scalability

---

## 🎯 Production Considerations

### What's Missing for Production

1. **Resource Limits:**
   ```yaml
   resources:
     requests:
       memory: "256Mi"
       cpu: "250m"
     limits:
       memory: "512Mi"
       cpu: "500m"
   ```

2. **Health Checks:**
   ```yaml
   livenessProbe:
     httpGet:
       path: /health
       port: 8080
   readinessProbe:
     httpGet:
       path: /health
       port: 8080
   ```

3. **Ingress Controller:** For real domains instead of LoadBalancer

4. **TLS/SSL:** HTTPS certificates

5. **Secrets Management:** Use Sealed Secrets or external secret managers

6. **Monitoring:** Prometheus + Grafana

7. **Logging:** ELK stack or Loki

---

## 🔗 Related Documentation

- [K3d Documentation](https://k3d.io/)
- [Kubernetes Concepts](https://kubernetes.io/docs/concepts/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
