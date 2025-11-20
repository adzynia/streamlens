# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-20

### Added
- Initial release of StreamLens LLM observability platform
- Ingestion API service for receiving LLM request/response events
- Metrics Processor for stream processing and aggregation
- Metrics API for querying aggregated metrics
- Real-time stream processing with Redpanda (Kafka-compatible)
- 1-minute tumbling window aggregations per (tenant, route, model)
- PostgreSQL storage for metrics with proper indexing
- Dual-sink architecture (Kafka topic + Postgres)
- Docker Compose setup for easy local deployment
- Comprehensive documentation (README, ARCHITECTURE, QUICKSTART)
- Sample traffic generator script
- Makefile for common development tasks
- Health check endpoints for all services
- Graceful shutdown handling

### Features
- **Metrics Computed:**
  - Request count per window
  - Error count and rate
  - Average and P95 latency
  - Average prompt tokens
  - Average completion tokens
  - Estimated cost (USD)
- **Event Schema Support:**
  - LLM request events with metadata
  - LLM response events with timing and tokens
  - Flexible metadata fields for custom tagging
- **Query API:**
  - Filter by tenant, route, and model
  - Configurable result limits
  - Time-windowed data retrieval
- **Production-Ready:**
  - Proper error handling and logging
  - Connection pooling for Postgres
  - Consumer group support for horizontal scaling
  - Race condition handling in stream processing

### Infrastructure
- Multi-service Docker Compose configuration
- Redpanda with proper port mappings
- PostgreSQL with automatic schema initialization
- Health checks for all containers
- Volume management for persistent data

[1.0.0]: https://github.com/YOUR_USERNAME/streamlens/releases/tag/v1.0.0
