#!/bin/bash

# Script to generate sample LLM traffic for testing

BASE_URL="${BASE_URL:-http://localhost:8080}"
NUM_REQUESTS="${NUM_REQUESTS:-20}"

echo "Generating $NUM_REQUESTS sample requests to $BASE_URL"

for i in $(seq 1 $NUM_REQUESTS); do
  REQUEST_ID="req-$(uuidgen | tr '[:upper:]' '[:lower:]')"
  TENANT_ID="tenant-$(( RANDOM % 3 + 1 ))"
  ROUTE="chat_support_v$(( RANDOM % 2 + 1 ))"
  MODEL="gpt-4.1-mini"
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")
  PROMPT_TOKENS=$(( RANDOM % 500 + 100 ))
  
  echo "[$i/$NUM_REQUESTS] Sending request: $REQUEST_ID"
  
  # Send request event
  REQUEST_PAYLOAD=$(cat <<EOF
{
  "request_id": "$REQUEST_ID",
  "tenant_id": "$TENANT_ID",
  "route": "$ROUTE",
  "model": "$MODEL",
  "timestamp": "$TIMESTAMP",
  "prompt_tokens": $PROMPT_TOKENS,
  "metadata": {
    "experiment": "test_traffic",
    "country": "US"
  }
}
EOF
)
  
  curl -s -X POST "$BASE_URL/v1/llm/request" \
    -H "Content-Type: application/json" \
    -d "$REQUEST_PAYLOAD" > /dev/null
  
  # Simulate processing delay
  sleep 0.$(( RANDOM % 9 + 1 ))
  
  # Send response event
  LATENCY_MS=$(( RANDOM % 2000 + 500 ))
  COMPLETION_TOKENS=$(( RANDOM % 800 + 200 ))
  ERROR_RATE=$(( RANDOM % 100 ))
  
  if [ $ERROR_RATE -lt 5 ]; then
    # 5% error rate
    RESPONSE_PAYLOAD=$(cat <<EOF
{
  "request_id": "$REQUEST_ID",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")",
  "latency_ms": $LATENCY_MS,
  "completion_tokens": $COMPLETION_TOKENS,
  "finish_reason": "error",
  "error": "rate_limit_exceeded"
}
EOF
)
  else
    RESPONSE_PAYLOAD=$(cat <<EOF
{
  "request_id": "$REQUEST_ID",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")",
  "latency_ms": $LATENCY_MS,
  "completion_tokens": $COMPLETION_TOKENS,
  "finish_reason": "stop"
}
EOF
)
  fi
  
  curl -s -X POST "$BASE_URL/v1/llm/response" \
    -H "Content-Type: application/json" \
    -d "$RESPONSE_PAYLOAD" > /dev/null
  
  sleep 0.2
done

echo ""
echo "✅ Generated $NUM_REQUESTS request/response pairs"
echo ""
echo "Wait ~2 minutes for metrics to aggregate, then query:"
echo "  curl 'http://localhost:8081/v1/metrics?tenant_id=tenant-1&limit=10'"
