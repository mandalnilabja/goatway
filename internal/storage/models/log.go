package models

import "time"

// RequestLog represents a logged API request
type RequestLog struct {
	ID               string    `json:"id"`
	RequestID        string    `json:"request_id"`
	CredentialID     string    `json:"credential_id,omitempty"`
	Model            string    `json:"model"`
	Provider         string    `json:"provider"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	IsStreaming      bool      `json:"is_streaming"`
	StatusCode       int       `json:"status_code"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	DurationMs       int64     `json:"duration_ms"`
	CreatedAt        time.Time `json:"created_at"`
}

// LogFilter contains parameters for filtering request logs
type LogFilter struct {
	CredentialID string
	Model        string
	Provider     string
	StatusCode   *int
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}
