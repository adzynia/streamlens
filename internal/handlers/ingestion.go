package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"streamlens/internal/kafka"
	"streamlens/internal/models"
	"time"
)

// IngestionHandler handles LLM telemetry ingestion
type IngestionHandler struct {
	producer *kafka.Producer
}

// NewIngestionHandler creates a new IngestionHandler
func NewIngestionHandler(producer *kafka.Producer) *IngestionHandler {
	return &IngestionHandler{producer: producer}
}

// HandleLLMRequest handles POST /v1/llm/request
func (h *IngestionHandler) HandleLLMRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LLMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := req.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Produce to Kafka with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.producer.ProduceJSON(ctx, kafka.TopicLLMRequests, req.RequestID, &req); err != nil {
		log.Printf("Failed to produce request: %v", err)
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "accepted"}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// HandleLLMResponse handles POST /v1/llm/response
func (h *IngestionHandler) HandleLLMResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resp models.LLMResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := resp.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Produce to Kafka with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.producer.ProduceJSON(ctx, kafka.TopicLLMResponses, resp.RequestID, &resp); err != nil {
		log.Printf("Failed to produce response: %v", err)
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "accepted"}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// HandleHealth handles GET /health
func (h *IngestionHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"}); err != nil {
		log.Printf("Failed to encode health response: %v", err)
	}
}
