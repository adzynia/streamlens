package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"streamlens/internal/models"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// MetricsStore provides database operations for metrics
type MetricsStore struct {
	db *sql.DB
}

// NewMetricsStore creates a new MetricsStore
func NewMetricsStore(dsn string) (*MetricsStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to Postgres")
	return &MetricsStore{db: db}, nil
}

// InsertMetrics inserts an aggregated metrics record into the database
func (s *MetricsStore) InsertMetrics(ctx context.Context, metrics *models.LLMMetrics) error {
	query := `
		INSERT INTO llm_metrics (
			tenant_id, route, model, window_start, window_end,
			requests, errors, avg_latency_ms, p95_latency_ms,
			avg_prompt_tokens, avg_completion_tokens, estimated_cost_usd
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (tenant_id, route, model, window_start)
		DO UPDATE SET
			window_end = EXCLUDED.window_end,
			requests = EXCLUDED.requests,
			errors = EXCLUDED.errors,
			avg_latency_ms = EXCLUDED.avg_latency_ms,
			p95_latency_ms = EXCLUDED.p95_latency_ms,
			avg_prompt_tokens = EXCLUDED.avg_prompt_tokens,
			avg_completion_tokens = EXCLUDED.avg_completion_tokens,
			estimated_cost_usd = EXCLUDED.estimated_cost_usd
	`

	_, err := s.db.ExecContext(ctx, query,
		metrics.TenantID,
		metrics.Route,
		metrics.Model,
		metrics.WindowStart,
		metrics.WindowEnd,
		metrics.Requests,
		metrics.Errors,
		metrics.AvgLatencyMs,
		metrics.P95LatencyMs,
		metrics.AvgPromptTokens,
		metrics.AvgCompletionTokens,
		metrics.EstimatedCostUSD,
	)

	return err
}

// QueryMetrics retrieves metrics based on filters
func (s *MetricsStore) QueryMetrics(ctx context.Context, tenantID string, route, model *string, limit int) ([]models.LLMMetrics, error) {
	// Build query dynamically based on filters
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, tenantID)
	argIndex++

	if route != nil && *route != "" {
		conditions = append(conditions, fmt.Sprintf("route = $%d", argIndex))
		args = append(args, *route)
		argIndex++
	}

	if model != nil && *model != "" {
		conditions = append(conditions, fmt.Sprintf("model = $%d", argIndex))
		args = append(args, *model)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Default limit if not specified
	if limit <= 0 {
		limit = 60
	}

	query := fmt.Sprintf(`
		SELECT tenant_id, route, model, window_start, window_end,
		       requests, errors, avg_latency_ms, p95_latency_ms,
		       avg_prompt_tokens, avg_completion_tokens, estimated_cost_usd
		FROM llm_metrics
		WHERE %s
		ORDER BY window_start DESC
		LIMIT $%d
	`, whereClause, argIndex)

	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.LLMMetrics
	for rows.Next() {
		var m models.LLMMetrics
		err := rows.Scan(
			&m.TenantID,
			&m.Route,
			&m.Model,
			&m.WindowStart,
			&m.WindowEnd,
			&m.Requests,
			&m.Errors,
			&m.AvgLatencyMs,
			&m.P95LatencyMs,
			&m.AvgPromptTokens,
			&m.AvgCompletionTokens,
			&m.EstimatedCostUSD,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Close closes the database connection
func (s *MetricsStore) Close() error {
	log.Println("Closing Postgres connection")
	return s.db.Close()
}
