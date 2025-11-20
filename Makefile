.PHONY: help deps build run-ingestion run-processor run-metrics-api test docker-up docker-down docker-build clean generate-traffic

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	go mod download
	go mod tidy

build: ## Build all services
	@echo "Building ingestion-api..."
	go build -o bin/ingestion-api ./cmd/ingestion-api
	@echo "Building metrics-processor..."
	go build -o bin/metrics-processor ./cmd/metrics-processor
	@echo "Building metrics-api..."
	go build -o bin/metrics-api ./cmd/metrics-api
	@echo "Build complete!"

run-ingestion: ## Run ingestion API locally
	@echo "Starting Ingestion API..."
	go run ./cmd/ingestion-api/main.go

run-processor: ## Run metrics processor locally
	@echo "Starting Metrics Processor..."
	go run ./cmd/metrics-processor/main.go

run-metrics-api: ## Run metrics API locally
	@echo "Starting Metrics API..."
	HTTP_PORT=8081 go run ./cmd/metrics-api/main.go

test: ## Run tests
	go test -v ./...

docker-build: ## Build Docker images
	docker compose -f deploy/docker-compose.yml build

docker-up: ## Start all services with Docker Compose
	docker compose -f deploy/docker-compose.yml up -d

docker-down: ## Stop all services
	docker compose -f deploy/docker-compose.yml down

docker-logs: ## View logs from all services
	docker compose -f deploy/docker-compose.yml logs -f

clean: ## Clean build artifacts
	rm -rf bin/
	docker compose -f deploy/docker-compose.yml down -v

generate-traffic: ## Generate sample traffic (requires services to be running)
	@echo "Generating sample LLM traffic..."
	./scripts/generate-traffic.sh
