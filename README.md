# StreamLens - LLM Observability Service

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A production-ready LLM observability platform built with Go and Redpanda (Kafka). StreamLens ingests telemetry events from LLM-powered applications, processes them via stream processing to compute real-time metrics, and exposes a query API for dashboards.

> **Portfolio Project**: This project demonstrates stream processing, microservices architecture, and real-time data pipelines for LLM observability.

## 🏗️ Architecture

```
┌─────────────────┐
│  LLM Apps       │
└────────┬────────┘
         │ HTTP
         ▼
┌─────────────────┐      ┌──────────────┐
│ Ingestion API   │─────▶│  Redpanda    │
│  (Port 8080)    │      │  (Kafka API) │
└─────────────────┘      └──────┬───────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │ Metrics Processor     │
                    │ - Join requests/resp  │
                    │ - Window aggregation  │
                    │ - Compute metrics     │
                    └───────┬───────────────┘
                            │
                    ┌───────┴────────┐
                    ▼                ▼
            ┌──────────────┐  ┌──────────┐
            │  Postgres    │  │ Redpanda │
            │  (Metrics)   │  │ (Metrics)│
            └──────┬───────┘  └──────────┘
                   │
                   ▼
          ┌─────────────────┐
          │  Metrics API    │
          │  (Port 8081)    │
          └─────────────────┘
```

## 📊 Data Flow

1. **Ingestion**: LLM apps send request/response events to Ingestion API
2. **Streaming**: Events flow through Redpanda topics (`llm.requests`, `llm.responses`)
3. **Processing**: Metrics Processor joins events and aggregates into 1-minute windows
4. **Storage**: Computed metrics stored in Postgres and published to `llm.metrics` topic
5. **Query**: Metrics API serves aggregated metrics for dashboards

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- Make (optional, for convenience)

### Run Everything with Docker

```bash
# Start all services (Redpanda, Postgres, and Go services)
make docker-up

# Or manually:
docker compose -f deploy/docker-compose.yml up -d

# View logs
make docker-logs
```

### Generate Sample Traffic

```bash
# Wait for services to be ready (~30 seconds), then:
make generate-traffic

# Or with custom parameters:
NUM_REQUESTS=50 BASE_URL=http://localhost:8080 ./scripts/generate-traffic.sh
```

### Query Metrics

After ~2 minutes (to allow windowing), query the metrics:

```bash
# Get metrics for tenant-1
curl 'http://localhost:8081/v1/metrics?tenant_id=tenant-1&limit=10' | jq

# Filter by route and model
curl 'http://localhost:8081/v1/metrics?tenant_id=tenant-1&route=chat_support_v1&model=gpt-4.1-mini' | jq
```

## 🛠️ Development

### Local Development (without Docker)

1. **Start infrastructure** (Redpanda + Postgres):
   ```bash
   # In docker-compose, comment out the Go services and run:
   docker compose -f deploy/docker-compose.yml up redpanda postgres
   ```

2. **Install dependencies**:
   ```bash
   make deps
   ```

3. **Run services locally** (in separate terminals):
   ```bash
   # Terminal 1: Ingestion API
   make run-ingestion

   # Terminal 2: Metrics Processor
   make run-processor

   # Terminal 3: Metrics API
   make run-metrics-api
   ```

### Build Binaries

```bash
make build

# Binaries will be in ./bin/
./bin/ingestion-api
./bin/metrics-processor
./bin/metrics-api
```

## 📡 API Reference

### Ingestion API (Port 8080)

#### POST `/v1/llm/request`
Ingest an LLM request event.

**Request Body**:
```json
{
  "request_id": "uuid",
  "tenant_id": "acme-corp",
  "route": "chat_support_v2",
  "model": "gpt-4.1-mini",
  "timestamp": "2025-11-19T10:00:00.000Z",
  "prompt_tokens": 321,
  "user_id_hash": "sha256-hash",
  "metadata": {
    "experiment": "prompt_v3",
    "country": "SE"
  }
}
```

#### POST `/v1/llm/response`
Ingest an LLM response event.

**Request Body**:
```json
{
  "request_id": "uuid",
  "timestamp": "2025-11-19T10:00:01.123Z",
  "latency_ms": 1123,
  "completion_tokens": 512,
  "finish_reason": "stop",
  "error": null
}
```

### Metrics API (Port 8081)

#### GET `/v1/metrics`
Query aggregated metrics.

**Query Parameters**:
- `tenant_id` (required): Filter by tenant
- `route` (optional): Filter by route
- `model` (optional): Filter by model
- `limit` (optional): Number of windows to return (default: 60)

**Response**:
```json
{
  "metrics": [
    {
      "tenant_id": "acme-corp",
      "route": "chat_support_v2",
      "model": "gpt-4.1-mini",
      "window_start": "2025-11-19T10:00:00Z",
      "window_end": "2025-11-19T10:01:00Z",
      "requests": 1234,
      "errors": 12,
      "avg_latency_ms": 732.4,
      "p95_latency_ms": 1234.0,
      "avg_prompt_tokens": 300.1,
      "avg_completion_tokens": 420.6,
      "estimated_cost_usd": 2.31
    }
  ],
  "count": 1
}
```

## 🔧 Configuration

All services are configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `KAFKA_BROKERS` | Comma-separated Kafka broker addresses | `localhost:19092` |
| `POSTGRES_DSN` | PostgreSQL connection string | `postgres://streamlens:streamlens@localhost:5432/streamlens?sslmode=disable` |
| `HTTP_PORT` | HTTP server port | `8080` (ingestion), `8081` (metrics) |
| `CONSUMER_GROUP` | Kafka consumer group name | `metrics-processor-group` |

## 📂 Project Structure

```
streamlens/
├── cmd/
│   ├── ingestion-api/       # HTTP ingestion service
│   ├── metrics-processor/   # Stream processor
│   └── metrics-api/         # HTTP metrics query service
├── internal/
│   ├── config/              # Configuration management
│   ├── handlers/            # HTTP handlers
│   ├── kafka/               # Kafka producer/consumer wrappers
│   ├── models/              # Event schemas
│   ├── processor/           # Stream processing logic
│   └── store/               # Postgres storage layer
├── deploy/
│   ├── docker-compose.yml   # Docker Compose config
│   ├── Dockerfile           # Multi-service Dockerfile
│   └── init.sql             # Postgres schema
├── scripts/
│   └── generate-traffic.sh  # Sample data generator
├── Makefile                 # Development tasks
└── README.md
```

## 🧪 Testing

```bash
# Run all tests
make test

# Test specific package
go test -v ./internal/processor
```

## 🎯 Key Features

- **Stream Processing**: Real-time joining of requests/responses by `request_id`
- **Windowed Aggregation**: 1-minute tumbling windows per (tenant, route, model)
- **Metrics Computation**:
  - Request count
  - Error count
  - Average & P95 latency
  - Token usage (prompt/completion)
  - Estimated cost
- **Dual Sink**: Metrics written to both Kafka topic and Postgres
- **Graceful Shutdown**: All services handle SIGTERM/SIGINT correctly
- **Production-Ready**: Proper error handling, logging, and connection pooling

## 🔍 Monitoring

### Health Checks

```bash
# Ingestion API
curl http://localhost:8080/health

# Metrics API
curl http://localhost:8081/health
```

### Redpanda Console

Access Redpanda metrics and topics at:
- Admin API: http://localhost:19644

### View Logs

```bash
# All services
docker compose -f deploy/docker-compose.yml logs -f

# Specific service
docker compose -f deploy/docker-compose.yml logs -f metrics-processor
```

## 🧹 Cleanup

```bash
# Stop services
make docker-down

# Stop and remove volumes
make clean
```

## 🎬 Demo

For a quick demo, follow the [Quick Start Guide](QUICKSTART.md):

```bash
# Start all services
make docker-up

# Generate sample traffic
make generate-traffic

# Query metrics (wait ~2 minutes for windowing)
curl 'http://localhost:8081/v1/metrics?tenant_id=tenant-1' | jq
```

**Sample Output:**
```json
{
  "metrics": [
    {
      "tenant_id": "tenant-1",
      "route": "chat_support_v1",
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
  ]
}
```

## 🚧 Future Enhancements

- [ ] Add `llm.evaluations` topic support
- [ ] Implement distributed tracing (OpenTelemetry)
- [ ] Add Prometheus metrics endpoint
- [ ] Dashboard frontend (Grafana/React)
- [ ] Multi-region deployment support
- [ ] Schema registry integration
- [ ] Rate limiting on ingestion API
- [ ] Backfilling for historical data

## 🤝 Contributing

Contributions are welcome! Please check out our [Contributing Guide](CONTRIBUTING.md) to get started.

- 🐛 [Report a bug](../../issues/new?labels=bug)
- 💡 [Request a feature](../../issues/new?labels=enhancement)
- 📖 [Improve documentation](CONTRIBUTING.md)

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Redpanda](https://redpanda.com/) - Kafka-compatible streaming platform
- [franz-go](https://github.com/twmb/franz-go) - High-performance Kafka client for Go
- [OpenLLMetry](https://github.com/traceloop/openllmetry) - OpenTelemetry for LLMs

## 📧 Contact

For questions, discussions, or feedback:
- Open a [GitHub Issue](../../issues)
- Start a [Discussion](../../discussions)

---

**Suggested GitHub Topics:** `golang` `kafka` `observability` `llm` `stream-processing` `real-time` `metrics` `microservices` `redpanda` `postgresql`
