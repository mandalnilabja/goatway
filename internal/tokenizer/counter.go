package tokenizer

import (
	"encoding/json"
	"strings"

	"github.com/mandalnilabja/goatway/internal/types"
)

// Message token overhead varies by model family.
// These values are based on OpenAI's documentation.
const (
	// Per-message overhead tokens
	messageOverheadGPT4  = 3 // <|start|>role<|end|>
	messageOverheadGPT35 = 4 // Slightly different format

	// Reply priming tokens (assistant response start)
	replyPrimingTokens = 3

	// Name field overhead (if present)
	nameOverhead = 1

	// Image token constants (OpenAI rules)
	imageBaseTokens     = 85  // Base cost for any image
	imageTileTokens     = 170 // Cost per 512x512 tile
	imageLowDetailTiles = 1   // Low detail uses 1 tile
	imageHighDetailMax  = 4   // High detail max tiles (simplified)
)

// CountMessages counts tokens for a slice of messages.
func (t *TiktokenTokenizer) CountMessages(messages []types.Message, model string) (int, error) {
	total := 0
	overhead := t.getMessageOverhead(model)

	for _, msg := range messages {
		tokens, err := t.countMessage(msg, model)
		if err != nil {
			return 0, err
		}
		total += tokens + overhead
	}

	// Add reply priming tokens
	total += replyPrimingTokens

	return total, nil
}

// CountRequest counts total prompt tokens for a full request.
func (t *TiktokenTokenizer) CountRequest(req *types.ChatCompletionRequest) (int, error) {
	total := 0

	// Count message tokens
	msgTokens, err := t.CountMessages(req.Messages, req.Model)
	if err != nil {
		return 0, err
	}
	total += msgTokens

	// Count tool definitions
	if len(req.Tools) > 0 {
		toolTokens, err := t.countTools(req.Tools, req.Model)
		if err != nil {
			return 0, err
		}
		total += toolTokens
	}

	return total, nil
}

// countMessage counts tokens for a single message.
func (t *TiktokenTokenizer) countMessage(msg types.Message, model string) (int, error) {
	total := 0

	// Count role tokens
	roleTokens, err := t.CountTokens(msg.Role, model)
	if err != nil {
		return 0, err
	}
	total += roleTokens

	// Count content tokens
	contentTokens, err := t.countContent(msg.Content, model)
	if err != nil {
		return 0, err
	}
	total += contentTokens

	// Count name tokens if present
	if msg.Name != "" {
		nameTokens, err := t.CountTokens(msg.Name, model)
		if err != nil {
			return 0, err
		}
		total += nameTokens + nameOverhead
	}

	// Count tool calls if present (assistant messages)
	if len(msg.ToolCalls) > 0 {
		toolCallTokens, err := t.countToolCalls(msg.ToolCalls, model)
		if err != nil {
			return 0, err
		}
		total += toolCallTokens
	}

	// Count tool call ID if present (tool response messages)
	if msg.ToolCallID != "" {
		idTokens, err := t.CountTokens(msg.ToolCallID, model)
		if err != nil {
			return 0, err
		}
		total += idTokens
	}

	return total, nil
}

// countContent counts tokens for message content (text or multimodal).
func (t *TiktokenTokenizer) countContent(content types.Content, model string) (int, error) {
	// Simple text content
	if content.Text != "" {
		return t.CountTokens(content.Text, model)
	}

	// Multimodal content
	total := 0
	for _, part := range content.Parts {
		switch part.Type {
		case types.ContentTypeText:
			tokens, err := t.CountTokens(part.Text, model)
			if err != nil {
				return 0, err
			}
			total += tokens

		case types.ContentTypeImageURL:
			total += t.countImageTokens(part.ImageURL)
		}
	}

	return total, nil
}

// countImageTokens calculates token cost for an image based on OpenAI's rules.
func (t *TiktokenTokenizer) countImageTokens(img *types.ImageURL) int {
	if img == nil {
		return 0
	}

	detail := strings.ToLower(img.Detail)
	switch detail {
	case "low":
		// Low detail: fixed cost
		return imageBaseTokens + (imageLowDetailTiles * imageTileTokens)
	case "high":
		// High detail: base + tiles (simplified without actual image dimensions)
		// In practice, you'd need image dimensions to calculate exactly.
		// Using a reasonable estimate of 4 tiles for high detail.
		return imageBaseTokens + (imageHighDetailMax * imageTileTokens)
	default:
		// "auto" or unspecified: assume high detail
		return imageBaseTokens + (imageHighDetailMax * imageTileTokens)
	}
}

// countTools counts tokens for tool definitions.
func (t *TiktokenTokenizer) countTools(tools []types.Tool, model string) (int, error) {
	total := 0

	for _, tool := range tools {
		// Count function name
		nameTokens, err := t.CountTokens(tool.Function.Name, model)
		if err != nil {
			return 0, err
		}
		total += nameTokens

		// Count function description
		if tool.Function.Description != "" {
			descTokens, err := t.CountTokens(tool.Function.Description, model)
			if err != nil {
				return 0, err
			}
			total += descTokens
		}

		// Count parameters schema as JSON
		if tool.Function.Parameters != nil {
			paramsJSON, err := json.Marshal(tool.Function.Parameters)
			if err != nil {
				return 0, err
			}
			paramsTokens, err := t.CountTokens(string(paramsJSON), model)
			if err != nil {
				return 0, err
			}
			total += paramsTokens
		}

		// Add overhead for tool structure
		total += 7 // Approximate overhead for {"type":"function","function":{...}}
	}

	return total, nil
}

// countToolCalls counts tokens for tool calls in assistant messages.
func (t *TiktokenTokenizer) countToolCalls(calls []types.ToolCall, model string) (int, error) {
	total := 0

	for _, call := range calls {
		// Count tool call ID
		idTokens, err := t.CountTokens(call.ID, model)
		if err != nil {
			return 0, err
		}
		total += idTokens

		// Count function name
		nameTokens, err := t.CountTokens(call.Function.Name, model)
		if err != nil {
			return 0, err
		}
		total += nameTokens

		// Count function arguments
		if call.Function.Arguments != "" {
			argTokens, err := t.CountTokens(call.Function.Arguments, model)
			if err != nil {
				return 0, err
			}
			total += argTokens
		}

		// Add overhead for call structure
		total += 5 // Approximate overhead for {"type":"function","id":"...","function":{...}}
	}

	return total, nil
}

// getMessageOverhead returns the per-message token overhead for a model.
func (t *TiktokenTokenizer) getMessageOverhead(model string) int {
	modelLower := strings.ToLower(model)
	if strings.HasPrefix(modelLower, "gpt-3.5") {
		return messageOverheadGPT35
	}
	return messageOverheadGPT4
}
