package admin

import "encoding/json"

// CreateCredentialRequest is the request body for creating a credential.
type CreateCredentialRequest struct {
	Provider  string          `json:"provider"`
	Name      string          `json:"name"`
	Data      json.RawMessage `json:"data"` // Provider-specific credential data
	IsDefault bool            `json:"is_default"`
}

// UpdateCredentialRequest is the request body for updating a credential.
type UpdateCredentialRequest struct {
	Provider  *string          `json:"provider,omitempty"`
	Name      *string          `json:"name,omitempty"`
	Data      *json.RawMessage `json:"data,omitempty"` // Provider-specific credential data
	IsDefault *bool            `json:"is_default,omitempty"`
}
