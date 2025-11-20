package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"streamlens/internal/kafka"
	"streamlens/internal/models"
	"streamlens/internal/store"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	// WindowDuration is the size of the tumbling window (1 minute)
	WindowDuration = 1 * time.Minute
	// StateRetentionDuration is how long to keep request/response pairs in memory
	StateRetentionDuration = 5 * time.Minute
)

// MetricsProcessor handles stream processing of LLM events
type MetricsProcessor struct {
	consumer *kafka.Consumer
	producer *kafka.Producer
	store    *store.MetricsStore

	// In-memory state for joining requests and responses
	requestState  map[string]*models.LLMRequest
	responseState map[string]*models.LLMResponse
	stateMu       sync.RWMutex

	// Windowed aggregation state
	windowAggregates map[string]*WindowAggregate
	windowMu         sync.RWMutex

	// Ticker for window processing
	windowTicker *time.Ticker
}

// WindowAggregate holds aggregated data for a time window
type WindowAggregate struct {
	TenantID    string
	Route       string
	Model       string
	WindowStart time.Time
	WindowEnd   time.Time

	Requests     int
	Errors       int
	Latencies    []int // For percentile calculation
	PromptTokens []int
	CompTokens   []int
}

// NewMetricsProcessor creates a new metrics processor
func NewMetricsProcessor(consumer *kafka.Consumer, producer *kafka.Producer, store *store.MetricsStore) *MetricsProcessor {
	return &MetricsProcessor{
		consumer:         consumer,
		producer:         producer,
		store:            store,
		requestState:     make(map[string]*models.LLMRequest),
		responseState:    make(map[string]*models.LLMResponse),
		windowAggregates: make(map[string]*WindowAggregate),
		windowTicker:     time.NewTicker(WindowDuration),
	}
}

// Run starts the metrics processor
func (p *MetricsProcessor) Run(ctx context.Context) error {
	log.Println("Starting metrics processor...")

	// Start background goroutine for window processing
	go p.processWindows(ctx)

	// Start background goroutine for state cleanup
	go p.cleanupState(ctx)

	// Main consume loop
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping processor...")
			return ctx.Err()
		default:
			fetches := p.consumer.Poll(ctx)

			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					log.Printf("Fetch error: %v", err.Err)
				}
				continue
			}

			// Process records
			var recordsToCommit []*kgo.Record
			fetches.EachRecord(func(record *kgo.Record) {
				if err := p.processRecord(ctx, record); err != nil {
					log.Printf("Error processing record: %v", err)
				} else {
					recordsToCommit = append(recordsToCommit, record)
				}
			})

			// Commit offsets
			if len(recordsToCommit) > 0 {
				if err := p.consumer.CommitRecords(ctx, recordsToCommit...); err != nil {
					log.Printf("Failed to commit offsets: %v", err)
				}
			}
		}
	}
}

// processRecord handles a single Kafka record
func (p *MetricsProcessor) processRecord(ctx context.Context, record *kgo.Record) error {
	switch record.Topic {
	case kafka.TopicLLMRequests:
		return p.processRequest(record)
	case kafka.TopicLLMResponses:
		return p.processResponse(ctx, record)
	default:
		return fmt.Errorf("unknown topic: %s", record.Topic)
	}
}

// processRequest stores a request in state
func (p *MetricsProcessor) processRequest(record *kgo.Record) error {
	var req models.LLMRequest
	if err := json.Unmarshal(record.Value, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	p.stateMu.Lock()
	p.requestState[req.RequestID] = &req
	p.stateMu.Unlock()

	return nil
}

// processResponse joins with request and aggregates into windows
func (p *MetricsProcessor) processResponse(ctx context.Context, record *kgo.Record) error {
	var resp models.LLMResponse
	if err := json.Unmarshal(record.Value, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Store response
	p.stateMu.Lock()
	p.responseState[resp.RequestID] = &resp
	p.stateMu.Unlock()

	// Try to join with request
	p.stateMu.RLock()
	req, found := p.requestState[resp.RequestID]
	p.stateMu.RUnlock()

	if !found {
		// Request not found yet - response will be processed when window closes
		return nil
	}

	// Aggregate into window
	p.aggregateEvent(req, &resp)

	return nil
}

// aggregateEvent adds an event to the appropriate window
func (p *MetricsProcessor) aggregateEvent(req *models.LLMRequest, resp *models.LLMResponse) {
	// Calculate window boundaries
	windowStart := req.Timestamp.Truncate(WindowDuration)
	windowEnd := windowStart.Add(WindowDuration)

	// Create window key
	key := fmt.Sprintf("%s|%s|%s|%d", req.TenantID, req.Route, req.Model, windowStart.Unix())

	p.windowMu.Lock()
	defer p.windowMu.Unlock()

	agg, exists := p.windowAggregates[key]
	if !exists {
		agg = &WindowAggregate{
			TenantID:     req.TenantID,
			Route:        req.Route,
			Model:        req.Model,
			WindowStart:  windowStart,
			WindowEnd:    windowEnd,
			Latencies:    []int{},
			PromptTokens: []int{},
			CompTokens:   []int{},
		}
		p.windowAggregates[key] = agg
	}

	// Update aggregates
	agg.Requests++

	if resp.Error != nil && *resp.Error != "" {
		agg.Errors++
	}

	agg.Latencies = append(agg.Latencies, resp.LatencyMs)
	agg.PromptTokens = append(agg.PromptTokens, req.PromptTokens)
	agg.CompTokens = append(agg.CompTokens, resp.CompletionTokens)
}

// processWindows periodically flushes completed windows
func (p *MetricsProcessor) processWindows(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.windowTicker.C:
			p.flushCompletedWindows(ctx)
		}
	}
}

// flushCompletedWindows writes completed windows to Kafka and Postgres
func (p *MetricsProcessor) flushCompletedWindows(ctx context.Context) {
	now := time.Now()
	cutoff := now.Add(-WindowDuration) // Windows older than 1 minute

	p.windowMu.Lock()
	defer p.windowMu.Unlock()

	for key, agg := range p.windowAggregates {
		if agg.WindowEnd.Before(cutoff) {
			// Compute final metrics
			metrics := p.computeMetrics(agg)

			// Produce to Kafka
			metricsKey := fmt.Sprintf("%s|%s|%s", metrics.TenantID, metrics.Route, metrics.Model)
			if err := p.producer.ProduceJSON(ctx, kafka.TopicLLMMetrics, metricsKey, metrics); err != nil {
				log.Printf("Failed to produce metrics: %v", err)
				continue
			}

			// Write to Postgres
			if err := p.store.InsertMetrics(ctx, metrics); err != nil {
				log.Printf("Failed to insert metrics to DB: %v", err)
				continue
			}

			log.Printf("Flushed window: %s [%s to %s] - %d requests, %d errors",
				key, agg.WindowStart.Format(time.RFC3339), agg.WindowEnd.Format(time.RFC3339),
				agg.Requests, agg.Errors)

			// Remove from memory
			delete(p.windowAggregates, key)
		}
	}
}

// computeMetrics calculates final metrics from aggregate
func (p *MetricsProcessor) computeMetrics(agg *WindowAggregate) *models.LLMMetrics {
	metrics := &models.LLMMetrics{
		TenantID:    agg.TenantID,
		Route:       agg.Route,
		Model:       agg.Model,
		WindowStart: agg.WindowStart,
		WindowEnd:   agg.WindowEnd,
		Requests:    agg.Requests,
		Errors:      agg.Errors,
	}

	// Calculate average latency
	if len(agg.Latencies) > 0 {
		sum := 0
		for _, l := range agg.Latencies {
			sum += l
		}
		metrics.AvgLatencyMs = float64(sum) / float64(len(agg.Latencies))

		// Calculate P95 latency
		metrics.P95LatencyMs = calculatePercentile(agg.Latencies, 0.95)
	}

	// Calculate average tokens
	if len(agg.PromptTokens) > 0 {
		sum := 0
		for _, t := range agg.PromptTokens {
			sum += t
		}
		metrics.AvgPromptTokens = float64(sum) / float64(len(agg.PromptTokens))
	}

	if len(agg.CompTokens) > 0 {
		sum := 0
		for _, t := range agg.CompTokens {
			sum += t
		}
		metrics.AvgCompletionTokens = float64(sum) / float64(len(agg.CompTokens))
	}

	// Estimate cost (simple model: $0.01 per 1000 prompt tokens, $0.03 per 1000 completion tokens)
	metrics.EstimatedCostUSD = (metrics.AvgPromptTokens * 0.01 / 1000.0 * float64(agg.Requests)) +
		(metrics.AvgCompletionTokens * 0.03 / 1000.0 * float64(agg.Requests))

	return metrics
}

// calculatePercentile calculates the nth percentile of a slice
func calculatePercentile(values []int, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]int, len(values))
	copy(sorted, values)
	sort.Ints(sorted)

	index := int(float64(len(sorted)-1) * percentile)
	return float64(sorted[index])
}

// cleanupState periodically removes old state
func (p *MetricsProcessor) cleanupState(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			cutoff := now.Add(-StateRetentionDuration)

			p.stateMu.Lock()

			// Clean old requests
			for id, req := range p.requestState {
				if req.Timestamp.Before(cutoff) {
					delete(p.requestState, id)
				}
			}

			// Clean old responses
			for id, resp := range p.responseState {
				if resp.Timestamp.Before(cutoff) {
					delete(p.responseState, id)
				}
			}

			p.stateMu.Unlock()
		}
	}
}

// Close shuts down the processor
func (p *MetricsProcessor) Close() {
	p.windowTicker.Stop()
	log.Println("Metrics processor closed")
}
