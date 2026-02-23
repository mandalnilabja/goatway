// Package models contains data models for storage operations.
package models

import (
	"encoding/json"
	"time"
)

// Credential represents a stored API credential for an LLM provider.
// Data contains provider-specific credential fields as JSON.
type Credential struct {
	ID        string          `json:"id"`
	Provider  string          `json:"provider"` // openrouter, openai, anthropic, azure
	Name      string          `json:"name"`     // User-friendly name
	Data      json.RawMessage `json:"data"`     // Provider-specific credential data (encrypted at rest)
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// CredentialPreview is a safe representation of a credential (secrets masked).
type CredentialPreview struct {
	ID          string          `json:"id"`
	Provider    string          `json:"provider"`
	Name        string          `json:"name"`
	DataPreview json.RawMessage `json:"data_preview"` // Masked credential data
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Provider-specific credential types

// APIKeyCredential is for providers that only need an API key (OpenRouter, OpenAI, Anthropic).
type APIKeyCredential struct {
	APIKey string `json:"api_key"`
}

// AzureCredential contains Azure OpenAI-specific fields.
type AzureCredential struct {
	Endpoint   string `json:"endpoint"`
	APIKey     string `json:"api_key"`
	Deployment string `json:"deployment"`
	APIVersion string `json:"api_version"`
}

// ToPreview converts a Credential to a safe CredentialPreview with masked secrets.
func (c *Credential) ToPreview() *CredentialPreview {
	return &CredentialPreview{
		ID:          c.ID,
		Provider:    c.Provider,
		Name:        c.Name,
		DataPreview: maskCredentialData(c.Provider, c.Data),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// maskCredentialData masks sensitive fields in credential data based on provider type.
func maskCredentialData(provider string, data json.RawMessage) json.RawMessage {
	switch provider {
	case "azure":
		var cred AzureCredential
		if err := json.Unmarshal(data, &cred); err == nil {
			cred.APIKey = maskSecret(cred.APIKey)
			masked, _ := json.Marshal(cred)
			return masked
		}
	default:
		var cred APIKeyCredential
		if err := json.Unmarshal(data, &cred); err == nil {
			cred.APIKey = maskSecret(cred.APIKey)
			masked, _ := json.Marshal(cred)
			return masked
		}
	}
	return json.RawMessage(`{"api_key":"***"}`)
}

// maskSecret creates a masked preview of a secret string.
func maskSecret(secret string) string {
	if len(secret) <= 10 {
		return "***"
	}
	return secret[:6] + "..." + secret[len(secret)-4:]
}

// GetAPIKey extracts the API key from credential data (works for all provider types).
func (c *Credential) GetAPIKey() string {
	var cred APIKeyCredential
	if err := json.Unmarshal(c.Data, &cred); err == nil {
		return cred.APIKey
	}
	return ""
}

// GetAzureCredential extracts Azure-specific credential data.
func (c *Credential) GetAzureCredential() (*AzureCredential, error) {
	var cred AzureCredential
	if err := json.Unmarshal(c.Data, &cred); err != nil {
		return nil, err
	}
	return &cred, nil
}
