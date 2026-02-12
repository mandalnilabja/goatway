package types

// CompletionRequest represents a legacy OpenAI completions API request.
// This endpoint is deprecated but still needed for client compatibility.
type CompletionRequest struct {
	// Required: ID of the model to use
	Model string `json:"model"`

	// Required: Prompt(s) to generate completions for
	// Can be string, array of strings, array of tokens, or array of token arrays
	Prompt CompletionPrompt `json:"prompt"`

	// Optional: Suffix after the inserted completion
	Suffix string `json:"suffix,omitempty"`

	// Optional: Maximum number of tokens to generate (default varies by model)
	MaxTokens *int `json:"max_tokens,omitempty"`

	// Optional: Sampling temperature (0-2, default 1)
	Temperature *float64 `json:"temperature,omitempty"`

	// Optional: Nucleus sampling parameter (0-1, default 1)
	TopP *float64 `json:"top_p,omitempty"`

	// Optional: Number of completions to generate (default 1)
	N *int `json:"n,omitempty"`

	// Optional: Stream back partial progress
	Stream bool `json:"stream,omitempty"`

	// Optional: Include log probabilities
	Logprobs *int `json:"logprobs,omitempty"`

	// Optional: Echo back the prompt with completion
	Echo bool `json:"echo,omitempty"`

	// Optional: Stop sequences (up to 4)
	Stop Stop `json:"stop,omitempty"`

	// Optional: Presence penalty (-2 to 2, default 0)
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// Optional: Frequency penalty (-2 to 2, default 0)
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// Optional: Generate best_of completions and return the best
	BestOf *int `json:"best_of,omitempty"`

	// Optional: Modify likelihood of specified tokens
	LogitBias map[string]float64 `json:"logit_bias,omitempty"`

	// Optional: Unique identifier for the end-user
	User string `json:"user,omitempty"`

	// Optional: Seed for deterministic sampling
	Seed *int `json:"seed,omitempty"`
}

// CompletionPrompt handles various prompt input formats.
type CompletionPrompt struct {
	Values []string
}

// MarshalJSON implements custom marshaling for CompletionPrompt.
func (p CompletionPrompt) MarshalJSON() ([]byte, error) {
	if len(p.Values) == 0 {
		return []byte(`""`), nil
	}
	if len(p.Values) == 1 {
		return []byte(`"` + p.Values[0] + `"`), nil
	}
	return marshalStringArray(p.Values)
}

// UnmarshalJSON implements custom unmarshaling for CompletionPrompt.
func (p *CompletionPrompt) UnmarshalJSON(data []byte) error {
	p.Values = nil
	// Try string first
	var single string
	if err := unmarshalString(data, &single); err == nil {
		p.Values = []string{single}
		return nil
	}
	// Try array of strings
	return unmarshalStringArray(data, &p.Values)
}

// CompletionResponse represents a legacy completions API response.
type CompletionResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"` // "text_completion"
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
	Choices           []CompletionChoice `json:"choices"`
	Usage             *CompletionUsage   `json:"usage,omitempty"`
}

// CompletionChoice represents a single completion choice.
type CompletionChoice struct {
	Text         string            `json:"text"`
	Index        int               `json:"index"`
	Logprobs     *CompletionLogprobs `json:"logprobs,omitempty"`
	FinishReason string            `json:"finish_reason"`
}

// CompletionLogprobs contains log probability information.
type CompletionLogprobs struct {
	Tokens        []string             `json:"tokens"`
	TokenLogprobs []float64            `json:"token_logprobs"`
	TopLogprobs   []map[string]float64 `json:"top_logprobs"`
	TextOffset    []int                `json:"text_offset"`
}

// CompletionUsage represents token usage for completions.
type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CompletionStreamChunk represents a streaming completion chunk.
type CompletionStreamChunk struct {
	ID                string                   `json:"id"`
	Object            string                   `json:"object"` // "text_completion"
	Created           int64                    `json:"created"`
	Model             string                   `json:"model"`
	SystemFingerprint string                   `json:"system_fingerprint,omitempty"`
	Choices           []CompletionStreamChoice `json:"choices"`
}

// CompletionStreamChoice represents a streaming choice.
type CompletionStreamChoice struct {
	Text         string          `json:"text"`
	Index        int             `json:"index"`
	Logprobs     *CompletionLogprobs `json:"logprobs,omitempty"`
	FinishReason *string         `json:"finish_reason,omitempty"`
}
