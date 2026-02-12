package models

import "time"

// DailyUsage represents aggregated usage stats for a day
type DailyUsage struct {
	Date             string `json:"date"` // YYYY-MM-DD
	CredentialID     string `json:"credential_id,omitempty"`
	Model            string `json:"model"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	ErrorCount       int    `json:"error_count"`
}

// ModelStats represents usage statistics for a specific model
type ModelStats struct {
	Model            string `json:"model"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	ErrorCount       int    `json:"error_count"`
}

// UsageStats represents aggregated usage statistics
type UsageStats struct {
	TotalRequests         int                    `json:"total_requests"`
	TotalTokens           int                    `json:"total_tokens"`
	TotalPromptTokens     int                    `json:"prompt_tokens"`
	TotalCompletionTokens int                    `json:"completion_tokens"`
	ErrorCount            int                    `json:"error_count"`
	ModelBreakdown        map[string]*ModelStats `json:"models,omitempty"`
}

// StatsFilter contains parameters for filtering usage statistics
type StatsFilter struct {
	CredentialID string
	Model        string
	Provider     string
	StartDate    *time.Time
	EndDate      *time.Time
}
