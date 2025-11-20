#!/bin/bash

# Script to test the APIs manually

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
METRICS_URL="${METRICS_URL:-http://localhost:8081}"

echo "🧪 Testing StreamLens APIs"
echo "=========================="
echo ""

# Test health endpoints
echo "1. Testing health endpoints..."
echo "   Ingestion API:"
curl -s "$BASE_URL/health" | jq '.'
echo ""
echo "   Metrics API:"
curl -s "$METRICS_URL/health" | jq '.'
echo ""

# Test ingestion of a request
REQUEST_ID="test-$(uuidgen | tr '[:upper:]' '[:lower:]')"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")

echo "2. Sending LLM Request (request_id: $REQUEST_ID)..."
curl -s -X POST "$BASE_URL/v1/llm/request" \
  -H "Content-Type: application/json" \
  -d "{
    \"request_id\": \"$REQUEST_ID\",
    \"tenant_id\": \"test-tenant\",
    \"route\": \"chat_api\",
    \"model\": \"gpt-4.1-mini\",
    \"timestamp\": \"$TIMESTAMP\",
    \"prompt_tokens\": 250,
    \"metadata\": {
      \"test\": true
    }
  }" | jq '.'
echo ""

# Test ingestion of a response
sleep 1
echo "3. Sending LLM Response (request_id: $REQUEST_ID)..."
curl -s -X POST "$BASE_URL/v1/llm/response" \
  -H "Content-Type: application/json" \
  -d "{
    \"request_id\": \"$REQUEST_ID\",
    \"timestamp\": \"$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")\",
    \"latency_ms\": 850,
    \"completion_tokens\": 400,
    \"finish_reason\": \"stop\"
  }" | jq '.'
echo ""

# Query metrics
echo "4. Querying metrics for test-tenant..."
echo "   (Note: May be empty if no windows have been flushed yet)"
curl -s "$METRICS_URL/v1/metrics?tenant_id=test-tenant&limit=5" | jq '.'
echo ""

echo "✅ API tests complete!"
echo ""
echo "💡 Tip: Run 'make generate-traffic' to create more realistic test data"
