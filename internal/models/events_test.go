package models

import (
	"testing"
	"time"
)

func TestLLMRequest_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		request LLMRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: LLMRequest{
				RequestID:    "req-123",
				TenantID:     "tenant-1",
				Route:        "chat_support",
				Model:        "gpt-4",
				Timestamp:    now,
				PromptTokens: 100,
			},
			wantErr: nil,
		},
		{
			name: "missing request_id",
			request: LLMRequest{
				TenantID:     "tenant-1",
				Route:        "chat_support",
				Model:        "gpt-4",
				Timestamp:    now,
				PromptTokens: 100,
			},
			wantErr: ErrMissingRequestID,
		},
		{
			name: "missing tenant_id",
			request: LLMRequest{
				RequestID:    "req-123",
				Route:        "chat_support",
				Model:        "gpt-4",
				Timestamp:    now,
				PromptTokens: 100,
			},
			wantErr: ErrMissingTenantID,
		},
		{
			name: "missing route",
			request: LLMRequest{
				RequestID:    "req-123",
				TenantID:     "tenant-1",
				Model:        "gpt-4",
				Timestamp:    now,
				PromptTokens: 100,
			},
			wantErr: ErrMissingRoute,
		},
		{
			name: "missing model",
			request: LLMRequest{
				RequestID:    "req-123",
				TenantID:     "tenant-1",
				Route:        "chat_support",
				Timestamp:    now,
				PromptTokens: 100,
			},
			wantErr: ErrMissingModel,
		},
		{
			name: "missing timestamp",
			request: LLMRequest{
				RequestID:    "req-123",
				TenantID:     "tenant-1",
				Route:        "chat_support",
				Model:        "gpt-4",
				PromptTokens: 100,
			},
			wantErr: ErrMissingTimestamp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLLMResponse_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		response LLMResponse
		wantErr  error
	}{
		{
			name: "valid response",
			response: LLMResponse{
				RequestID:        "req-123",
				Timestamp:        now,
				LatencyMs:        500,
				CompletionTokens: 200,
				FinishReason:     "stop",
			},
			wantErr: nil,
		},
		{
			name: "valid response with error",
			response: LLMResponse{
				RequestID:        "req-123",
				Timestamp:        now,
				LatencyMs:        500,
				CompletionTokens: 0,
				FinishReason:     "error",
				Error:            stringPtr("rate_limit_exceeded"),
			},
			wantErr: nil,
		},
		{
			name: "missing request_id",
			response: LLMResponse{
				Timestamp:        now,
				LatencyMs:        500,
				CompletionTokens: 200,
				FinishReason:     "stop",
			},
			wantErr: ErrMissingRequestID,
		},
		{
			name: "missing timestamp",
			response: LLMResponse{
				RequestID:        "req-123",
				LatencyMs:        500,
				CompletionTokens: 200,
				FinishReason:     "stop",
			},
			wantErr: ErrMissingTimestamp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
