# StreamLens Quick Start Guide

Get StreamLens up and running in 5 minutes!

## 🚀 Option 1: Docker (Recommended)

### Step 1: Start all services

```bash
make docker-up
```

This will start:
- Redpanda (Kafka) on port 19092
- Postgres on port 5432
- Ingestion API on port 8080
- Metrics Processor (background)
- Metrics API on port 8081

### Step 2: Wait for services to be ready

```bash
# Watch logs to see when services are ready
make docker-logs

# Look for these messages:
# - "Ingestion API listening on port 8080"
# - "Metrics API listening on port 8081"
# - "Starting metrics processor..."
```

### Step 3: Generate sample traffic

```bash
make generate-traffic
```

This will send 20 request/response pairs with realistic latency and token data.

### Step 4: Query metrics

Wait ~2 minutes for the windowing to complete, then:

```bash
# Get all metrics for tenant-1
curl 'http://localhost:8081/v1/metrics?tenant_id=tenant-1' | jq

# Filter by route
curl 'http://localhost:8081/v1/metrics?tenant_id=tenant-1&route=chat_support_v1' | jq
```

### Step 5: Explore

```bash
# Health checks
curl http://localhost:8080/health
curl http://localhost:8081/health

# Send a custom request
curl -X POST http://localhost:8080/v1/llm/request \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "custom-123",
    "tenant_id": "my-tenant",
    "route": "my-route",
    "model": "gpt-4",
    "timestamp": "2025-11-20T10:00:00.000Z",
    "prompt_tokens": 100
  }'
```

### Step 6: Cleanup

```bash
make docker-down
# Or to remove volumes too:
make clean
```

---

## 🛠️ Option 2: Local Development

### Step 1: Start infrastructure only

```bash
docker compose -f deploy/docker-compose.yml up -d redpanda postgres
```

### Step 2: Install dependencies

```bash
make deps
```

### Step 3: Run services locally (3 terminals)

**Terminal 1 - Ingestion API:**
```bash
make run-ingestion
```

**Terminal 2 - Metrics Processor:**
```bash
make run-processor
```

**Terminal 3 - Metrics API:**
```bash
make run-metrics-api
```

### Step 4: Test

Use the same testing steps as Docker option above.

---

## 📊 Understanding the Data Flow

1. **Request Event** → Ingestion API → `llm.requests` topic
2. **Response Event** → Ingestion API → `llm.responses` topic
3. **Processor** reads both topics, joins by `request_id`
4. **Windowing** aggregates into 1-minute buckets
5. **Metrics** written to:
   - Postgres (`llm_metrics` table)
   - Kafka (`llm.metrics` topic)
6. **Query** via Metrics API from Postgres

---

## 🔍 Troubleshooting

### Services not starting?

```bash
# Check if ports are already in use
lsof -i :8080
lsof -i :8081
lsof -i :19092
lsof -i :5432

# View detailed logs
docker compose -f deploy/docker-compose.yml logs ingestion-api
docker compose -f deploy/docker-compose.yml logs metrics-processor
```

### No metrics appearing?

- Wait at least 2 minutes after sending traffic (windowing delay)
- Check processor logs: `docker compose -f deploy/docker-compose.yml logs metrics-processor`
- Verify events are in Kafka topics (would need rpk CLI or Kafka tool)
- Check Postgres: `docker exec -it postgres psql -U streamlens -c "SELECT * FROM llm_metrics;"`

### Connection errors?

Make sure all services are healthy:
```bash
docker compose -f deploy/docker-compose.yml ps
```

All should show "Up" and "healthy" status.

---

## 📚 Next Steps

- Read the full [README.md](README.md) for API documentation
- Explore the code in `internal/` to understand the implementation
- Modify the traffic generator script to test your own scenarios
- Add custom metadata fields to track your specific use cases
- Build a dashboard on top of the Metrics API

---

## 💡 Tips

- The processor uses 1-minute tumbling windows, so there's a natural ~60-90 second delay
- Estimated costs use a simple model - customize in `internal/processor/processor.go`
- State is kept in memory - for production, consider adding Redis or RocksDB
- Each service has graceful shutdown (Ctrl+C works properly)

Happy observing! 🎉
