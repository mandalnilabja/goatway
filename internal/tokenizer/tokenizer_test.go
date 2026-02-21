package tokenizer

import (
	"testing"
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
