package admin

// CreateCredentialRequest is the request body for creating a credential.
type CreateCredentialRequest struct {
	Provider  string `json:"provider"`
	Name      string `json:"name"`
	APIKey    string `json:"api_key"`
	IsDefault bool   `json:"is_default"`
}

// UpdateCredentialRequest is the request body for updating a credential.
type UpdateCredentialRequest struct {
	Provider  *string `json:"provider,omitempty"`
	Name      *string `json:"name,omitempty"`
	APIKey    *string `json:"api_key,omitempty"`
	IsDefault *bool   `json:"is_default,omitempty"`
}
