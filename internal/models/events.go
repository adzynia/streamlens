package models

import "time"

// LLMRequest represents a request event from an LLM-powered application
type LLMRequest struct {
	RequestID    string                 `json:"request_id"`
	TenantID     string                 `json:"tenant_id"`
	Route        string                 `json:"route"`
	Model        string                 `json:"model"`
	Timestamp    time.Time              `json:"timestamp"`
	PromptTokens int                    `json:"prompt_tokens"`
	UserIDHash   *string                `json:"user_id_hash,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LLMResponse represents a response event from an LLM service
type LLMResponse struct {
	RequestID        string     `json:"request_id"`
	Timestamp        time.Time  `json:"timestamp"`
	LatencyMs        int        `json:"latency_ms"`
	CompletionTokens int        `json:"completion_tokens"`
	FinishReason     string     `json:"finish_reason"`
	Error            *string    `json:"error,omitempty"`
}

// LLMMetrics represents aggregated metrics for a time window
type LLMMetrics struct {
	TenantID              string    `json:"tenant_id"`
	Route                 string    `json:"route"`
	Model                 string    `json:"model"`
	WindowStart           time.Time `json:"window_start"`
	WindowEnd             time.Time `json:"window_end"`
	Requests              int       `json:"requests"`
	Errors                int       `json:"errors"`
	AvgLatencyMs          float64   `json:"avg_latency_ms"`
	P95LatencyMs          float64   `json:"p95_latency_ms"`
	AvgPromptTokens       float64   `json:"avg_prompt_tokens"`
	AvgCompletionTokens   float64   `json:"avg_completion_tokens"`
	EstimatedCostUSD      float64   `json:"estimated_cost_usd"`
}

// Validate checks if LLMRequest has all required fields
func (r *LLMRequest) Validate() error {
	if r.RequestID == "" {
		return ErrMissingRequestID
	}
	if r.TenantID == "" {
		return ErrMissingTenantID
	}
	if r.Route == "" {
		return ErrMissingRoute
	}
	if r.Model == "" {
		return ErrMissingModel
	}
	if r.Timestamp.IsZero() {
		return ErrMissingTimestamp
	}
	return nil
}

// Validate checks if LLMResponse has all required fields
func (r *LLMResponse) Validate() error {
	if r.RequestID == "" {
		return ErrMissingRequestID
	}
	if r.Timestamp.IsZero() {
		return ErrMissingTimestamp
	}
	return nil
}
