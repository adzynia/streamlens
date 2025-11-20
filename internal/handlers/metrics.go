package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"streamlens/internal/store"
	"time"
)

// MetricsHandler handles metrics query requests
type MetricsHandler struct {
	store *store.MetricsStore
}

// NewMetricsHandler creates a new MetricsHandler
func NewMetricsHandler(store *store.MetricsStore) *MetricsHandler {
	return &MetricsHandler{store: store}
}

// HandleGetMetrics handles GET /v1/metrics
func (h *MetricsHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	route := r.URL.Query().Get("route")
	model := r.URL.Query().Get("model")
	limitStr := r.URL.Query().Get("limit")

	// Parse optional limit
	limit := 60 // default
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	// Convert to pointers for optional params
	var routePtr, modelPtr *string
	if route != "" {
		routePtr = &route
	}
	if model != "" {
		modelPtr = &model
	}

	// Query database with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	metrics, err := h.store.QueryMetrics(ctx, tenantID, routePtr, modelPtr, limit)
	if err != nil {
		log.Printf("Failed to query metrics: %v", err)
		http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
		return
	}

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

// HandleHealth handles GET /health
func (h *MetricsHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
