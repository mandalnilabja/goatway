package azurefoundry

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/mandalnilabja/goatway/internal/types"
)

// handleStreamingResponse processes SSE streaming responses.
func handleStreamingResponse(w http.ResponseWriter, resp *http.Response, result *types.ProxyResult) (*types.ProxyResult, error) {
	// Copy headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		result.Error = io.ErrNoProgress
		return result, nil
	}

	// Process stream while forwarding to client
	processor := NewStreamProcessor()
	err := processor.ProcessReader(resp.Body, func(chunk []byte) error {
		if _, wErr := w.Write(chunk); wErr != nil {
			return wErr
		}
		flusher.Flush()
		return nil
	})

	// Extract results from processor
	result.FinishReason = processor.GetFinishReason()
	if processor.GetModel() != "" {
		result.Model = processor.GetModel()
	}

	// Use upstream usage if available
	if usage := processor.GetUsage(); usage != nil {
		result.PromptTokens = usage.PromptTokens
		result.CompletionTokens = usage.CompletionTokens
		result.TotalTokens = usage.TotalTokens
	}

	if err != nil {
		result.Error = err
	}
	return result, err
}

// handleJSONResponse processes non-streaming JSON responses.
func handleJSONResponse(w http.ResponseWriter, resp *http.Response, result *types.ProxyResult) (*types.ProxyResult, error) {
	// Read full response for parsing
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		http.Error(w, "Failed to read response", http.StatusBadGateway)
		return result, err
	}

	// Parse response to extract usage
	var completion types.ChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err == nil {
		if completion.Usage != nil {
			result.PromptTokens = completion.Usage.PromptTokens
			result.CompletionTokens = completion.Usage.CompletionTokens
			result.TotalTokens = completion.Usage.TotalTokens
		}
		if len(completion.Choices) > 0 {
			result.FinishReason = completion.Choices[0].FinishReason
		}
		if completion.Model != "" {
			result.Model = completion.Model
		}
	}

	// Forward response to client
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)

	return result, nil
}

// handleErrorResponse forwards error responses and extracts error info.
func handleErrorResponse(w http.ResponseWriter, resp *http.Response, result *types.ProxyResult) (*types.ProxyResult, error) {
	body, _ := io.ReadAll(resp.Body)

	// Try to extract error message
	var apiErr types.APIError
	if err := json.Unmarshal(body, &apiErr); err == nil {
		result.ErrorMessage = apiErr.Error.Message
	}

	// Forward error to client
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)

	return result, nil
}
