package openrouter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/mandalnilabja/goatway/internal/types"
)

// StreamProcessor parses SSE chunks and extracts metadata.
type StreamProcessor struct {
	contentBuffer strings.Builder
	usage         *types.Usage
	finishReason  string
	model         string
}

// NewStreamProcessor creates a new SSE stream processor.
func NewStreamProcessor() *StreamProcessor {
	return &StreamProcessor{}
}

// ProcessReader reads and processes an SSE stream, calling onChunk for each raw chunk.
// Returns after the stream ends or on error.
func (p *StreamProcessor) ProcessReader(r io.Reader, onChunk func([]byte) error) error {
	scanner := bufio.NewScanner(r)
	// Set a larger buffer for potentially large chunks
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 256*1024)

	for scanner.Scan() {
		line := scanner.Bytes()

		// Forward the raw line plus newline
		chunk := append(line, '\n')
		if err := onChunk(chunk); err != nil {
			return err
		}

		// Parse SSE data lines
		p.processLine(line)
	}

	return scanner.Err()
}

// processLine parses a single SSE line.
func (p *StreamProcessor) processLine(line []byte) {
	// Skip empty lines and non-data lines
	if !bytes.HasPrefix(line, []byte("data: ")) {
		return
	}

	data := bytes.TrimPrefix(line, []byte("data: "))

	// Skip [DONE] marker
	if bytes.Equal(data, []byte("[DONE]")) {
		return
	}

	// Parse the JSON chunk
	var chunk types.ChatCompletionChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return // Skip malformed chunks
	}

	// Extract model if not set
	if p.model == "" && chunk.Model != "" {
		p.model = chunk.Model
	}

	// Extract usage from final chunk (if stream_options.include_usage=true)
	if chunk.Usage != nil {
		p.usage = chunk.Usage
	}

	// Process choices
	for _, choice := range chunk.Choices {
		// Accumulate content
		if choice.Delta.Content != "" {
			p.contentBuffer.WriteString(choice.Delta.Content)
		}

		// Extract finish reason
		if choice.FinishReason != nil && *choice.FinishReason != "" {
			p.finishReason = *choice.FinishReason
		}
	}
}

// GetContent returns the accumulated content from the stream.
func (p *StreamProcessor) GetContent() string {
	return p.contentBuffer.String()
}

// GetUsage returns the usage info if provided by upstream.
func (p *StreamProcessor) GetUsage() *types.Usage {
	return p.usage
}

// GetFinishReason returns the finish reason from the stream.
func (p *StreamProcessor) GetFinishReason() string {
	return p.finishReason
}

// GetModel returns the model from the stream.
func (p *StreamProcessor) GetModel() string {
	return p.model
}

// HasUpstreamUsage returns true if upstream provided usage info.
func (p *StreamProcessor) HasUpstreamUsage() bool {
	return p.usage != nil
}
