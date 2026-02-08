package types

// ChatCompletionResponse represents a non-streaming chat completion response.
type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"` // "chat.completion"
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             *Usage   `json:"usage,omitempty"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
	ServiceTier       string   `json:"service_tier,omitempty"`
}

// Object constants
const (
	ObjectChatCompletion      = "chat.completion"
	ObjectChatCompletionChunk = "chat.completion.chunk"
)

// Choice represents a single completion choice.
type Choice struct {
	Index        int             `json:"index"`
	Message      Message         `json:"message"`
	FinishReason string          `json:"finish_reason,omitempty"`
	Logprobs     *ChoiceLogprobs `json:"logprobs,omitempty"`
}

// FinishReason constants
const (
	FinishReasonStop          = "stop"
	FinishReasonLength        = "length"
	FinishReasonToolCalls     = "tool_calls"
	FinishReasonContentFilter = "content_filter"
)

// Usage represents token usage statistics.
type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	CompletionTokensDetails *CompletionTokenDetails `json:"completion_tokens_details,omitempty"`
	PromptTokensDetails     *PromptTokenDetails     `json:"prompt_tokens_details,omitempty"`
}

// CompletionTokenDetails provides breakdown of completion tokens.
type CompletionTokenDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens,omitempty"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens,omitempty"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens,omitempty"`
}

// PromptTokenDetails provides breakdown of prompt tokens.
type PromptTokenDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
}

// ChoiceLogprobs contains log probability information for a choice.
type ChoiceLogprobs struct {
	Content []TokenLogprob `json:"content,omitempty"`
}

// TokenLogprob represents log probability for a single token.
type TokenLogprob struct {
	Token       string            `json:"token"`
	Logprob     float64           `json:"logprob"`
	Bytes       []int             `json:"bytes,omitempty"`
	TopLogprobs []TopLogprobEntry `json:"top_logprobs,omitempty"`
}

// TopLogprobEntry represents a top token and its probability.
type TopLogprobEntry struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}

// TotalTokens returns the sum of prompt and completion tokens.
func (u *Usage) TotalUsage() int {
	if u == nil {
		return 0
	}
	return u.PromptTokens + u.CompletionTokens
}
