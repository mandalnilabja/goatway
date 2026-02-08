package types

// ChatCompletionRequest represents an OpenAI chat completion request.
// All optional fields use pointers to distinguish between unset and zero values.
type ChatCompletionRequest struct {
	// Required fields
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`

	// Sampling parameters
	Temperature         *float64 `json:"temperature,omitempty"` // 0-2, default 1
	TopP                *float64 `json:"top_p,omitempty"`       // 0-1, default 1
	N                   *int     `json:"n,omitempty"`           // Number of completions
	MaxTokens           *int     `json:"max_tokens,omitempty"`  // Deprecated: use max_completion_tokens
	MaxCompletionTokens *int     `json:"max_completion_tokens,omitempty"`
	PresencePenalty     *float64 `json:"presence_penalty,omitempty"`  // -2 to 2, default 0
	FrequencyPenalty    *float64 `json:"frequency_penalty,omitempty"` // -2 to 2, default 0

	// Stopping conditions
	Stop Stop `json:"stop,omitempty"` // String or array of up to 4 strings

	// Streaming
	Stream        bool           `json:"stream,omitempty"`
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`

	// Tool/function calling
	Tools      []Tool      `json:"tools,omitempty"`
	ToolChoice *ToolChoice `json:"tool_choice,omitempty"`

	// Response format
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// Advanced options
	Seed        *int   `json:"seed,omitempty"`         // For deterministic outputs
	User        string `json:"user,omitempty"`         // End-user identifier
	Logprobs    *bool  `json:"logprobs,omitempty"`     // Return log probabilities
	TopLogprobs *int   `json:"top_logprobs,omitempty"` // 0-20
}

// StreamOptions controls streaming behavior.
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"` // Include usage in final chunk
}

// ResponseFormat specifies the output format.
type ResponseFormat struct {
	Type       string      `json:"type"`                  // "text", "json_object", "json_schema"
	JSONSchema *JSONSchema `json:"json_schema,omitempty"` // For structured outputs
}

// JSONSchema defines a JSON schema for structured outputs.
type JSONSchema struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Schema      interface{} `json:"schema"` // The actual JSON Schema
	Strict      *bool       `json:"strict,omitempty"`
}

// Stop represents stop sequences that can be a string or array.
type Stop struct {
	Values []string
}

// MarshalJSON implements custom marshaling for Stop.
func (s Stop) MarshalJSON() ([]byte, error) {
	if len(s.Values) == 0 {
		return []byte("null"), nil
	}
	if len(s.Values) == 1 {
		return []byte(`"` + s.Values[0] + `"`), nil
	}
	return marshalStringArray(s.Values)
}

// UnmarshalJSON implements custom unmarshaling for Stop.
func (s *Stop) UnmarshalJSON(data []byte) error {
	s.Values = nil
	// Try string first
	var single string
	if err := unmarshalString(data, &single); err == nil {
		s.Values = []string{single}
		return nil
	}
	// Try array
	return unmarshalStringArray(data, &s.Values)
}

// IsStreaming returns true if this is a streaming request.
func (r *ChatCompletionRequest) IsStreaming() bool {
	return r.Stream
}

// GetMaxTokens returns the effective max tokens limit.
func (r *ChatCompletionRequest) GetMaxTokens() int {
	if r.MaxCompletionTokens != nil {
		return *r.MaxCompletionTokens
	}
	if r.MaxTokens != nil {
		return *r.MaxTokens
	}
	return 0 // No limit specified
}
