package tokenizer

import (
	"encoding/json"

	"github.com/mandalnilabja/goatway/internal/types"
)

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
