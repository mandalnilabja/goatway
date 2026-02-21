package tokenizer

import (
	"testing"

	"github.com/mandalnilabja/goatway/internal/types"
)

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
