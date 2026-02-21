package admin

import "time"

// CreateAPIKeyRequest is the request body for creating an API key.
type CreateAPIKeyRequest struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`     // ["proxy", "admin"]
	RateLimit int      `json:"rate_limit"` // Requests per minute (0 = unlimited)
	ExpiresIn *int     `json:"expires_in"` // Seconds until expiry (optional)
}

// CreateAPIKeyResponse includes the plaintext key (shown only once).
type CreateAPIKeyResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"` // Plaintext - shown only once!
	KeyPrefix string     `json:"key_prefix"`
	Scopes    []string   `json:"scopes"`
	RateLimit int        `json:"rate_limit"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// UpdateAPIKeyRequest is the request body for updating an API key.
type UpdateAPIKeyRequest struct {
	Name      *string  `json:"name"`
	Scopes    []string `json:"scopes"`
	RateLimit *int     `json:"rate_limit"`
	IsActive  *bool    `json:"is_active"`
}
