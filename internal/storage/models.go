package storage

import "time"

// Credential represents a stored API credential for an LLM provider
type Credential struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"` // openrouter, openai, anthropic, azure
	Name      string    `json:"name"`     // User-friendly name
	APIKey    string    `json:"api_key"`  // Encrypted at rest
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CredentialPreview is a safe representation of a credential (key masked)
type CredentialPreview struct {
	ID            string    `json:"id"`
	Provider      string    `json:"provider"`
	Name          string    `json:"name"`
	APIKeyPreview string    `json:"api_key_preview"` // e.g., "sk-or-v1-xxx...xxx"
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MaskAPIKey creates a masked preview of an API key
func MaskAPIKey(key string) string {
	if len(key) <= 10 {
		return "***"
	}
	return key[:6] + "..." + key[len(key)-4:]
}

// ToPreview converts a Credential to a safe CredentialPreview
func (c *Credential) ToPreview() *CredentialPreview {
	return &CredentialPreview{
		ID:            c.ID,
		Provider:      c.Provider,
		Name:          c.Name,
		APIKeyPreview: MaskAPIKey(c.APIKey),
		IsDefault:     c.IsDefault,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

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

// StatsFilter contains parameters for filtering usage statistics
type StatsFilter struct {
	CredentialID string
	Model        string
	Provider     string
	StartDate    *time.Time
	EndDate      *time.Time
}
