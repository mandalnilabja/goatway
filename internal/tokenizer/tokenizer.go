// Package tokenizer provides token counting for OpenAI-compatible requests.
package tokenizer

import (
	"strings"
	"sync"

	"github.com/pkoukk/tiktoken-go"

	"github.com/mandalnilabja/goatway/internal/types"
)

// Tokenizer counts tokens for chat completion requests.
type Tokenizer interface {
	// CountTokens counts tokens in a text string for a given model.
	CountTokens(text string, model string) (int, error)

	// CountMessages counts tokens for a slice of messages.
	CountMessages(messages []types.Message, model string) (int, error)

	// CountRequest counts total prompt tokens for a full request.
	CountRequest(req *types.ChatCompletionRequest) (int, error)
}

// Encoding names used by tiktoken.
const (
	EncodingCL100kBase = "cl100k_base" // GPT-4, GPT-3.5-turbo
	EncodingO200kBase  = "o200k_base"  // GPT-4o, o1 models
)

// modelEncoding pairs a prefix with its encoding.
type modelEncoding struct {
	prefix   string
	encoding string
}

// modelEncodings lists model prefixes and their encodings.
// Ordered by prefix length (longest first) to ensure correct matching.
var modelEncodings = []modelEncoding{
	// Longer prefixes first to avoid partial matches
	{"text-embedding", EncodingCL100kBase},
	{"gpt-4o", EncodingO200kBase}, // Must come before "gpt-4"
	{"gpt-3.5", EncodingCL100kBase},
	{"gpt-4", EncodingCL100kBase},
	{"chatgpt", EncodingO200kBase},
	{"o1", EncodingO200kBase},
	{"o3", EncodingO200kBase},
}

// TiktokenTokenizer implements Tokenizer using tiktoken-go.
type TiktokenTokenizer struct {
	mu        sync.RWMutex
	encodings map[string]*tiktoken.Tiktoken
}

// New creates a new TiktokenTokenizer.
func New() *TiktokenTokenizer {
	return &TiktokenTokenizer{
		encodings: make(map[string]*tiktoken.Tiktoken),
	}
}

// getEncoding returns the tiktoken encoding for a model, with caching.
func (t *TiktokenTokenizer) getEncoding(model string) (*tiktoken.Tiktoken, error) {
	encodingName := t.resolveEncoding(model)

	// Check cache first
	t.mu.RLock()
	enc, ok := t.encodings[encodingName]
	t.mu.RUnlock()
	if ok {
		return enc, nil
	}

	// Create new encoding
	t.mu.Lock()
	defer t.mu.Unlock()

	// Double-check after acquiring write lock
	if enc, ok = t.encodings[encodingName]; ok {
		return enc, nil
	}

	enc, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		return nil, err
	}
	t.encodings[encodingName] = enc
	return enc, nil
}

// resolveEncoding determines the encoding name for a model.
func (t *TiktokenTokenizer) resolveEncoding(model string) string {
	modelLower := strings.ToLower(model)

	// Check for prefix matches (ordered by length, longest first)
	for _, me := range modelEncodings {
		if strings.HasPrefix(modelLower, me.prefix) {
			return me.encoding
		}
	}

	// Default to cl100k_base for unknown models (including Claude, etc.)
	return EncodingCL100kBase
}

// CountTokens counts tokens in a text string for a given model.
func (t *TiktokenTokenizer) CountTokens(text string, model string) (int, error) {
	enc, err := t.getEncoding(model)
	if err != nil {
		return 0, err
	}
	tokens := enc.Encode(text, nil, nil)
	return len(tokens), nil
}
