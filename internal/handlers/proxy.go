package handlers

import (
	"io"
	"net/http"
)

// OpenAIProxy forwards requests to OpenRouter faithfully
func (h *Repo) OpenAIProxy(w http.ResponseWriter, r *http.Request) {
	// 1. Define the upstream URL (OpenRouter)
	// You might want to make this configurable via env vars later
	targetURL := "https://openrouter.ai/api/v1/chat/completions"

	// 2. Create the upstream request
	// We use r.Context() to ensure if the client disconnects, we cancel the upstream request
	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// 3. Header Passthrough (Surgical)
	// We copy strictly what is needed.
	for k, v := range r.Header {
		// Skip hop-by-hop headers which Go might have already handled or shouldn't be forwarded
		if k == "Content-Length" || k == "Connection" || k == "Host" {
			continue
		}
		upstreamReq.Header[k] = v
	}

	// 4. Inject OpenRouter specific headers
	// OpenRouter requires these for rankings/visibility
	upstreamReq.Header.Set("HTTP-Referer", "https://github.com/mandalnilabja/goatway")
	upstreamReq.Header.Set("X-Title", "Goatway Proxy")

	// 5. Setup the Client
	// CRITICAL: DisableCompression is required for correct streaming.
	// If true, Go asks for gzip. If we just copy gzip bytes to the client,
	// the client (expecting text/event-stream) will fail to parse the chunks.
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	// 6. Execute Request
	resp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "Bad Gateway: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 7. Copy Response Headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	// 8. Stream Data (The "Pump")
	// We assert the writer is a Flusher to support streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		// This should practically never happen in standard Go HTTP servers
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Create a buffer (32KB is a standard clear balance between CPU and Syscalls)
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// Write the chunk we just read
			if _, wErr := w.Write(buf[:n]); wErr != nil {
				// Client disconnected or network error
				return
			}
			// FLUSH IMMEDIATELY. This is the "async" magic.
			// Without this, Go might buffer 4KB before sending, causing "laggy" streams.
			flusher.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			// Log error if needed, but the stream is already dirty
			break
		}
	}
}
