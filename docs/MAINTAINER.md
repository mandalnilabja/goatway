# Goatway Maintainer Documentation

This document provides comprehensive technical documentation for engineers maintaining the Goatway codebase.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Directory Structure](#directory-structure)
4. [Request Flow](#request-flow)
5. [Streaming Implementation](#streaming-implementation)
6. [Core Components](#core-components)
7. [Database Schema](#database-schema)
8. [Configuration](#configuration)
9. [API Reference](#api-reference)
10. [Adding New Providers](#adding-new-providers)
11. [Development Workflow](#development-workflow)

---

## Overview

Goatway is a **lightweight, streaming-safe, OpenAI-compatible HTTP proxy** for routing requests to different LLM providers. It is written in Go with minimal dependencies and prioritizes:

- **Streaming correctness** over features
- **Low latency** over abstraction
- **Explicitness** over magic
- **Stability** over refactors

### Core Capabilities

- OpenAI-compatible `/v1/chat/completions` endpoint
- SSE (Server-Sent Events) streaming support
- Multi-provider support (currently OpenRouter)
- Credential management with encrypted storage
- Usage tracking and logging
- Token counting for requests
- Admin API for management
- Optional Web UI

---

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Client                               │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Middleware Stack                            │
│   ┌─────────┐  ┌──────────────────┐  ┌───────────────────────┐  │
│   │  CORS   │→ │   Request ID     │→ │   Request Logger      │  │
│   └─────────┘  └──────────────────┘  └───────────────────────┘  │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Router (http.ServeMux)                   │
│  ┌──────────────────┐  ┌─────────────────┐  ┌────────────────┐  │
│  │ /v1/chat/complete│  │    /api/admin/* │  │   /api/health  │  │
│  └──────────────────┘  └─────────────────┘  └────────────────┘  │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Handler Repository                          │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌─────────────┐   │
│  │  Provider │  │  Storage  │  │   Cache   │  │  Tokenizer  │   │
│  └───────────┘  └───────────┘  └───────────┘  └─────────────┘   │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Provider Layer                              │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │               OpenRouter Provider                           │ │
│  │  - PrepareRequest()  - ProxyRequest()  - BaseURL()         │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Upstream LLM Provider                         │
│                  (e.g., OpenRouter API)                          │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| `main.go` | Application entry point, dependency wiring |
| `Router` | HTTP route registration and middleware chain |
| `Handler Repository` | HTTP handlers with injected dependencies |
| `Provider` | LLM provider abstraction and proxying |
| `Storage` | SQLite persistence layer |
| `Tokenizer` | Token counting for requests |
| `Middleware` | Cross-cutting concerns (CORS, auth, logging) |

---

## Directory Structure

```
goatway/
├── cmd/
│   └── api/
│       └── main.go              # Entry point: wires config → provider → handlers → router → server
│
├── internal/
│   ├── app/
│   │   ├── router.go            # Route registration (http.ServeMux), middleware chain
│   │   └── server.go            # HTTP server wrapper with timeouts
│   │
│   ├── config/
│   │   ├── config.go            # Environment-based configuration loading
│   │   └── paths.go             # Data directory and file path resolution
│   │
│   ├── provider/
│   │   ├── provider.go          # Provider interface definition
│   │   └── openrouter/
│   │       ├── client.go        # OpenRouter provider implementation
│   │       ├── response.go      # Response handling (streaming/JSON/error)
│   │       └── stream.go        # SSE stream processor
│   │
│   ├── storage/
│   │   ├── storage.go           # Storage interface definition and factory
│   │   ├── argon2.go            # Argon2 password hashing
│   │   ├── keygen.go            # API key generation
│   │   ├── models/
│   │   │   ├── credential.go    # Credential model
│   │   │   ├── apikey.go        # Client API key model
│   │   │   ├── log.go           # Request log model
│   │   │   └── usage.go         # Usage statistics models
│   │   ├── sqlite/
│   │   │   ├── sqlite.go        # SQLite storage implementation and schema
│   │   │   ├── credentials.go   # Credential CRUD operations
│   │   │   ├── apikeys.go       # API key CRUD operations
│   │   │   ├── logs.go          # Request logging operations
│   │   │   ├── usage.go         # Usage statistics operations
│   │   │   ├── admin.go         # Admin settings operations
│   │   │   ├── helpers.go       # SQL helper functions
│   │   │   └── errors.go        # Storage error definitions
│   │   └── encryption/
│   │       └── aes.go           # AES-256-GCM encryption for API keys
│   │
│   ├── tokenizer/
│   │   ├── tokenizer.go         # Tokenizer interface and tiktoken implementation
│   │   ├── counter.go           # Token counting for full requests
│   │   ├── counter_content.go   # Content token counting
│   │   └── counter_tools.go     # Tool definition token counting
│   │
│   ├── transport/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── handler.go       # Handler repository (composes domain handlers)
│   │       │   ├── admin/
│   │       │   │   ├── admin.go         # Admin handlers constructor
│   │       │   │   ├── credentials.go   # Credential management endpoints
│   │       │   │   ├── apikeys.go       # API key management endpoints
│   │       │   │   ├── usage.go         # Usage statistics endpoints
│   │       │   │   └── system.go        # System info endpoints
│   │       │   ├── proxy/
│   │       │   │   ├── proxy.go         # Proxy handlers constructor and shared logic
│   │       │   │   ├── chat.go          # POST /v1/chat/completions
│   │       │   │   ├── completions.go   # POST /v1/completions (legacy)
│   │       │   │   ├── models.go        # GET /v1/models
│   │       │   │   ├── embeddings.go    # POST /v1/embeddings
│   │       │   │   ├── audio.go         # Audio endpoints
│   │       │   │   ├── images.go        # Image endpoints
│   │       │   │   └── moderations.go   # Moderation endpoint
│   │       │   ├── webui/
│   │       │   │   ├── webui.go         # Web UI handlers constructor
│   │       │   │   ├── serve.go         # Static file serving
│   │       │   │   └── auth.go          # Login, Logout, LoginPage
│   │       │   ├── infra/
│   │       │   │   ├── infra.go         # Infrastructure handlers constructor
│   │       │   │   ├── health.go        # Health check endpoints
│   │       │   │   └── cache.go         # Cache demo endpoint
│   │       │   └── shared/
│   │       │       └── shared.go        # writeJSON, writeError utilities
│   │       └── middleware/
│   │           ├── cors.go          # CORS middleware
│   │           ├── requestid.go     # Request ID middleware
│   │           ├── logging.go       # Request logging middleware
│   │           └── auth/
│   │               ├── admin.go     # Admin authentication
│   │               ├── apikey.go    # API key authentication
│   │               └── session.go   # Session authentication
│   │
│   └── types/
│       ├── request.go           # ChatCompletionRequest types
│       ├── response.go          # ChatCompletionResponse types
│       ├── stream.go            # Streaming chunk types
│       ├── message.go           # Message and Content types
│       ├── tools.go             # Tool/Function calling types
│       ├── errors.go            # OpenAI-compatible error types
│       ├── completions.go       # Legacy completions types
│       ├── embeddings.go        # Embeddings types
│       ├── audio.go             # Audio types
│       ├── images.go            # Image types
│       ├── moderations.go       # Moderation types
│       └── json.go              # JSON marshaling helpers
│
├── web/
│   ├── embed.go                 # Embedded web UI assets
│   ├── index.html               # Web UI HTML
│   └── static/                  # CSS and JS assets
│
├── docs/
│   └── MAINTAINER.md            # This file
│
├── CLAUDE.md                    # AI agent instructions
├── AGENTS.md                    # Mandatory rules for AI agents
├── CONTRIBUTING.md              # Contribution guidelines
├── README.md                    # Project overview
├── Makefile                     # Build and development commands
└── go.mod                       # Go module definition
```

---

## Request Flow

### Proxy Request Flow (`POST /v1/chat/completions`)

```
1. Request arrives at server
   │
2. Middleware chain executes:
   │  ├── CORS headers added
   │  ├── Request ID generated/extracted
   │  └── Request logged
   │
3. Router dispatches to OpenAIProxy handler
   │
4. Handler processes request:
   │  ├── Read and buffer request body
   │  ├── Parse JSON to ChatCompletionRequest
   │  ├── Resolve API key (header or default credential)
   │  └── Count prompt tokens
   │
5. Provider.ProxyRequest() called:
   │  ├── Create upstream request with body
   │  ├── Copy headers (filter hop-by-hop)
   │  ├── Add Authorization header
   │  ├── Add provider-specific headers
   │  └── Execute request with DisableCompression=true
   │
6. Response handling:
   │  ├── If text/event-stream: handleStreamingResponse()
   │  │   ├── Copy response headers
   │  │   ├── Process SSE chunks via StreamProcessor
   │  │   ├── Forward each chunk immediately
   │  │   ├── Flush after each write
   │  │   └── Extract usage/model from final chunk
   │  │
   │  └── If JSON: handleJSONResponse()
   │      ├── Read full response
   │      ├── Parse for usage stats
   │      └── Forward to client
   │
7. Async logging:
   │  ├── Log request to storage
   │  └── Update daily usage aggregates
   │
8. Response complete
```

### Key Files in Request Flow

| Step | File | Function |
|------|------|----------|
| Routing | [router.go](../internal/app/router.go) | `NewRouter()` |
| Handler | [chat.go](../internal/transport/http/handler/proxy/chat.go) | `ChatCompletions()` |
| API Key | [proxy.go](../internal/transport/http/handler/proxy/proxy.go) | `resolveAPIKey()` |
| Proxying | [client.go](../internal/provider/openrouter/client.go) | `ProxyRequest()` |
| Streaming | [response.go](../internal/provider/openrouter/response.go) | `handleStreamingResponse()` |
| SSE Parsing | [stream.go](../internal/provider/openrouter/stream.go) | `StreamProcessor` |
| Logging | [chat.go](../internal/transport/http/handler/proxy/chat.go) | `logRequest()` |

---

## Streaming Implementation

Streaming is the most critical aspect of Goatway. Any changes to streaming code require extreme care.

### Critical Rules

1. **`http.Transport.DisableCompression` MUST be `true`**
   - Prevents gzip encoding which breaks SSE parsing
   - Located in [client.go](../internal/provider/openrouter/client.go)

2. **Flush after every write**
   - SSE requires immediate flushing
   - Never buffer chunks
   - Located in [response.go](../internal/provider/openrouter/response.go)

3. **No buffering or accumulation**
   - Stream processor observes but doesn't transform
   - Chunks forwarded as-is

4. **Context propagation**
   - Client context flows to upstream
   - Cancellation handled properly

### Streaming Architecture

```go
// openrouter_response.go - handleStreamingResponse()

// 1. Get flusher interface
flusher, ok := w.(http.Flusher)

// 2. Process stream with callback
processor := NewStreamProcessor()
err := processor.ProcessReader(resp.Body, func(chunk []byte) error {
    // 3. Write chunk immediately
    if _, wErr := w.Write(chunk); wErr != nil {
        return wErr
    }
    // 4. Flush to client
    flusher.Flush()
    return nil
})

// 5. Extract metadata from processor
result.FinishReason = processor.GetFinishReason()
result.Model = processor.GetModel()
```

### StreamProcessor

The `StreamProcessor` in [stream.go](../internal/provider/openrouter/stream.go) parses SSE chunks while forwarding them:

```go
type StreamProcessor struct {
    contentBuffer strings.Builder  // Accumulated content (for metrics)
    usage         *types.Usage     // Usage from final chunk
    finishReason  string           // Finish reason
    model         string           // Model name
}
```

It processes lines looking for `data: ` prefixes and parses JSON to extract metadata without modifying the stream.

---

## Core Components

### Provider Interface

Defined in [provider.go](../internal/provider/provider.go):

```go
type Provider interface {
    Name() string
    BaseURL() string
    PrepareRequest(ctx context.Context, req *http.Request) error
    ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *ProxyOptions) (*ProxyResult, error)
}
```

| Method | Purpose |
|--------|---------|
| `Name()` | Provider identifier (e.g., "openrouter") |
| `BaseURL()` | API endpoint URL |
| `PrepareRequest()` | Add provider-specific headers |
| `ProxyRequest()` | Execute the proxy with streaming support |

### ProxyOptions

```go
type ProxyOptions struct {
    APIKey       string     // API key for this request
    RequestID    string     // Request tracing ID
    PromptTokens int        // Pre-calculated token count
    Model        string     // Model from request
    IsStreaming  bool       // Streaming request flag
    Body         io.Reader  // Request body (already read)
}
```

### ProxyResult

```go
type ProxyResult struct {
    Model            string
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    StatusCode       int
    FinishReason     string
    Duration         time.Duration
    IsStreaming      bool
    Error            error
    ErrorMessage     string
}
```

### Handler Repository

The `Repo` struct in [handler.go](../internal/transport/http/handler/handler.go) composes domain-specific handlers:

```go
type Repo struct {
    Admin *admin.Handlers
    WebUI *webui.Handlers
    Proxy *proxy.Handlers
    Infra *infra.Handlers
}
```

Each domain package (admin, webui, proxy, infra) has its own `Handlers` struct with only the dependencies it needs, enabling clean separation of concerns.

### Storage Interface

Defined in [storage.go](../internal/storage/storage.go):

```go
type Storage interface {
    // Credential operations
    CreateCredential(cred *Credential) error
    GetCredential(id string) (*Credential, error)
    GetDefaultCredential(provider string) (*Credential, error)
    ListCredentials() ([]*Credential, error)
    UpdateCredential(cred *Credential) error
    DeleteCredential(id string) error
    SetDefaultCredential(id string) error

    // Request logging
    LogRequest(log *RequestLog) error
    GetRequestLogs(filter LogFilter) ([]*RequestLog, error)
    DeleteRequestLogs(olderThan string) (int64, error)

    // Usage statistics
    GetUsageStats(filter StatsFilter) (*UsageStats, error)
    GetDailyUsage(startDate, endDate string) ([]*DailyUsage, error)
    UpdateDailyUsage(usage *DailyUsage) error

    // Maintenance
    Close() error
}
```

### Tokenizer Interface

Defined in [tokenizer.go](../internal/tokenizer/tokenizer.go):

```go
type Tokenizer interface {
    CountTokens(text string, model string) (int, error)
    CountMessages(messages []types.Message, model string) (int, error)
    CountRequest(req *types.ChatCompletionRequest) (int, error)
}
```

Uses tiktoken-go with encoding selection based on model prefix:
- `gpt-4o`, `o1`, `o3`, `chatgpt` → `o200k_base`
- `gpt-4`, `gpt-3.5`, `text-embedding` → `cl100k_base`

---

## Database Schema

SQLite database with WAL mode for better concurrency.

### Tables

#### credentials

```sql
CREATE TABLE credentials (
    id          TEXT PRIMARY KEY,
    provider    TEXT NOT NULL,
    name        TEXT NOT NULL,
    api_key     TEXT NOT NULL,      -- Encrypted with AES-256-GCM
    is_default  INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### request_logs

```sql
CREATE TABLE request_logs (
    id                TEXT PRIMARY KEY,
    request_id        TEXT NOT NULL,
    credential_id     TEXT,           -- FK to credentials
    model             TEXT NOT NULL,
    provider          TEXT NOT NULL,
    prompt_tokens     INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens      INTEGER DEFAULT 0,
    is_streaming      INTEGER DEFAULT 0,
    status_code       INTEGER,
    error_message     TEXT,
    duration_ms       INTEGER,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### usage_daily

```sql
CREATE TABLE usage_daily (
    date              TEXT NOT NULL,
    credential_id     TEXT,           -- FK to credentials
    model             TEXT NOT NULL,
    request_count     INTEGER DEFAULT 0,
    prompt_tokens     INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens      INTEGER DEFAULT 0,
    error_count       INTEGER DEFAULT 0,
    PRIMARY KEY (date, credential_id, model)
);
```

### Indexes

```sql
CREATE INDEX idx_logs_created ON request_logs(created_at);
CREATE INDEX idx_logs_model ON request_logs(model);
CREATE INDEX idx_logs_credential ON request_logs(credential_id);
CREATE INDEX idx_usage_date ON usage_daily(date);
CREATE INDEX idx_creds_provider ON credentials(provider);
```

### Encryption

API keys are encrypted at rest using AES-256-GCM. The encryption key is derived from:

1. `GOATWAY_ENCRYPTION_KEY` environment variable (if set)
2. Machine-specific key (hostname + home dir + OS/arch)

See [aes.go](../internal/storage/encryption/aes.go) for implementation.

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDR` | `:8080` | Server listen address |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `text` | Log format (text, json) |
| `LLM_PROVIDER` | `openrouter` | Default LLM provider |
| `OPENROUTER_API_KEY` | | OpenRouter API key |
| `GOATWAY_DATA_DIR` | | Data directory override |
| `GOATWAY_ENCRYPTION_KEY` | | Encryption key for API keys |
| `GOATWAY_ADMIN_PASSWORD` | | Admin API password |
| `ENABLE_WEB_UI` | `true` | Enable web UI |

### CLI Flags

```bash
goatway [flags]

Flags:
  -addr string        Server address (overrides SERVER_ADDR)
  -data-dir string    Data directory (overrides GOATWAY_DATA_DIR)
  -version            Print version and exit
  -v                  Print version and exit (shorthand)
```

### Data Directory Resolution

Priority order (see [paths.go](../internal/config/paths.go)):

1. `GOATWAY_DATA_DIR` environment variable
2. `$XDG_DATA_HOME/goatway` (Linux)
3. `%APPDATA%\goatway` (Windows)
4. `~/.goatway` (default)

---

## API Reference

### Proxy Endpoints

#### POST /v1/chat/completions

OpenAI-compatible chat completion endpoint.

**Request:**
```json
{
  "model": "openai/gpt-4o",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "stream": true
}
```

**Headers:**
- `Authorization: Bearer <api_key>` (optional if default credential set)
- `Content-Type: application/json`

**Response:** OpenAI-compatible streaming or JSON response.

#### GET /v1/models

List available models from upstream provider.

#### GET /v1/models/{model}

Get details for a specific model.

### Admin Endpoints

All admin endpoints support optional authentication via `Authorization: Bearer <admin_password>`.

#### Credentials

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/credentials` | Create credential |
| GET | `/api/admin/credentials` | List credentials |
| GET | `/api/admin/credentials/{id}` | Get credential |
| PUT | `/api/admin/credentials/{id}` | Update credential |
| DELETE | `/api/admin/credentials/{id}` | Delete credential |
| POST | `/api/admin/credentials/{id}/default` | Set as default |

#### Usage & Logs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/usage` | Get usage statistics |
| GET | `/api/admin/usage/daily` | Get daily usage breakdown |
| GET | `/api/admin/logs` | Get request logs |
| DELETE | `/api/admin/logs` | Delete old logs |

#### System

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/health` | Health check with DB status |
| GET | `/api/admin/info` | System info and stats |

### Health Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Basic health check |
| GET | `/` | Home page |

---

## Adding New Providers

To add a new LLM provider:

### 1. Create Provider File

Create `internal/provider/<name>.go`:

```go
package provider

type NewProvider struct {
    APIKey string
}

func NewNewProvider(apiKey string) *NewProvider {
    return &NewProvider{APIKey: apiKey}
}

func (p *NewProvider) Name() string {
    return "newprovider"
}

func (p *NewProvider) BaseURL() string {
    return "https://api.newprovider.com/v1/chat/completions"
}

func (p *NewProvider) PrepareRequest(ctx context.Context, req *http.Request) error {
    // Add provider-specific headers
    return nil
}

func (p *NewProvider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *ProxyOptions) (*ProxyResult, error) {
    // Can reuse OpenRouter's implementation if API is compatible
    // Or implement custom logic
}
```

### 2. Update Configuration

Add to [config.go](../internal/config/config.go):

```go
type Config struct {
    // ...
    NewProviderAPIKey string
}

func Load() *Config {
    return &Config{
        // ...
        NewProviderAPIKey: getEnv("NEWPROVIDER_API_KEY", ""),
    }
}
```

### 3. Update Main

Add provider case in [main.go](../cmd/api/main.go):

```go
switch cfg.Provider {
case "openrouter":
    llmProvider = provider.NewOpenRouterProvider(cfg.OpenRouterAPIKey)
case "newprovider":
    llmProvider = provider.NewNewProvider(cfg.NewProviderAPIKey)
default:
    // ...
}
```

### 4. Important Considerations

- **Streaming:** Must follow streaming rules exactly
- **Headers:** Filter hop-by-hop headers
- **Compression:** `DisableCompression: true` is mandatory
- **Testing:** Add tests for new provider

---

## Development Workflow

### Build Commands

```bash
make build        # Build binary to bin/goatway
make run          # Run the server (go run)
make test         # Run all tests
make fmt          # Format code with goimports
make fmt-check    # Check formatting without modifying
make lint         # Run golangci-lint
make tools        # Install dev tools
make clean        # Remove build artifacts
make build-all    # Build for all platforms
make release      # Create release archives
```

### Running Locally

```bash
# With environment variables
export OPENROUTER_API_KEY="your-key"
make run

# With CLI flags
./bin/goatway -addr :9000 -data-dir ./data
```

### Testing

```bash
# Run all tests
make test

# Run specific test
go test -v ./internal/storage/...

# Run with coverage
go test -cover ./...
```

### Code Style

- Use `goimports` for formatting
- Follow conventional commits: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`
- Keep files under 200 lines
- One file = one responsibility
- Prefer standard library over external dependencies

### Key Constraints

From [AGENTS.md](../AGENTS.md):

- **No breaking streaming** - Any change that risks streaming is out of bounds
- **No large refactors** without design discussion
- **No new dependencies** without approval
- **No frameworks** or middleware layers
- **Table-driven tests** preferred
- **Mock external HTTP** calls in tests

---

## Troubleshooting

### Common Issues

#### Streaming Not Working

1. Check `DisableCompression: true` in HTTP transport
2. Verify `Flush()` is called after each write
3. Ensure no middleware is buffering responses

#### API Key Not Found

1. Check `Authorization: Bearer` header format
2. Verify default credential is set in storage
3. Check credential provider matches request

#### Database Errors

1. Verify data directory is writable
2. Check SQLite file permissions (0700)
3. Look for WAL file corruption

#### High Memory Usage

1. Check for unbounded log retention
2. Verify cache configuration
3. Look for goroutine leaks

### Debug Logging

```bash
LOG_LEVEL=debug make run
```

### Database Inspection

```bash
sqlite3 ~/.goatway/goatway.db
.tables
.schema credentials
SELECT * FROM credentials;
```

---

## References

- [CLAUDE.md](../CLAUDE.md) - AI agent development instructions
- [AGENTS.md](../AGENTS.md) - Mandatory rules for code changes
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [README.md](../README.md) - Project overview
