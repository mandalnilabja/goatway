# Azure AI Foundry Provider Implementation Plan

## Context

Add Azure AI Foundry as a new LLM provider to enable routing requests to Azure-hosted models (e.g., DeepSeek-R1, Mistral). Azure AI Foundry uses the Azure AI Model Inference API which is OpenAI-compatible but requires different authentication and dynamic endpoint construction.

## Key API Details

| Aspect | OpenRouter | Azure AI Foundry |
|--------|------------|------------------|
| **Endpoint** | Fixed: `https://openrouter.ai/api/v1/chat/completions` | Dynamic: `https://{endpoint}/models/chat/completions?api-version={version}` |
| **Auth Header** | `Authorization: Bearer {key}` | `api-key: {key}` |
| **API Version** | Not required | Required query param (default: `2024-05-01-preview`) |
| **Request/Response** | OpenAI-compatible | OpenAI-compatible (identical) |
| **Streaming** | SSE `text/event-stream` | SSE `text/event-stream` (identical) |

## Existing Infrastructure

The `AzureCredential` struct already exists in [credential.go](../../internal/storage/models/credential.go):

```go
type AzureCredential struct {
    Endpoint   string `json:"endpoint"`    // e.g., "myresource.services.ai.azure.com"
    APIKey     string `json:"api_key"`
    Deployment string `json:"deployment"`  // Can store default model for Foundry
    APIVersion string `json:"api_version"` // Default: "2024-05-01-preview"
}
```

Helper method `GetAzureCredential()` already exists.

---

## Implementation Steps

### Step 1: Create `internal/provider/azurefoundry/client.go`

**Purpose**: Main provider implementing `types.Provider` interface.

**Reference**: [openrouter/client.go](../../internal/provider/openrouter/client.go)

**Key differences from OpenRouter**:
- Dynamic URL from credential endpoint + api-version query param
- `api-key` header instead of `Authorization: Bearer`
- No provider-specific headers (no `HTTP-Referer`, `X-Title`)

```go
package azurefoundry

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/mandalnilabja/goatway/internal/types"
)

const defaultAPIVersion = "2024-05-01-preview"

type Provider struct{}

func New() *Provider {
    return &Provider{}
}

func (p *Provider) Name() string {
    return "azurefoundry"
}

func (p *Provider) BaseURL() string {
    return "" // URL built dynamically from credential
}

func (p *Provider) PrepareRequest(ctx context.Context, req *http.Request) error {
    return nil // No provider-specific headers needed
}

func (p *Provider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *types.ProxyOptions) (*types.ProxyResult, error) {
    startTime := time.Now()
    result := &types.ProxyResult{
        Model:        opts.Model,
        PromptTokens: opts.PromptTokens,
        IsStreaming:  opts.IsStreaming,
    }

    // Must have credential
    if opts.Credential == nil {
        result.Error = types.ErrNoAPIKey
        result.StatusCode = http.StatusUnauthorized
        http.Error(w, "No credential configured", http.StatusUnauthorized)
        return result, types.ErrNoAPIKey
    }

    // Extract Azure-specific credential
    azureCred, err := opts.Credential.GetAzureCredential()
    if err != nil {
        result.Error = err
        result.StatusCode = http.StatusUnauthorized
        http.Error(w, "Invalid Azure credential", http.StatusUnauthorized)
        return result, err
    }

    // Build URL with api-version
    apiVersion := azureCred.APIVersion
    if apiVersion == "" {
        apiVersion = defaultAPIVersion
    }
    targetURL := fmt.Sprintf("https://%s/models/chat/completions?api-version=%s",
        azureCred.Endpoint, apiVersion)

    // Rewrite body with resolved model
    body, err := rewriteModelInBody(opts.Body, req.Body, opts.Model)
    if err != nil {
        result.Error = err
        result.StatusCode = http.StatusBadRequest
        http.Error(w, "Failed to process request body", http.StatusBadRequest)
        return result, err
    }

    // Create upstream request
    upstreamReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, body)
    if err != nil {
        result.Error = err
        result.StatusCode = http.StatusInternalServerError
        http.Error(w, "Failed to create request", http.StatusInternalServerError)
        return result, err
    }

    // Copy headers (skip hop-by-hop)
    for k, v := range req.Header {
        if k == "Content-Length" || k == "Connection" || k == "Host" || k == "Authorization" {
            continue
        }
        upstreamReq.Header[k] = v
    }

    // Set Azure-style authentication
    upstreamReq.Header.Set("api-key", azureCred.APIKey)

    // Setup client (DisableCompression required for streaming)
    client := &http.Client{
        Transport: &http.Transport{
            DisableCompression: true,
        },
    }

    // Execute request
    resp, err := client.Do(upstreamReq)
    if err != nil {
        result.Error = err
        result.StatusCode = http.StatusBadGateway
        http.Error(w, "Bad Gateway: "+err.Error(), http.StatusBadGateway)
        return result, err
    }
    defer resp.Body.Close()

    result.StatusCode = resp.StatusCode
    result.Duration = time.Since(startTime)

    // Handle error responses
    if resp.StatusCode >= 400 {
        return handleErrorResponse(w, resp, result)
    }

    // Route based on content type
    contentType := resp.Header.Get("Content-Type")
    if strings.Contains(contentType, "text/event-stream") {
        return handleStreamingResponse(w, resp, result)
    }
    return handleJSONResponse(w, resp, result)
}
```

---

### Step 2: Create `internal/provider/azurefoundry/body.go`

**Purpose**: Request body rewriting to inject resolved model name.

**Action**: Copy from [openrouter/body.go](../../internal/provider/openrouter/body.go) - identical logic.

```go
package azurefoundry

import (
    "bytes"
    "encoding/json"
    "io"
)

// rewriteModelInBody reads the request body and replaces the model field.
func rewriteModelInBody(optsBody io.Reader, reqBody io.Reader, resolvedModel string) (io.Reader, error) {
    var body io.Reader = reqBody
    if optsBody != nil {
        body = optsBody
    }

    bodyBytes, err := io.ReadAll(body)
    if err != nil {
        return nil, err
    }

    var payload map[string]any
    if err := json.Unmarshal(bodyBytes, &payload); err != nil {
        return nil, err
    }

    payload["model"] = resolvedModel

    rewritten, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    return bytes.NewReader(rewritten), nil
}
```

---

### Step 3: Create `internal/provider/azurefoundry/response.go`

**Purpose**: Handle streaming and JSON responses.

**Action**: Copy from [openrouter/response.go](../../internal/provider/openrouter/response.go) - Azure uses identical OpenAI response format.

```go
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
```

---

### Step 4: Create `internal/provider/azurefoundry/stream.go`

**Purpose**: SSE stream processing.

**Action**: Copy from [openrouter/stream.go](../../internal/provider/openrouter/stream.go) - Azure uses identical SSE format.

```go
package azurefoundry

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
func (p *StreamProcessor) ProcessReader(r io.Reader, onChunk func([]byte) error) error {
    scanner := bufio.NewScanner(r)
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
    if !bytes.HasPrefix(line, []byte("data: ")) {
        return
    }

    data := bytes.TrimPrefix(line, []byte("data: "))

    if bytes.Equal(data, []byte("[DONE]")) {
        return
    }

    var chunk types.ChatCompletionChunk
    if err := json.Unmarshal(data, &chunk); err != nil {
        return
    }

    if p.model == "" && chunk.Model != "" {
        p.model = chunk.Model
    }

    if chunk.Usage != nil {
        p.usage = chunk.Usage
    }

    for _, choice := range chunk.Choices {
        if choice.Delta.Content != "" {
            p.contentBuffer.WriteString(choice.Delta.Content)
        }
        if choice.FinishReason != nil && *choice.FinishReason != "" {
            p.finishReason = *choice.FinishReason
        }
    }
}

func (p *StreamProcessor) GetContent() string     { return p.contentBuffer.String() }
func (p *StreamProcessor) GetUsage() *types.Usage { return p.usage }
func (p *StreamProcessor) GetFinishReason() string { return p.finishReason }
func (p *StreamProcessor) GetModel() string       { return p.model }
func (p *StreamProcessor) HasUpstreamUsage() bool { return p.usage != nil }
```

---

### Step 5: Register provider in `internal/provider/registry.go`

**File**: [registry.go](../../internal/provider/registry.go)

```go
package provider

import (
    "github.com/mandalnilabja/goatway/internal/provider/openrouter"
    "github.com/mandalnilabja/goatway/internal/provider/azurefoundry"  // ADD
)

func NewProviders() map[string]Provider {
    return map[string]Provider{
        "openrouter":   openrouter.New(),
        "azurefoundry": azurefoundry.New(),  // ADD
    }
}
```

---

### Step 6: Update credential masking in `internal/storage/models/credential.go`

**File**: [credential.go](../../internal/storage/models/credential.go) line 60

**Change**: Update switch case to handle both providers.

```go
case "azure", "azurefoundry":  // Modified from just "azure"
    var cred AzureCredential
    // ... rest unchanged
```

---

### Step 7: Update default config template in `internal/config/file.go`

**File**: [file.go](../../internal/config/file.go)

**Change**: Add commented example in `defaultConfig` string.

```toml
# Azure AI Foundry example
# [[models]]
# slug = "deepseek-r1"
# provider = "azurefoundry"
# model = "DeepSeek-R1"
# credential_name = "my-azure-foundry-key"
```

---

## Files Summary

| Action | File | Est. Lines |
|--------|------|------------|
| CREATE | `internal/provider/azurefoundry/client.go` | ~85 |
| CREATE | `internal/provider/azurefoundry/body.go` | ~35 |
| CREATE | `internal/provider/azurefoundry/response.go` | ~95 |
| CREATE | `internal/provider/azurefoundry/stream.go` | ~85 |
| MODIFY | `internal/provider/registry.go` | +2 lines |
| MODIFY | `internal/storage/models/credential.go` | +1 word |
| MODIFY | `internal/config/file.go` | +5 lines |

---

## Critical Streaming Rules

Per project constraints (CLAUDE.md), the implementation MUST:

1. `http.Transport.DisableCompression = true`
2. Call `Flusher.Flush()` after each SSE write
3. Never buffer full responses
4. Propagate client context end-to-end
5. No retries or background goroutines in request path

---

## Credential Configuration

Users create credentials with provider `azurefoundry`:

```json
{
  "provider": "azurefoundry",
  "name": "my-foundry-key",
  "data": {
    "endpoint": "myresource.services.ai.azure.com",
    "api_key": "your-api-key",
    "api_version": "2024-05-01-preview"
  }
}
```

## Config Example

```toml
[[models]]
slug = "deepseek"
provider = "azurefoundry"
model = "DeepSeek-R1"
credential_name = "my-foundry-key"
```

---

## Verification

1. **Build**: `make build`
2. **Tests**: `make test`
3. **Lint**: `make lint`
4. **Manual test**:
   ```bash
   # Create credential
   curl -X POST http://localhost:8080/v1/admin/credentials \
     -H "Content-Type: application/json" \
     -d '{
       "provider": "azurefoundry",
       "name": "test-azure",
       "data": {
         "endpoint": "myresource.services.ai.azure.com",
         "api_key": "your-key",
         "api_version": "2024-05-01-preview"
       }
     }'

   # Test chat completion
   curl http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "DeepSeek-R1",
       "messages": [{"role": "user", "content": "Hello"}]
     }'
   ```
