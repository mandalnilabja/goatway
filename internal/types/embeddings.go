package types

// EmbeddingsRequest represents an OpenAI embeddings API request.
type EmbeddingsRequest struct {
	// Required: ID of the model to use
	Model string `json:"model"`

	// Required: Input text to embed, can be string or array of strings
	Input EmbeddingsInput `json:"input"`

	// Optional: Encoding format for the embeddings
	// Values: "float" (default) or "base64"
	EncodingFormat string `json:"encoding_format,omitempty"`

	// Optional: Number of dimensions for the output embeddings (only for some models)
	Dimensions *int `json:"dimensions,omitempty"`

	// Optional: Unique identifier for the end-user
	User string `json:"user,omitempty"`
}

// EmbeddingsInput handles both string and array inputs for embeddings.
type EmbeddingsInput struct {
	Values []string
}

// MarshalJSON implements custom marshaling for EmbeddingsInput.
func (e EmbeddingsInput) MarshalJSON() ([]byte, error) {
	if len(e.Values) == 0 {
		return []byte(`""`), nil
	}
	if len(e.Values) == 1 {
		return []byte(`"` + e.Values[0] + `"`), nil
	}
	return marshalStringArray(e.Values)
}

// UnmarshalJSON implements custom unmarshaling for EmbeddingsInput.
func (e *EmbeddingsInput) UnmarshalJSON(data []byte) error {
	e.Values = nil
	// Try string first
	var single string
	if err := unmarshalString(data, &single); err == nil {
		e.Values = []string{single}
		return nil
	}
	// Try array of strings
	return unmarshalStringArray(data, &e.Values)
}

// EmbeddingsResponse represents an OpenAI embeddings API response.
type EmbeddingsResponse struct {
	Object string           `json:"object"` // "list"
	Data   []EmbeddingData  `json:"data"`
	Model  string           `json:"model"`
	Usage  *EmbeddingsUsage `json:"usage,omitempty"`
}

// EmbeddingData represents a single embedding in the response.
type EmbeddingData struct {
	Object    string    `json:"object"` // "embedding"
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingsUsage represents token usage for embeddings.
type EmbeddingsUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
