-- Create metrics table
CREATE TABLE IF NOT EXISTS llm_metrics (
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

-- Create indexes for efficient querying
CREATE INDEX idx_llm_metrics_tenant_time ON llm_metrics(tenant_id, window_start DESC);
CREATE INDEX idx_llm_metrics_composite ON llm_metrics(tenant_id, route, model, window_start DESC);
