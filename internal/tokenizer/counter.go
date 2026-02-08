package tokenizer

import (
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

// getMessageOverhead returns the per-message token overhead for a model.
func (t *TiktokenTokenizer) getMessageOverhead(model string) int {
	modelLower := strings.ToLower(model)
	if strings.HasPrefix(modelLower, "gpt-3.5") {
		return messageOverheadGPT35
	}
	return messageOverheadGPT4
}
