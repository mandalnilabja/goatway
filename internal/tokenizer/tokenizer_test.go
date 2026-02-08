package tokenizer

import (
	"testing"

	"github.com/mandalnilabja/goatway/internal/types"
)

func TestNew(t *testing.T) {
	tok := New()
	if tok == nil {
		t.Fatal("New() returned nil")
	}
	if tok.encodings == nil {
		t.Fatal("encodings map is nil")
	}
}

func TestCountTokens(t *testing.T) {
	tok := New()

	tests := []struct {
		name     string
		text     string
		model    string
		minCount int // Token counts may vary slightly
		maxCount int
	}{
		{
			name:     "simple text gpt-4",
			text:     "Hello, world!",
			model:    "gpt-4",
			minCount: 3,
			maxCount: 5,
		},
		{
			name:     "simple text gpt-3.5",
			text:     "Hello, world!",
			model:    "gpt-3.5-turbo",
			minCount: 3,
			maxCount: 5,
		},
		{
			name:     "simple text gpt-4o",
			text:     "Hello, world!",
			model:    "gpt-4o",
			minCount: 3,
			maxCount: 5,
		},
		{
			name:     "unknown model defaults to cl100k",
			text:     "Hello, world!",
			model:    "claude-3-opus",
			minCount: 3,
			maxCount: 5,
		},
		{
			name:     "empty text",
			text:     "",
			model:    "gpt-4",
			minCount: 0,
			maxCount: 0,
		},
		{
			name:     "longer text",
			text:     "The quick brown fox jumps over the lazy dog.",
			model:    "gpt-4",
			minCount: 8,
			maxCount: 12,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.CountTokens(tc.text, tc.model)
			if err != nil {
				t.Fatalf("CountTokens() error: %v", err)
			}
			if count < tc.minCount || count > tc.maxCount {
				t.Errorf("CountTokens() = %d, want between %d and %d",
					count, tc.minCount, tc.maxCount)
			}
		})
	}
}

func TestResolveEncoding(t *testing.T) {
	tok := New()

	tests := []struct {
		model    string
		expected string
	}{
		{"gpt-4", EncodingCL100kBase},
		{"gpt-4-turbo", EncodingCL100kBase},
		{"gpt-4-0125-preview", EncodingCL100kBase},
		{"gpt-3.5-turbo", EncodingCL100kBase},
		{"gpt-3.5-turbo-16k", EncodingCL100kBase},
		{"gpt-4o", EncodingO200kBase},
		{"gpt-4o-mini", EncodingO200kBase},
		{"o1-preview", EncodingO200kBase},
		{"o1-mini", EncodingO200kBase},
		{"chatgpt-4o-latest", EncodingO200kBase},
		// Unknown models default to cl100k_base
		{"claude-3-opus", EncodingCL100kBase},
		{"claude-3-sonnet", EncodingCL100kBase},
		{"mistral-7b", EncodingCL100kBase},
		{"unknown-model", EncodingCL100kBase},
	}

	for _, tc := range tests {
		t.Run(tc.model, func(t *testing.T) {
			result := tok.resolveEncoding(tc.model)
			if result != tc.expected {
				t.Errorf("resolveEncoding(%q) = %q, want %q",
					tc.model, result, tc.expected)
			}
		})
	}
}

func TestCountMessages(t *testing.T) {
	tok := New()

	tests := []struct {
		name     string
		messages []types.Message
		model    string
		minCount int
		maxCount int
	}{
		{
			name: "single user message",
			messages: []types.Message{
				types.NewTextMessage(types.RoleUser, "Hello!"),
			},
			model:    "gpt-4",
			minCount: 5,
			maxCount: 10,
		},
		{
			name: "system and user messages",
			messages: []types.Message{
				types.NewTextMessage(types.RoleSystem, "You are a helpful assistant."),
				types.NewTextMessage(types.RoleUser, "Hello!"),
			},
			model:    "gpt-4",
			minCount: 12,
			maxCount: 20,
		},
		{
			name: "conversation with assistant",
			messages: []types.Message{
				types.NewTextMessage(types.RoleSystem, "You are helpful."),
				types.NewTextMessage(types.RoleUser, "Hi"),
				types.NewTextMessage(types.RoleAssistant, "Hello! How can I help?"),
				types.NewTextMessage(types.RoleUser, "What is 2+2?"),
			},
			model:    "gpt-4",
			minCount: 25,
			maxCount: 40,
		},
		{
			name:     "empty messages",
			messages: []types.Message{},
			model:    "gpt-4",
			minCount: 3, // Reply priming only
			maxCount: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.CountMessages(tc.messages, tc.model)
			if err != nil {
				t.Fatalf("CountMessages() error: %v", err)
			}
			if count < tc.minCount || count > tc.maxCount {
				t.Errorf("CountMessages() = %d, want between %d and %d",
					count, tc.minCount, tc.maxCount)
			}
		})
	}
}

func TestCountRequest(t *testing.T) {
	tok := New()

	tests := []struct {
		name     string
		req      *types.ChatCompletionRequest
		minCount int
		maxCount int
	}{
		{
			name: "simple request",
			req: &types.ChatCompletionRequest{
				Model: "gpt-4",
				Messages: []types.Message{
					types.NewTextMessage(types.RoleUser, "Hello!"),
				},
			},
			minCount: 5,
			maxCount: 10,
		},
		{
			name: "request with tools",
			req: &types.ChatCompletionRequest{
				Model: "gpt-4",
				Messages: []types.Message{
					types.NewTextMessage(types.RoleUser, "What's the weather?"),
				},
				Tools: []types.Tool{
					types.NewTool("get_weather", "Get weather for a location", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "City name",
							},
						},
						"required": []string{"location"},
					}),
				},
			},
			minCount: 30,
			maxCount: 60,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.CountRequest(tc.req)
			if err != nil {
				t.Fatalf("CountRequest() error: %v", err)
			}
			if count < tc.minCount || count > tc.maxCount {
				t.Errorf("CountRequest() = %d, want between %d and %d",
					count, tc.minCount, tc.maxCount)
			}
		})
	}
}

func TestCountImageTokens(t *testing.T) {
	tok := New()

	tests := []struct {
		name     string
		image    *types.ImageURL
		expected int
	}{
		{
			name:     "nil image",
			image:    nil,
			expected: 0,
		},
		{
			name:     "low detail",
			image:    &types.ImageURL{URL: "http://example.com/img.jpg", Detail: "low"},
			expected: imageBaseTokens + imageLowDetailTiles*imageTileTokens,
		},
		{
			name:     "high detail",
			image:    &types.ImageURL{URL: "http://example.com/img.jpg", Detail: "high"},
			expected: imageBaseTokens + imageHighDetailMax*imageTileTokens,
		},
		{
			name:     "auto detail",
			image:    &types.ImageURL{URL: "http://example.com/img.jpg", Detail: "auto"},
			expected: imageBaseTokens + imageHighDetailMax*imageTileTokens,
		},
		{
			name:     "no detail specified",
			image:    &types.ImageURL{URL: "http://example.com/img.jpg"},
			expected: imageBaseTokens + imageHighDetailMax*imageTileTokens,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tok.countImageTokens(tc.image)
			if result != tc.expected {
				t.Errorf("countImageTokens() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestEncodingCaching(t *testing.T) {
	tok := New()

	// Count tokens twice with same model - should use cached encoding
	_, err := tok.CountTokens("hello", "gpt-4")
	if err != nil {
		t.Fatalf("first CountTokens() error: %v", err)
	}

	_, err = tok.CountTokens("world", "gpt-4")
	if err != nil {
		t.Fatalf("second CountTokens() error: %v", err)
	}

	// Check that encoding was cached
	tok.mu.RLock()
	defer tok.mu.RUnlock()
	if len(tok.encodings) != 1 {
		t.Errorf("expected 1 cached encoding, got %d", len(tok.encodings))
	}
}
