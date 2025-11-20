# StreamLens Architecture

This document provides a deep dive into StreamLens' architecture, design decisions, and implementation details.

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Component Details](#component-details)
3. [Data Flow](#data-flow)
4. [Stream Processing](#stream-processing)
5. [Design Decisions](#design-decisions)
6. [Production Considerations](#production-considerations)

---

## High-Level Architecture

StreamLens follows an event-driven, microservices architecture:

```
┌────────────────────────────────────────────────────────┐
│                     LLM Applications                   │
└───────────┬────────────────────────────────────────────┘
            │ POST /v1/llm/request
            │ POST /v1/llm/response
            ▼
┌───────────────────────────────────────────────────────┐
│              Ingestion API (Port 8080)                │
│  • Validates events                                   │
│  • Produces to Kafka topics                          │
└───────────┬───────────────────────────────────────────┘
            │
            ▼
┌───────────────────────────────────────────────────────┐
│                     Redpanda                          │
│  Topics:                                              │
│    • llm.requests  (key: request_id)                 │
│    • llm.responses (key: request_id)                 │
│    • llm.metrics   (key: tenant|route|model)         │
└───────────┬───────────────────────────────────────────┘
            │
            ▼
┌───────────────────────────────────────────────────────┐
│         Metrics Processor (Consumer Group)            │
│  • Consumes requests & responses                      │
│  • Joins by request_id (in-memory state)            │
│  • Aggregates into 1-min windows                     │
│  • Computes metrics (avg, p95, cost, etc.)          │
└───────────┬───────────────────────────────────────────┘
            │
            ├─────────────────┬─────────────────────────┐
            ▼                 ▼                         ▼
    ┌───────────────┐ ┌──────────────┐     ┌──────────────────┐
    │   Postgres    │ │   Redpanda   │     │  (Future: S3,    │
    │ llm_metrics   │ │ llm.metrics  │     │   DataWarehouse) │
    │     table     │ │    topic     │     └──────────────────┘
    └───────┬───────┘ └──────────────┘
            │
            ▼
┌───────────────────────────────────────────────────────┐
│              Metrics API (Port 8081)                  │
│  • Queries Postgres                                   │
│  • Serves aggregated metrics                         │
│  • Filters by tenant, route, model                   │
└───────────────────────────────────────────────────────┘
```

---

## Component Details

### 1. Ingestion API

**Purpose**: Ingest LLM telemetry events from applications

**Technology**: Go HTTP service using `chi` router

**Key responsibilities**:
- Validate incoming JSON payloads
- Check required fields (request_id, tenant_id, route, model, timestamp)
- Produce events to appropriate Kafka topics with correct keys
- Handle errors and timeouts gracefully
- Provide health check endpoint

**Endpoints**:
- `POST /v1/llm/request` - Ingest request events
- `POST /v1/llm/response` - Ingest response events
- `GET /health` - Health check

**Configuration**:
- `KAFKA_BROKERS` - Kafka broker addresses
- `HTTP_PORT` - Server port (default: 8080)

**Scalability**: Stateless - can be horizontally scaled behind a load balancer

---

### 2. Metrics Processor

**Purpose**: Stream process events to compute real-time metrics

**Technology**: Go service using franz-go Kafka client

**Key responsibilities**:
- Consume from `llm.requests` and `llm.responses` topics
- Join request/response pairs by `request_id` (in-memory state store)
- Aggregate events into 1-minute tumbling windows
- Compute metrics: count, errors, avg/p95 latency, tokens, cost
- Produce aggregated metrics to `llm.metrics` topic
- Write metrics to Postgres for querying
- Clean up old state to prevent memory leaks

**Architecture**:
```
┌─────────────────────────────────────────────────────┐
│           Metrics Processor Process                 │
│                                                     │
│  ┌──────────────────────────────────────────────┐ │
│  │  Consumer (Consumer Group)                   │ │
│  │  • Reads llm.requests & llm.responses        │ │
│  └───────────────┬──────────────────────────────┘ │
│                  │                                 │
│  ┌───────────────▼──────────────────────────────┐ │
│  │  In-Memory State Store                       │ │
│  │  • requestState: map[request_id]Request      │ │
│  │  • responseState: map[request_id]Response    │ │
│  │  • TTL: 5 minutes                            │ │
│  └───────────────┬──────────────────────────────┘ │
│                  │                                 │
│  ┌───────────────▼──────────────────────────────┐ │
│  │  Window Aggregates                           │ │
│  │  • Key: tenant|route|model|window_start      │ │
│  │  • Value: WindowAggregate (counts, arrays)   │ │
│  │  • Flushed every minute                      │ │
│  └───────────────┬──────────────────────────────┘ │
│                  │                                 │
│       ┌──────────┴──────────┐                     │
│       ▼                     ▼                     │
│  ┌─────────┐         ┌──────────┐                │
│  │Producer │         │ Postgres │                │
│  │(Kafka)  │         │  Store   │                │
│  └─────────┘         └──────────┘                │
└─────────────────────────────────────────────────────┘
```

**Window Processing**:
- **Type**: Tumbling windows (non-overlapping)
- **Duration**: 1 minute
- **Trigger**: Time-based (every 60 seconds)
- **Aggregation Key**: `{tenant_id, route, model}`

**Metrics Computed**:
- Request count
- Error count (responses with non-null error field)
- Average latency (mean of all latencies)
- P95 latency (95th percentile)
- Average prompt tokens
- Average completion tokens
- Estimated cost (configurable pricing model)

**Scalability**: 
- Single instance (consumer group with 1 member)
- For scaling: increase topic partitions and consumer instances
- State is currently in-memory (could be externalized to Redis/RocksDB)

---

### 3. Metrics API

**Purpose**: Query aggregated metrics for dashboards

**Technology**: Go HTTP service using `chi` router

**Key responsibilities**:
- Query Postgres for metrics
- Filter by tenant_id (required), route, model (optional)
- Return last N windows of data
- Provide health check endpoint

**Endpoints**:
- `GET /v1/metrics?tenant_id=X&route=Y&model=Z&limit=N`
- `GET /health`

**Query Performance**:
- Indexed by (tenant_id, route, model, window_start)
- Descending order by window_start for recent data
- Configurable limit (default: 60 windows = 1 hour)

**Scalability**: Stateless - can be horizontally scaled

---

### 4. Redpanda (Kafka)

**Purpose**: Event streaming backbone

**Topics**:

| Topic | Key | Value | Purpose |
|-------|-----|-------|---------|
| `llm.requests` | request_id | LLMRequest JSON | Inbound request events |
| `llm.responses` | request_id | LLMResponse JSON | Inbound response events |
| `llm.metrics` | tenant\|route\|model | LLMMetrics JSON | Computed metrics (output) |

**Configuration**:
- Single broker in dev (can be clustered in production)
- Auto-create topics enabled
- Kafka-compatible API on port 19092

---

### 5. PostgreSQL

**Purpose**: Persistent storage for metrics

**Schema**:
```sql
CREATE TABLE llm_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    route VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    window_start TIMESTAMP NOT NULL,
    window_end TIMESTAMP NOT NULL,
    requests INTEGER NOT NULL DEFAULT 0,
    errors INTEGER NOT NULL DEFAULT 0,
    avg_latency_ms DOUBLE PRECISION,
    p95_latency_ms DOUBLE PRECISION,
    avg_prompt_tokens DOUBLE PRECISION,
    avg_completion_tokens DOUBLE PRECISION,
    estimated_cost_usd DOUBLE PRECISION,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, route, model, window_start)
);

CREATE INDEX idx_llm_metrics_tenant_time 
    ON llm_metrics(tenant_id, window_start DESC);
    
CREATE INDEX idx_llm_metrics_composite 
    ON llm_metrics(tenant_id, route, model, window_start DESC);
```

**Data Retention**: 
- Not implemented yet
- Recommend partitioning by window_start for efficient pruning
- Could archive old data to S3/data warehouse

---

## Data Flow

### Request/Response Lifecycle

1. **Application sends request event**:
   ```
   POST /v1/llm/request
   {
     "request_id": "req-123",
     "tenant_id": "acme",
     "route": "chat",
     "model": "gpt-4",
     "timestamp": "2025-11-20T10:00:30.000Z",
     "prompt_tokens": 100
   }
   ```

2. **Ingestion API validates and produces to Kafka**:
   ```
   Topic: llm.requests
   Key: "req-123"
   Value: { ... }
   ```

3. **Application sends response event** (after LLM returns):
   ```
   POST /v1/llm/response
   {
     "request_id": "req-123",
     "timestamp": "2025-11-20T10:00:31.200Z",
     "latency_ms": 1200,
     "completion_tokens": 250,
     "finish_reason": "stop"
   }
   ```

4. **Metrics Processor consumes both events**:
   - Stores request in state: `requestState["req-123"] = {...}`
   - Stores response in state: `responseState["req-123"] = {...}`
   - Joins them by request_id
   - Aggregates into window: `acme|chat|gpt-4|2025-11-20T10:00:00`

5. **Window closes** (at 10:01:00):
   - Processor computes final metrics from aggregate
   - Produces to `llm.metrics` topic
   - Inserts into Postgres `llm_metrics` table

6. **Dashboard queries metrics**:
   ```
   GET /v1/metrics?tenant_id=acme
   
   Returns: [{
     "tenant_id": "acme",
     "route": "chat",
     "model": "gpt-4",
     "window_start": "2025-11-20T10:00:00Z",
     "window_end": "2025-11-20T10:01:00Z",
     "requests": 45,
     "errors": 2,
     "avg_latency_ms": 1150.5,
     "p95_latency_ms": 2300.0,
     ...
   }]
   ```

---

## Stream Processing

### Join Strategy

**Problem**: Match requests with their responses

**Solution**: In-memory state store with TTL

**Implementation**:
```go
type MetricsProcessor struct {
    requestState  map[string]*LLMRequest   // keyed by request_id
    responseState map[string]*LLMResponse  // keyed by request_id
    stateMu       sync.RWMutex             // thread-safe access
}
```

**Behavior**:
- When request arrives: store in `requestState`
- When response arrives: store in `responseState`, check for request
- If both exist: join and aggregate
- If only one exists: wait for the other (up to 5 minutes)
- Background cleanup: remove entries older than 5 minutes

**Trade-offs**:
- ✅ Simple, fast, no external dependencies
- ✅ Works well for high-throughput, low-latency scenarios
- ❌ State lost on restart (could be solved with state store)
- ❌ Memory usage grows with long-lived requests

---

### Windowing Strategy

**Type**: Tumbling windows (fixed, non-overlapping)

**Duration**: 60 seconds

**Alignment**: Truncated to minute boundaries (e.g., 10:00:00, 10:01:00)

**Trigger**: Time-based (wall clock every 60 seconds)

**Late Events**: Not handled (would require watermarking)

**Implementation**:
```go
// Window key includes timestamp
key := fmt.Sprintf("%s|%s|%s|%d", 
    tenantID, route, model, windowStart.Unix())

// Store aggregates per window
windowAggregates[key] = &WindowAggregate{
    WindowStart: windowStart,
    WindowEnd:   windowStart.Add(1 * time.Minute),
    Requests:    []...,
    Latencies:   []int{},
    ...
}

// Flush windows older than 1 minute
if agg.WindowEnd.Before(now.Add(-1 * time.Minute)) {
    computeAndEmit(agg)
    delete(windowAggregates, key)
}
```

---

## Design Decisions

### Why Redpanda (Kafka)?

- ✅ Kafka-compatible API (standard tooling works)
- ✅ Easier to deploy than Apache Kafka
- ✅ High throughput, low latency
- ✅ Built-in replication and durability

### Why franz-go?

- ✅ Modern, idiomatic Go client
- ✅ High performance
- ✅ Good error handling and context support
- ✅ Active maintenance

### Why PostgreSQL?

- ✅ Reliable, well-understood
- ✅ Good query performance with indexes
- ✅ ACID guarantees
- ✅ Easy to operate

### Why In-Memory State?

- ✅ Simplicity for MVP
- ✅ Fast access
- ❌ Not production-ready (state lost on restart)
- 🔧 Future: RocksDB or Redis for state

### Why 1-Minute Windows?

- ✅ Good balance between freshness and overhead
- ✅ Matches common observability tools (Grafana, Datadog)
- ✅ Reduces database write rate
- 🔧 Could be made configurable

---

## Production Considerations

### Scaling

**Ingestion API**:
- Horizontal: Add more instances behind load balancer
- Vertical: Increase CPU/memory for higher throughput

**Metrics Processor**:
- Horizontal: Increase topic partitions, add consumer instances
- State: Move to external store (Redis, RocksDB)
- Checkpointing: Commit offsets after successful writes

**Metrics API**:
- Horizontal: Add read replicas for Postgres
- Caching: Add Redis cache for frequently queried metrics
- CDN: Cache results at edge for public dashboards

### Monitoring

**Add**:
- Prometheus metrics on each service
- OpenTelemetry tracing
- Custom metrics: processing lag, window size, error rates
- Alerting: consumer lag > threshold, error rate spike

### Reliability

**Current**:
- ✅ Graceful shutdown on SIGTERM
- ✅ Error handling and logging
- ✅ Connection pooling

**Improvements**:
- Circuit breakers for downstream services
- Retry logic with exponential backoff
- Dead letter queue for failed messages
- Health checks in Docker Compose

### Security

**Add**:
- TLS for Kafka connections
- Authentication/authorization on APIs (JWT, API keys)
- Rate limiting on ingestion API
- Input sanitization and validation
- Network policies (service mesh)

### Data Quality

**Add**:
- Schema validation (Avro, Protobuf)
- Data quality checks (missing fields, outliers)
- Deduplication (idempotent processing)
- Backfilling for historical data

### Cost

**Current**: Minimal (development setup)

**Production**:
- Redpanda: ~$500-2000/month (cluster with replication)
- Postgres: ~$200-1000/month (managed service)
- Compute: ~$300-1500/month (3 services with autoscaling)
- Storage: ~$50-500/month (depends on retention)

**Total**: ~$1050-5000/month for production setup

---

## Future Enhancements

1. **State Management**: Externalize to RocksDB/Redis
2. **Schema Registry**: Add Avro/Protobuf for type safety
3. **Exactly-Once Semantics**: Kafka transactions for duplicate handling
4. **Distributed Tracing**: OpenTelemetry for request tracking
5. **Dashboard**: React/Vue frontend with charts
6. **Alerting**: Real-time alerts on metrics thresholds
7. **Multi-Region**: Active-active or active-passive setup
8. **Data Warehouse**: Export to Snowflake/BigQuery for analytics

---

For more details, see the [README.md](README.md) and [QUICKSTART.md](QUICKSTART.md).
