// Package models contains data models for storage operations.
package models

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
