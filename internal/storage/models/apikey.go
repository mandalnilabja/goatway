package models

import "time"

// ClientAPIKey represents a Goatway client API key for authentication
type ClientAPIKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyHash    string     `json:"-"`          // Argon2id hash (never exposed in JSON)
	KeyPrefix  string     `json:"key_prefix"` // First 11 chars (e.g., "gw_a1B2c3D4")
	Scopes     []string   `json:"scopes"`     // ["proxy", "admin"]
	RateLimit  int        `json:"rate_limit"` // Requests per minute (0 = unlimited)
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// ClientAPIKeyPreview is a safe representation (no hash)
type ClientAPIKeyPreview struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	Scopes     []string   `json:"scopes"`
	RateLimit  int        `json:"rate_limit"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// ToPreview converts ClientAPIKey to safe preview
func (k *ClientAPIKey) ToPreview() *ClientAPIKeyPreview {
	return &ClientAPIKeyPreview{
		ID:         k.ID,
		Name:       k.Name,
		KeyPrefix:  k.KeyPrefix,
		Scopes:     k.Scopes,
		RateLimit:  k.RateLimit,
		IsActive:   k.IsActive,
		LastUsedAt: k.LastUsedAt,
		CreatedAt:  k.CreatedAt,
		ExpiresAt:  k.ExpiresAt,
	}
}

// HasScope checks if the key has a specific scope
func (k *ClientAPIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// IsExpired checks if the key has expired
func (k *ClientAPIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}
