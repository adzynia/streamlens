# StreamLens - Project Summary

## What We Built

A production-ready **LLM Observability Service** in Go using Redpanda (Kafka) as the streaming backbone. The system ingests telemetry from LLM-powered applications, processes events in real-time with stream processing, and provides a REST API for querying aggregated metrics.

## Architecture Components

### 1. Ingestion API (HTTP Service)
- **Port**: 8080
- **Endpoints**: 
  - `POST /v1/llm/request` - Ingest LLM request events
  - `POST /v1/llm/response` - Ingest LLM response events
  - `GET /health` - Health check
- **Functionality**: Validates events and publishes to Redpanda topics

### 2. Metrics Processor (Stream Processor)
- **Technology**: franz-go Kafka client
- **Functionality**:
  - Consumes from `llm.requests` and `llm.responses` topics
  - Joins requests/responses by `request_id` (in-memory state)
  - Aggregates into 1-minute tumbling windows per (tenant, route, model)
  - Computes metrics: count, errors, avg/p95 latency, tokens, cost
  - Writes to Postgres + produces to `llm.metrics` topic

### 3. Metrics API (HTTP Service)
- **Port**: 8081
- **Endpoints**: 
  - `GET /v1/metrics?tenant_id=X&route=Y&model=Z&limit=N`
  - `GET /health`
- **Functionality**: Query aggregated metrics from Postgres

## Tech Stack

- **Language**: Go 1.22
- **Message Broker**: Redpanda (Kafka-compatible)
- **Kafka Client**: franz-go (production-grade Go client)
- **Database**: PostgreSQL (with indexes for fast queries)
- **HTTP Framework**: chi router
- **Infrastructure**: Docker Compose

## Key Features

✅ **Stream Processing**
- Real-time event joining by request_id
- 1-minute tumbling windows
- Thread-safe in-memory state management
- Automatic state cleanup (5min TTL)

✅ **Metrics Computation**
- Request/error counts
- Average & P95 latency
- Token usage (prompt/completion)
- Cost estimation (configurable pricing)

✅ **Production-Ready**
- Graceful shutdown (SIGTERM/SIGINT)
- Error handling and logging
- Connection pooling
- Health checks
- Configurable via environment variables

✅ **Developer Experience**
- Makefile with common tasks
- Docker Compose for one-command startup
- Traffic generator script for testing
- Comprehensive documentation

## Project Structure

```
streamlens/
├── cmd/                      # Service entry points
│   ├── ingestion-api/
│   ├── metrics-processor/
│   └── metrics-api/
├── internal/                 # Shared packages
│   ├── config/              # Config management
│   ├── handlers/            # HTTP handlers
│   ├── kafka/               # Kafka wrappers
│   ├── models/              # Data models
│   ├── processor/           # Stream processing logic
│   └── store/               # Postgres layer
├── deploy/                   # Infrastructure
│   ├── docker-compose.yml
│   ├── Dockerfile
│   └── init.sql
├── scripts/                  # Utilities
│   ├── generate-traffic.sh
│   └── test-api.sh
├── Makefile
├── README.md                # User guide
├── QUICKSTART.md           # Quick start
├── ARCHITECTURE.md         # Deep dive
└── PROJECT_SUMMARY.md      # This file
```

## Files Created

**Go Services (17 files)**:
- 3 main.go files (one per service)
- 7 internal packages
- Build artifacts in go.mod/go.sum

**Infrastructure (4 files)**:
- Docker Compose configuration
- Dockerfile (multi-service build)
- PostgreSQL schema (init.sql)
- Environment template

**Scripts (2 files)**:
- Traffic generator (generate-traffic.sh)
- API test script (test-api.sh)

**Documentation (5 files)**:
- README.md (comprehensive user guide)
- QUICKSTART.md (5-minute tutorial)
- ARCHITECTURE.md (deep technical dive)
- PROJECT_SUMMARY.md (this file)
- .env.example (config template)

**Build & Dev (2 files)**:
- Makefile (development workflows)
- .gitignore (version control)

**Total**: ~30 files created

## How to Use

### Quick Start (Docker)

```bash
# 1. Start everything
make docker-up

# 2. Generate sample traffic
make generate-traffic

# 3. Wait ~2 minutes, then query
curl 'http://localhost:8081/v1/metrics?tenant_id=tenant-1' | jq
```

### Local Development

```bash
# 1. Start infrastructure
docker compose -f deploy/docker-compose.yml up -d redpanda postgres

# 2. Install dependencies
make deps

# 3. Run services (3 terminals)
make run-ingestion
make run-processor
make run-metrics-api
```

## Example Usage

### Send Events

```bash
# Request event
curl -X POST http://localhost:8080/v1/llm/request \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req-123",
    "tenant_id": "acme",
    "route": "chat_support",
    "model": "gpt-4.1-mini",
    "timestamp": "2025-11-20T10:00:00.000Z",
    "prompt_tokens": 100
  }'

# Response event
curl -X POST http://localhost:8080/v1/llm/response \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req-123",
    "timestamp": "2025-11-20T10:00:01.200Z",
    "latency_ms": 1200,
    "completion_tokens": 250,
    "finish_reason": "stop"
  }'
```

### Query Metrics

```bash
# All metrics for a tenant
curl 'http://localhost:8081/v1/metrics?tenant_id=acme' | jq

# Filtered by route and model
curl 'http://localhost:8081/v1/metrics?tenant_id=acme&route=chat_support&model=gpt-4.1-mini' | jq
```

## Data Flow Summary

1. **Event Ingestion**: LLM apps → Ingestion API → Redpanda topics
2. **Stream Processing**: Processor joins events, windows, aggregates
3. **Storage**: Metrics → Postgres + Kafka topic
4. **Query**: Dashboard → Metrics API → Postgres

## Metrics Schema

```json
{
  "tenant_id": "acme",
  "route": "chat_support",
  "model": "gpt-4.1-mini",
  "window_start": "2025-11-20T10:00:00Z",
  "window_end": "2025-11-20T10:01:00Z",
  "requests": 1234,
  "errors": 12,
  "avg_latency_ms": 732.4,
  "p95_latency_ms": 1234.0,
  "avg_prompt_tokens": 300.1,
  "avg_completion_tokens": 420.6,
  "estimated_cost_usd": 2.31
}
```

## Ports

| Service | Port | Purpose |
|---------|------|---------|
| Ingestion API | 8080 | Event ingestion |
| Metrics API | 8081 | Query metrics |
| Redpanda | 19092 | Kafka API |
| Postgres | 5432 | Database |
| Redpanda Admin | 19644 | Admin API |

## Next Steps

1. **Run it**: `make docker-up`
2. **Test it**: `make generate-traffic`
3. **Query it**: Check metrics after 2 minutes
4. **Customize it**: Modify code in `internal/`
5. **Scale it**: See ARCHITECTURE.md for production tips

## Design Highlights

### Stream Processing
- In-memory state store for request/response joining
- Tumbling windows (1-minute, non-overlapping)
- Automatic cleanup of old state
- Thread-safe with RWMutex

### Metrics Computation
- Request count and error tracking
- Latency percentiles (P95)
- Token usage aggregation
- Simple cost estimation model

### Production Features
- Graceful shutdown everywhere
- Proper error handling
- Connection pooling (Postgres)
- Consumer group for processor
- Idiomatic Go patterns

## Commands Reference

```bash
# Docker
make docker-up          # Start all services
make docker-down        # Stop services
make docker-logs        # View logs
make clean              # Stop and remove volumes

# Local Development
make deps               # Download dependencies
make build              # Build binaries
make run-ingestion      # Run ingestion API
make run-processor      # Run metrics processor
make run-metrics-api    # Run metrics API

# Testing
make test               # Run tests
make generate-traffic   # Generate sample data
./scripts/test-api.sh   # Test APIs manually
```

## Documentation

- **README.md**: Full user guide with API docs
- **QUICKSTART.md**: Get running in 5 minutes
- **ARCHITECTURE.md**: Deep dive into design and implementation
- **PROJECT_SUMMARY.md**: This overview

---

**Built with ❤️ using Go, Redpanda, and PostgreSQL**

Ready to observe your LLMs! 🚀
