package types

// ChatCompletionChunk represents a streaming chunk response.
type ChatCompletionChunk struct {
	ID                string        `json:"id"`
	Object            string        `json:"object"` // "chat.completion.chunk"
	Created           int64         `json:"created"`
	Model             string        `json:"model"`
	Choices           []ChunkChoice `json:"choices"`
	Usage             *Usage        `json:"usage,omitempty"` // Only in final chunk if requested
	SystemFingerprint string        `json:"system_fingerprint,omitempty"`
	ServiceTier       string        `json:"service_tier,omitempty"`
}

// ChunkChoice represents a choice in a streaming chunk.
type ChunkChoice struct {
	Index        int             `json:"index"`
	Delta        Delta           `json:"delta"`
	FinishReason *string         `json:"finish_reason"` // Pointer to distinguish null from ""
	Logprobs     *ChoiceLogprobs `json:"logprobs,omitempty"`
}

// Delta represents the incremental content in a streaming chunk.
type Delta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// IsFinalChunk returns true if this chunk signals the end of generation.
func (c *ChunkChoice) IsFinalChunk() bool {
	return c.FinishReason != nil
}

// GetFinishReason returns the finish reason or empty string if not final.
func (c *ChunkChoice) GetFinishReason() string {
	if c.FinishReason == nil {
		return ""
	}
	return *c.FinishReason
}

// SSE formatting helpers

// SSEPrefix is the Server-Sent Events data prefix.
const SSEPrefix = "data: "

// SSEDone is the final SSE message indicating stream end.
const SSEDone = "data: [DONE]\n\n"

// FormatSSE formats a chunk for Server-Sent Events transmission.
func FormatSSE(data []byte) []byte {
	result := make([]byte, 0, len(SSEPrefix)+len(data)+2)
	result = append(result, SSEPrefix...)
	result = append(result, data...)
	result = append(result, '\n', '\n')
	return result
}
