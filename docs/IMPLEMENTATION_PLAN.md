# Implementation Plan: Local-First OpenAI-Compatible Proxy

## Executive Summary

Transform Goatway into a **local-first, standalone** OpenAI-compatible proxy with:
- Single binary distribution (easy install via Homebrew, curl, or Go install)
- Local SQLite database for credentials, token usage, and request logs
- Built-in Web UI for credential management (add/edit/delete API keys)
- Per-request token counting with persistent local storage
- Complete OpenAI API format compliance

**Design Principle**: Everything runs locally. No cloud dependencies. Your data stays on your machine.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         GOATWAY                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐    │
│  │   Web UI     │     │  Proxy API   │     │  Admin API   │    │
│  │  (Embedded)  │     │ /v1/chat/... │     │ /api/admin/* │    │
│  │  Port 8080   │     │  Port 8080   │     │  Port 8080   │    │
│  └──────┬───────┘     └──────┬───────┘     └──────┬───────┘    │
│         │                    │                     │            │
│         └────────────────────┼─────────────────────┘            │
│                              │                                   │
│                    ┌─────────▼─────────┐                        │
│                    │   Handler Layer   │                        │
│                    │   (Repo struct)   │                        │
│                    └─────────┬─────────┘                        │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         │                    │                    │             │
│  ┌──────▼──────┐     ┌──────▼──────┐     ┌──────▼──────┐      │
│  │  Provider   │     │  Tokenizer  │     │   Storage   │      │
│  │  (OpenRouter)│     │  (tiktoken) │     │  (SQLite)   │      │
│  └─────────────┘     └─────────────┘     └─────────────┘      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  ~/.goatway/                   │
              │  ├── goatway.db (SQLite)      │
              │  └── config.yaml (optional)   │
              └───────────────────────────────┘
```

---

## Current State Analysis

### Existing Architecture (Post-Refactor)

```
cmd/api/main.go                          → Entry point (minimal wiring)
internal/config/config.go                → Environment-based configuration
internal/provider/
  ├── provider.go                        → Provider interface
  └── openrouter.go                      → OpenRouter implementation
internal/app/
  ├── router.go                          → Route registration
  └── server.go                          → Server lifecycle
internal/transport/http/handler/
  ├── repo.go                            → Handler repository (Cache, Provider)
  ├── health.go                          → Health check endpoints
  ├── cache.go                           → Cache demo handler
  └── proxy.go                           → Provider-agnostic proxy
internal/storage/                        → ✅ COMPLETED (Phase 1)
  ├── storage.go                         → Storage interface
  ├── models.go                          → Data models
  ├── encryption.go                      → AES-256-GCM encryption
  ├── sqlite.go                          → SQLite implementation
  ├── sqlite_test.go                     → Storage tests
  └── encryption_test.go                 → Encryption tests
pkg/logger/                              → Empty (placeholder)
```

### Gaps for Local-First Implementation

1. ~~No persistent storage~~ → ✅ SQLite storage layer implemented
2. No Web UI for configuration
3. ~~No local logging/analytics database~~ → ✅ Schema implemented, not yet integrated
4. No embedded static file serving
5. No admin API for credential management
6. Distribution not streamlined
7. Storage not integrated into proxy flow
8. No data directory management (paths.go)

---

## Implementation Phases

### Phase 1: SQLite Storage Layer
**Goal**: Local persistent storage for all data

#### New File: `internal/storage/sqlite.go` (~150 lines)

**Database Schema:**
```sql
-- Credentials table (encrypted API keys)
CREATE TABLE credentials (
    id          TEXT PRIMARY KEY,
    provider    TEXT NOT NULL,           -- 'openrouter', 'openai', 'anthropic'
    name        TEXT NOT NULL,           -- User-friendly name
    api_key     TEXT NOT NULL,           -- Encrypted with local key
    is_default  INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Request logs table
CREATE TABLE request_logs (
    id                TEXT PRIMARY KEY,
    request_id        TEXT NOT NULL,
    credential_id     TEXT,
    model             TEXT NOT NULL,
    provider          TEXT NOT NULL,
    prompt_tokens     INTEGER,
    completion_tokens INTEGER,
    total_tokens      INTEGER,
    is_streaming      INTEGER,
    status_code       INTEGER,
    error_message     TEXT,
    duration_ms       INTEGER,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (credential_id) REFERENCES credentials(id)
);

-- Usage aggregates (daily rollups for fast queries)
CREATE TABLE usage_daily (
    date              TEXT NOT NULL,      -- YYYY-MM-DD
    credential_id     TEXT,
    model             TEXT NOT NULL,
    request_count     INTEGER DEFAULT 0,
    prompt_tokens     INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens      INTEGER DEFAULT 0,
    error_count       INTEGER DEFAULT 0,

    PRIMARY KEY (date, credential_id, model),
    FOREIGN KEY (credential_id) REFERENCES credentials(id)
);

-- Indexes for common queries
CREATE INDEX idx_logs_created ON request_logs(created_at);
CREATE INDEX idx_logs_model ON request_logs(model);
CREATE INDEX idx_logs_credential ON request_logs(credential_id);
CREATE INDEX idx_usage_date ON usage_daily(date);
```

**Storage Interface:**
```go
type Storage interface {
    // Credentials
    CreateCredential(cred *Credential) error
    GetCredential(id string) (*Credential, error)
    GetDefaultCredential(provider string) (*Credential, error)
    ListCredentials() ([]*Credential, error)
    UpdateCredential(cred *Credential) error
    DeleteCredential(id string) error
    SetDefaultCredential(id string) error

    // Request Logging
    LogRequest(log *RequestLog) error
    GetRequestLogs(filter LogFilter) ([]*RequestLog, error)

    // Usage Stats
    GetUsageStats(filter StatsFilter) (*UsageStats, error)
    GetDailyUsage(startDate, endDate string) ([]*DailyUsage, error)

    // Maintenance
    Close() error
    Migrate() error
}
```

#### New File: `internal/storage/models.go` (~80 lines)

**Data Models:**
```go
type Credential struct {
    ID        string
    Provider  string    // openrouter, openai, anthropic
    Name      string    // "My OpenRouter Key"
    APIKey    string    // Encrypted at rest
    IsDefault bool
    CreatedAt time.Time
    UpdatedAt time.Time
}

type RequestLog struct {
    ID               string
    RequestID        string
    CredentialID     string
    Model            string
    Provider         string
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    IsStreaming      bool
    StatusCode       int
    ErrorMessage     string
    DurationMs       int64
    CreatedAt        time.Time
}

type UsageStats struct {
    TotalRequests     int
    TotalTokens       int
    TotalPromptTokens int
    TotalCompletionTokens int
    ErrorCount        int
    ModelBreakdown    map[string]*ModelStats
}

type DailyUsage struct {
    Date              string
    CredentialID      string
    Model             string
    RequestCount      int
    PromptTokens      int
    CompletionTokens  int
    TotalTokens       int
    ErrorCount        int
}
```

#### New File: `internal/storage/encryption.go` (~60 lines)

**Local Encryption for API Keys:**
```go
// Uses machine-specific key derivation
// Keys encrypted with AES-256-GCM
// Key derived from: machine ID + user-provided password (optional)

type Encryptor interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
}

// Default: derives key from machine ID (no password required)
// Optional: user can set GOATWAY_ENCRYPTION_KEY for additional security
```

---

### Phase 2: Data Directory & Configuration ✅ COMPLETED

**Goal**: Establish local data directory structure

**Current State**: `internal/config/config.go` exists with basic env var loading for
`SERVER_ADDR`, `LLM_PROVIDER`, provider API keys, `LOG_LEVEL`.

**Tasks**:

- [x] Create `internal/config/paths.go` (~50 lines)
- [x] Update `internal/config/config.go` with new fields
- [ ] Add optional YAML config file support (deferred to future)

#### New File: `internal/config/paths.go` (~50 lines)

**Directory Structure:**

```
~/.goatway/                    # Linux/macOS
%APPDATA%\goatway\             # Windows

Contents:
├── goatway.db                 # SQLite database
├── config.yaml                # Optional config overrides
└── logs/                      # Optional file logs
    └── goatway.log
```

**Path Resolution:**

```go
func DataDir() string {
    // 1. Check GOATWAY_DATA_DIR env var
    // 2. Use XDG_DATA_HOME if set (Linux)
    // 3. Default to ~/.goatway or %APPDATA%\goatway
}

func DBPath() string {
    return filepath.Join(DataDir(), "goatway.db")
}

func ConfigPath() string {
    return filepath.Join(DataDir(), "config.yaml")
}
```

#### Update: `internal/config/config.go`

**Fields to Add** (existing fields: ServerAddr, Provider, *APIKey, LogLevel):

```go
type Config struct {
    // Existing fields...
    ServerAddr       string // ✅ exists
    Provider         string // ✅ exists
    OpenRouterAPIKey string // ✅ exists
    LogLevel         string // ✅ exists

    // NEW: Storage
    DataDir    string `yaml:"data_dir" env:"GOATWAY_DATA_DIR"`

    // NEW: Logging
    LogFormat  string `yaml:"log_format" env:"LOG_FORMAT"`

    // NEW: Security
    EncryptionKey string `yaml:"-" env:"GOATWAY_ENCRYPTION_KEY"`
    AdminPassword string `yaml:"-" env:"GOATWAY_ADMIN_PASSWORD"`

    // NEW: Features
    EnableWebUI bool `yaml:"enable_web_ui" env:"ENABLE_WEB_UI"`
}

// Load order: defaults → config.yaml → env vars
func Load() (*Config, error)
```

---

### Phase 3: Admin API for Credential Management ✅ COMPLETED

**Goal**: REST API for managing credentials

**Dependencies**: Phase 1 ✅, Phase 2 ✅

**Tasks**:

- [x] Create admin handler files (split into 3 files for maintainability)
- [x] Add admin routes to `internal/app/router.go`
- [x] Update `internal/transport/http/handler/repo.go` to include Storage

#### New File: `internal/transport/http/handler/admin.go` (~180 lines)

**Endpoints:**

```
Credentials:
  POST   /api/admin/credentials          → Create credential
  GET    /api/admin/credentials          → List all credentials
  GET    /api/admin/credentials/{id}     → Get credential (key masked)
  PUT    /api/admin/credentials/{id}     → Update credential
  DELETE /api/admin/credentials/{id}     → Delete credential
  POST   /api/admin/credentials/{id}/default → Set as default

Usage & Logs:
  GET    /api/admin/usage                → Get usage statistics
  GET    /api/admin/usage/daily          → Get daily breakdown
  GET    /api/admin/logs                 → Get request logs (paginated)
  DELETE /api/admin/logs                 → Clear old logs

System:
  GET    /api/admin/health               → Health check with DB status
  GET    /api/admin/info                 → Version, uptime, stats
```

**Request/Response Examples:**

```json
// POST /api/admin/credentials
Request:
{
    "provider": "openrouter",
    "name": "My OpenRouter Key",
    "api_key": "sk-or-v1-xxxxx",
    "is_default": true
}

Response:
{
    "id": "cred_abc123",
    "provider": "openrouter",
    "name": "My OpenRouter Key",
    "api_key_preview": "sk-or-v1-xxx...xxx",
    "is_default": true,
    "created_at": "2024-01-15T10:30:00Z"
}
```

```json
// GET /api/admin/usage?start_date=2024-01-01&end_date=2024-01-31
Response:
{
    "total_requests": 1542,
    "total_tokens": 523400,
    "prompt_tokens": 312000,
    "completion_tokens": 211400,
    "error_count": 12,
    "models": {
        "gpt-4": {
            "requests": 500,
            "tokens": 250000
        },
        "claude-3-opus": {
            "requests": 1042,
            "tokens": 273400
        }
    }
}
```

```json
// GET /api/admin/logs?limit=50&offset=0&model=gpt-4
Response:
{
    "logs": [
        {
            "id": "log_xyz",
            "request_id": "req_abc",
            "model": "gpt-4",
            "prompt_tokens": 150,
            "completion_tokens": 87,
            "total_tokens": 237,
            "duration_ms": 2340,
            "status_code": 200,
            "created_at": "2024-01-15T10:30:00Z"
        }
    ],
    "total": 1542,
    "limit": 50,
    "offset": 0
}
```

**Security:**
```go
// Optional admin authentication
func AdminAuthMiddleware(password string) func(http.Handler) http.Handler {
    // If GOATWAY_ADMIN_PASSWORD is set, require it
    // Check Authorization header or session cookie
    // Allow localhost without auth by default
}
```

---

### Phase 4: Embedded Web UI ⏳ PENDING

**Goal**: Browser-based dashboard for local management

**Dependencies**: Phase 3 (Admin API)

**Tasks**:

- [ ] Create `web/` directory structure
- [ ] Create `web/index.html` - SPA shell
- [ ] Create `web/static/css/styles.css` - Minimal CSS
- [ ] Create `web/static/js/app.js` - Vanilla JS app
- [ ] Create `internal/transport/http/handler/webui.go` - Static file serving
- [ ] Add Web UI routes to router

#### Directory: `web/` (Frontend Assets)

**Structure:**
```
web/
├── index.html              # Single page app shell
├── static/
│   ├── css/
│   │   └── styles.css      # Minimal CSS (no framework)
│   └── js/
│       └── app.js          # Vanilla JS application
└── templates/              # Go templates (optional)
```

**UI Components:**

1. **Dashboard** (`/`)
   - Today's usage summary (requests, tokens, cost estimate)
   - Quick stats cards
   - Recent requests list
   - Error rate indicator

2. **Credentials** (`/credentials`)
   - List of configured API keys (masked)
   - Add new credential form
   - Edit/Delete actions
   - Set default credential
   - Test connection button

3. **Usage Analytics** (`/usage`)
   - Date range selector
   - Token usage chart (daily)
   - Model breakdown pie chart
   - Request count timeline
   - Export to CSV

4. **Request Logs** (`/logs`)
   - Paginated log table
   - Filter by model, status, date
   - Search by request ID
   - Request detail modal

5. **Settings** (`/settings`)
   - Server configuration display
   - Data directory path
   - Clear logs action
   - Export/Import credentials

#### New File: `internal/transport/http/handler/webui.go` (~60 lines)

**Static File Serving:**
```go
//go:embed web/*
var webFS embed.FS

func (h *Repo) ServeWebUI() http.Handler {
    // Serve embedded static files
    // SPA fallback to index.html for client-side routing
}
```

**Route Registration:**
```go
// Web UI routes
mux.Handle("GET /", repo.ServeWebUI())
mux.Handle("GET /static/", repo.ServeWebUI())
mux.Handle("GET /credentials", repo.ServeWebUI())
mux.Handle("GET /usage", repo.ServeWebUI())
mux.Handle("GET /logs", repo.ServeWebUI())
mux.Handle("GET /settings", repo.ServeWebUI())
```

---

### Phase 5: Request/Response Type Definitions ✅ COMPLETED

**Goal**: OpenAI-compatible type definitions

**Dependencies**: None (can be done in parallel with Phase 2-4)

**Tasks**:

- [x] Create `internal/types/` directory
- [x] Create `internal/types/message.go` (~100 lines) - Message types with multimodal content
- [x] Create `internal/types/request.go` (~95 lines) - ChatCompletionRequest with all fields
- [x] Create `internal/types/response.go` (~80 lines) - ChatCompletionResponse, Choice, Usage
- [x] Create `internal/types/stream.go` (~60 lines) - ChatCompletionChunk for streaming
- [x] Create `internal/types/tools.go` (~90 lines) - Tool/function calling types
- [x] Create `internal/types/errors.go` (~90 lines) - OpenAI error format
- [x] Create `internal/types/json.go` (~20 lines) - JSON helper functions

#### New File: `internal/types/openai.go` (~120 lines)

**Request Types:**
```
ChatCompletionRequest
├── Model (string, required)
├── Messages []Message
│   ├── Role (string: system/user/assistant)
│   ├── Content (string or []ContentPart)
│   └── Name (string, optional)
├── Stream (bool)
├── StreamOptions *StreamOptions
│   └── IncludeUsage (bool)
├── Temperature (*float64)
├── TopP (*float64)
├── MaxTokens (*int)
├── Stop (string or []string)
├── PresencePenalty (*float64)
├── FrequencyPenalty (*float64)
└── User (string)
```

**Response Types (Non-Streaming):**
```
ChatCompletionResponse
├── ID (string)
├── Object (string: "chat.completion")
├── Created (int64)
├── Model (string)
├── Choices []Choice
│   ├── Index (int)
│   ├── Message Message
│   ├── FinishReason (string)
│   └── Logprobs (*Logprobs)
├── Usage Usage
│   ├── PromptTokens (int)
│   ├── CompletionTokens (int)
│   └── TotalTokens (int)
└── SystemFingerprint (string)
```

**Response Types (Streaming):**
```
ChatCompletionChunk
├── ID (string)
├── Object (string: "chat.completion.chunk")
├── Created (int64)
├── Model (string)
├── Choices []ChunkChoice
│   ├── Index (int)
│   ├── Delta Delta
│   └── FinishReason (*string)
└── Usage *Usage (final chunk only)
```

#### New File: `internal/types/errors.go` (~50 lines)

**OpenAI Error Format:**
```go
type APIError struct {
    Error struct {
        Message string  `json:"message"`
        Type    string  `json:"type"`
        Param   *string `json:"param,omitempty"`
        Code    *string `json:"code,omitempty"`
    } `json:"error"`
}
```

---

### Phase 6: Token Counting Module ⏳ PENDING

**Goal**: Accurate token counting with local persistence

**Dependencies**: Phase 5 (types for Message struct)

**Tasks**:

- [ ] Create `internal/tokenizer/` directory
- [ ] Create `internal/tokenizer/tokenizer.go` (~80 lines)
- [ ] Create `internal/tokenizer/counter.go` (~100 lines)
- [ ] Add `github.com/pkoukk/tiktoken-go` dependency

#### New File: `internal/tokenizer/tokenizer.go` (~80 lines)

**Interface:**
```go
type Tokenizer interface {
    CountTokens(text string, model string) (int, error)
    CountMessages(messages []types.Message, model string) (int, error)
}
```

**Implementation:**
- Use `github.com/pkoukk/tiktoken-go` for accuracy
- Model-specific encoding selection (cl100k_base for GPT-4, etc.)
- Fallback to estimation for unknown models

#### New File: `internal/tokenizer/counter.go` (~100 lines)

**Message Counting Logic:**
```go
func (c *Counter) CountRequest(req *types.ChatCompletionRequest) (int, error) {
    encoding := c.getEncoding(req.Model)
    total := 0

    for _, msg := range req.Messages {
        // Per-message overhead varies by model
        total += c.messageOverhead(req.Model)
        total += encoding.CountTokens(msg.Content)
        if msg.Name != "" {
            total += encoding.CountTokens(msg.Name) + 1
        }
    }

    total += c.assistantPriming(req.Model)
    return total, nil
}
```

---

### Phase 7: Enhanced Proxy with Logging ⏳ PENDING

**Goal**: Integrate storage and token counting into proxy flow

**Dependencies**: Phase 1 ✅, Phase 2, Phase 5, Phase 6

**Current State**: `proxy.go` exists with basic provider-agnostic proxy.
`repo.go` has Cache and Provider, needs Storage and Tokenizer.

**Tasks**:

- [ ] Update `internal/transport/http/handler/repo.go` - Add Storage, Tokenizer
- [ ] Update `internal/transport/http/handler/proxy.go` - Add logging flow
- [ ] Create `internal/provider/stream.go` (~120 lines) - SSE parser
- [ ] Update `cmd/api/main.go` - Wire storage into repo

#### Update: `internal/transport/http/handler/proxy.go` (~180 lines)

**Enhanced Flow:**
```
1. Parse request body → extract model
2. Resolve credential (from header or default)
3. Count prompt tokens
4. Start timer
5. Proxy to upstream provider
6. For streaming: accumulate content, count completion tokens
7. For non-streaming: parse response, extract usage
8. Log to SQLite (async)
9. Return response to client
```

**Credential Resolution:**
```go
func (h *Repo) resolveCredential(r *http.Request) (*storage.Credential, error) {
    // 1. Check Authorization header for credential ID hint
    //    Format: "Bearer goatway:cred_xxx:actual_key" or just "Bearer key"

    // 2. Check X-Goatway-Credential header for credential ID

    // 3. Fall back to default credential for provider

    // 4. Error if no credential found
}
```

#### New File: `internal/provider/stream.go` (~120 lines)

**SSE Parser with Accumulation:**
```go
type StreamProcessor struct {
    model           string
    contentBuffer   strings.Builder
    finishReason    string
    promptTokens    int
}

func (p *StreamProcessor) ProcessChunk(data []byte) error {
    // Parse SSE data line
    // Extract delta.content, accumulate
    // Track finish_reason
}

func (p *StreamProcessor) GetCompletionText() string {
    return p.contentBuffer.String()
}
```

---

### Phase 8: Provider Interface Enhancement ✅ COMPLETED

**Goal**: Support credential injection and logging hooks

**Dependencies**: Phase 1 ✅, Phase 7 ✅

**Current State**: Provider interface enhanced with `ProxyOptions` and `ProxyResult` types.
Credential injection and result tracking fully implemented.

**Tasks**:

- [x] Update `internal/provider/provider.go` - Add ProxyOptions, ProxyResult types
- [x] Update `internal/provider/openrouter.go` - Credential injection, result tracking
- [x] Update any other provider implementations (N/A - only OpenRouter currently)

#### Update: `internal/provider/provider.go`

**Enhanced Interface:**
```go
type Provider interface {
    Name() string
    BaseURL() string

    // Proxy with full context
    ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *ProxyOptions) (*ProxyResult, error)
}

type ProxyOptions struct {
    Credential   *storage.Credential
    RequestID    string
    PromptTokens int
    OnComplete   func(result *ProxyResult)  // Async callback for logging
}

type ProxyResult struct {
    Model            string
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    FinishReason     string
    StatusCode       int
    Duration         time.Duration
    Error            error
}
```

#### Update: `internal/provider/openrouter.go`

**Credential Injection:**
```go
func (p *OpenRouterProvider) ProxyRequest(..., opts *ProxyOptions) (*ProxyResult, error) {
    // Use credential from opts instead of env var
    req.Header.Set("Authorization", "Bearer "+opts.Credential.APIKey)

    // Rest of proxy logic...

    // Call OnComplete callback for async logging
    if opts.OnComplete != nil {
        go opts.OnComplete(result)
    }
}
```

---

### Phase 9: Additional OpenAI Endpoints ✅ COMPLETED

**Goal**: Full API compatibility

**Dependencies**: Phase 5 (types) ✅

**Tasks**:

- [x] Create `internal/transport/http/handler/models.go` (~120 lines)
- [x] Add `/v1/models` routes to router

#### New File: `internal/transport/http/handler/models.go` (~120 lines)

**Endpoints:**
```go
// GET /v1/models - List available models (proxies to OpenRouter)
func (h *Repo) ListModels(w http.ResponseWriter, r *http.Request)

// GET /v1/models/{model} - Get model details (filters from OpenRouter list)
func (h *Repo) GetModel(w http.ResponseWriter, r *http.Request)
```

---

### Phase 10: Request ID & Middleware ✅ COMPLETED

**Goal**: Request tracing and common middleware

**Dependencies**: None (can be done in parallel)

**Tasks**:

- [x] Create `internal/transport/http/middleware/` directory
- [x] Create `internal/transport/http/middleware/middleware.go` (~130 lines)
- [x] Create `internal/transport/http/middleware/middleware_test.go` (~200 lines)
- [x] Wire middleware into router/server
- [x] Add AdminPassword to config.go

#### New File: `internal/transport/http/middleware/middleware.go` (~130 lines)

**Middleware Stack:**
```go
// Request ID generation - adds unique ID to each request
func RequestID(next http.Handler) http.Handler

// Request logging (to stdout via slog)
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler

// CORS for Web UI - enables cross-origin requests
func CORS(next http.Handler) http.Handler

// Admin auth (optional) - protects admin routes when password is set
func AdminAuth(password string) func(http.Handler) http.Handler

// Helper to get request ID from context
func GetRequestID(ctx context.Context) string
```

---

### Phase 11: CLI & Distribution ✅ COMPLETED

**Goal**: Easy installation and startup

**Dependencies**: All previous phases

**Tasks**:

- [x] Update `cmd/api/main.go` - Add CLI flags, banner, full initialization
- [x] Update `Makefile` - Add cross-platform build targets
- [x] Create `install.sh` - One-line install script
- [x] Test on macOS, Linux, Windows (cross-compilation verified)

#### Update: `cmd/api/main.go`

**CLI Interface:**
```go
func main() {
    // Parse flags
    flag.StringVar(&addr, "addr", ":8080", "Server address")
    flag.StringVar(&dataDir, "data-dir", "", "Data directory path")
    flag.BoolVar(&version, "version", false, "Print version")
    flag.Parse()

    // Print banner
    fmt.Println("Goatway - Local OpenAI-Compatible Proxy")
    fmt.Println("========================================")

    // Initialize storage
    storage := sqlite.New(config.DBPath())
    storage.Migrate()

    // Wire dependencies
    // Start server

    fmt.Printf("Web UI:     http://localhost%s\n", addr)
    fmt.Printf("Proxy API:  http://localhost%s/v1/chat/completions\n", addr)
    fmt.Printf("Data:       %s\n", config.DataDir())
}
```

#### New File: `Makefile` updates

**Distribution Targets:**
```makefile
# Build for all platforms
build-all:
    GOOS=darwin GOARCH=amd64 go build -o dist/goatway-darwin-amd64 ./cmd/api
    GOOS=darwin GOARCH=arm64 go build -o dist/goatway-darwin-arm64 ./cmd/api
    GOOS=linux GOARCH=amd64 go build -o dist/goatway-linux-amd64 ./cmd/api
    GOOS=windows GOARCH=amd64 go build -o dist/goatway-windows-amd64.exe ./cmd/api

# Install locally
install:
    go install ./cmd/api

# Create release archives
release: build-all
    # Create tarballs and zips
```

#### New File: `install.sh` (~30 lines)

**One-Line Install Script:**
```bash
#!/bin/bash
# curl -fsSL https://raw.githubusercontent.com/mandalnilabja/goatway/main/install.sh | bash

set -e
REPO="mandalnilabja/goatway"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Download and install
curl -fsSL "https://github.com/$REPO/releases/latest/download/goatway-$OS-$ARCH" -o /tmp/goatway
chmod +x /tmp/goatway
sudo mv /tmp/goatway "$INSTALL_DIR/goatway"

echo "Goatway installed! Run: goatway"
```

---

## File Structure Summary

### Completed Files (Phases 1-11) ✅

| File | Lines | Phase | Purpose |
|------|-------|-------|---------|
| `internal/storage/sqlite.go` | ~660 | 1 | SQLite storage implementation |
| `internal/storage/models.go` | ~110 | 1 | Data models for storage |
| `internal/storage/storage.go` | ~45 | 1 | Storage interface |
| `internal/storage/encryption.go` | ~100 | 1 | API key encryption |
| `internal/storage/sqlite_test.go` | ~320 | 1 | Storage tests |
| `internal/storage/encryption_test.go` | ~100 | 1 | Encryption tests |
| `internal/config/paths.go` | ~55 | 2 | Data directory resolution |
| `internal/transport/http/handler/admin_credentials.go` | ~175 | 3 | Credential CRUD endpoints |
| `internal/transport/http/handler/admin_usage.go` | ~145 | 3 | Usage and logs endpoints |
| `internal/transport/http/handler/admin_system.go` | ~70 | 3 | Health and info endpoints |
| `web/index.html` | ~35 | 4 | Web UI shell |
| `web/static/css/styles.css` | ~350 | 4 | Web UI styles |
| `web/static/js/api.js` | ~95 | 4 | API helper functions |
| `web/static/js/pages.js` | ~195 | 4 | Page rendering functions |
| `web/static/js/app.js` | ~195 | 4 | Router, charts, modals |
| `web/embed.go` | ~10 | 4 | Go embed directive |
| `internal/transport/http/handler/webui.go` | ~50 | 4 | Web UI serving |
| `internal/types/message.go` | ~100 | 5 | Message types with multimodal content |
| `internal/types/request.go` | ~95 | 5 | ChatCompletionRequest with all fields |
| `internal/types/response.go` | ~80 | 5 | ChatCompletionResponse, Choice, Usage |
| `internal/types/stream.go` | ~60 | 5 | ChatCompletionChunk for streaming |
| `internal/types/tools.go` | ~90 | 5 | Tool/function calling types |
| `internal/types/errors.go` | ~90 | 5 | OpenAI error format |
| `internal/types/json.go` | ~20 | 5 | JSON helper functions |
| `internal/tokenizer/tokenizer.go` | ~110 | 6 | Tokenizer interface and encoding |
| `internal/tokenizer/counter.go` | ~200 | 6 | Token counting for messages/tools |
| `internal/tokenizer/tokenizer_test.go` | ~260 | 6 | Tokenizer tests |
| `internal/provider/stream.go` | ~110 | 7 | SSE stream processor |
| `internal/transport/http/handler/models.go` | ~120 | 9 | /v1/models endpoints |
| `internal/transport/http/middleware/middleware.go` | ~130 | 10 | HTTP middleware (RequestID, CORS, Logger, AdminAuth) |
| `internal/transport/http/middleware/middleware_test.go` | ~200 | 10 | Middleware tests |
| `install.sh` | ~100 | 11 | One-line installation script |

### Files Modified (Phases 1-11) ✅

| File | Phase | Changes |
|------|-------|---------|
| `internal/config/config.go` | 2 | Added DataDir, EncryptionKey, EnableWebUI |
| `internal/transport/http/handler/repo.go` | 3, 7 | Added Storage, StartTime, Tokenizer |
| `internal/app/router.go` | 3, 4, 9 | Added admin routes, Web UI routes, /v1/models |
| `cmd/api/main.go` | 3, 4, 7, 11 | Wired storage, router options, tokenizer, CLI flags |
| `internal/transport/http/handler/proxy.go` | 7 | Integrated storage, token counting, logging |
| `internal/provider/provider.go` | 7, 8 | Added ProxyOptions, ProxyResult types |
| `internal/provider/openrouter.go` | 7, 8 | Credential injection, result tracking |
| `internal/config/config.go` | 10 | Added AdminPassword |
| `internal/app/router.go` | 10 | Added middleware chain, admin auth |
| `cmd/api/main.go` | 10 | Added logger setup, admin password config |
| `Makefile` | 11 | Added ldflags, build-all, install, release |

---

## Dependencies

### Dependencies Added (Phases 1-6) ✅

| Package                          | Purpose          | Phase | Status   |
|----------------------------------|------------------|-------|----------|
| `modernc.org/sqlite`             | SQLite (pure Go) | 1     | ✅ Added |
| `github.com/google/uuid`         | UUID generation  | 1     | ✅ Added |
| `github.com/pkoukk/tiktoken-go`  | Token counting   | 6     | ✅ Added |

**Note**: Used `modernc.org/sqlite` (pure Go) instead of CGO-based driver for easier cross-compilation.
YAML config support (`gopkg.in/yaml.v3`) was deferred to a future release.

---

## Data Flow Diagrams

### Credential Management Flow
```
User → Web UI → Admin API → Storage (SQLite)
                    ↓
              Encrypt API Key
                    ↓
              Store in credentials table
```

### Request Proxy Flow
```
Client Request
      ↓
Parse Request Body → Extract Model
      ↓
Resolve Credential → From header or default
      ↓
Count Prompt Tokens
      ↓
Proxy to OpenRouter → Inject Credential
      ↓
Stream Response → Accumulate content
      ↓
Count Completion Tokens
      ↓
Log to SQLite (async) → request_logs + usage_daily
      ↓
Return Response
```

### Usage Analytics Flow
```
Web UI → Admin API → Storage Query
                          ↓
              Aggregate from usage_daily table
                          ↓
              Return JSON → Render Charts
```

---

## Web UI Design

### Technology Choices
- **No framework**: Vanilla HTML/CSS/JS
- **No build step**: Direct browser loading
- **Embedded in binary**: `//go:embed` directive
- **Responsive**: Works on desktop and mobile

### Pages

**1. Dashboard (`/`)**
```
┌─────────────────────────────────────────────────┐
│  GOATWAY                          [Settings] ⚙  │
├─────────────────────────────────────────────────┤
│                                                  │
│  Today's Usage                                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│  │ Requests │ │  Tokens  │ │  Errors  │        │
│  │   127    │ │  45,230  │ │    2     │        │
│  └──────────┘ └──────────┘ └──────────┘        │
│                                                  │
│  Recent Requests                                 │
│  ┌─────────────────────────────────────────┐   │
│  │ Model      │ Tokens │ Duration │ Status │   │
│  ├─────────────────────────────────────────┤   │
│  │ gpt-4      │  237   │  2.3s    │  200   │   │
│  │ claude-3   │  512   │  1.8s    │  200   │   │
│  │ gpt-4      │  189   │  1.5s    │  200   │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  [Credentials] [Usage] [Logs]                   │
└─────────────────────────────────────────────────┘
```

**2. Credentials (`/credentials`)**
```
┌─────────────────────────────────────────────────┐
│  API Credentials                   [+ Add New]  │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌─────────────────────────────────────────┐   │
│  │ ★ My OpenRouter Key                      │   │
│  │   Provider: openrouter                   │   │
│  │   Key: sk-or-v1-xxx...xxx               │   │
│  │   [Test] [Edit] [Delete]                │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  ┌─────────────────────────────────────────┐   │
│  │   Backup OpenRouter                      │   │
│  │   Provider: openrouter                   │   │
│  │   Key: sk-or-v1-yyy...yyy               │   │
│  │   [Set Default] [Edit] [Delete]         │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
└─────────────────────────────────────────────────┘
```

**3. Usage (`/usage`)**
```
┌─────────────────────────────────────────────────┐
│  Usage Analytics        [Last 7 days ▼] [Export]│
├─────────────────────────────────────────────────┤
│                                                  │
│  Token Usage Over Time                          │
│  ┌─────────────────────────────────────────┐   │
│  │     ▄                                    │   │
│  │    ▄█▄    ▄                             │   │
│  │   ▄███▄  ▄█▄  ▄                         │   │
│  │  ▄█████▄▄███▄▄█▄                        │   │
│  │ ▄███████████████▄                       │   │
│  └─────────────────────────────────────────┘   │
│   Mon  Tue  Wed  Thu  Fri  Sat  Sun            │
│                                                  │
│  Model Breakdown                                │
│  ┌────────────────┐  gpt-4:      45%           │
│  │    ████        │  claude-3:   35%           │
│  │  ████████      │  gpt-3.5:    20%           │
│  │    ████        │                             │
│  └────────────────┘                             │
│                                                  │
└─────────────────────────────────────────────────┘
```

---

## Security Considerations

### API Key Protection
1. Keys encrypted at rest with AES-256-GCM
2. Machine-specific key derivation (no plaintext storage)
3. Optional user password for additional security
4. Keys never logged or exposed in responses

### Admin API Security
1. Localhost-only by default (no auth required)
2. Optional password for remote access
3. API key preview (masked) in responses
4. Rate limiting on auth attempts

### Database Security
1. SQLite file permissions (0600)
2. No sensitive data in request logs
3. Automatic log rotation/cleanup options

---

## Critical Streaming Constraints

Per AGENTS.md, these rules MUST be maintained:

1. **`http.Transport.DisableCompression = true`**
2. **Immediate `Flusher.Flush()` after each chunk**
3. **No full response buffering** - parse-as-you-stream
4. **Context propagation** - client cancellation stops upstream
5. **32KB buffer** for stream pump

### Token Counting During Streaming
- Accumulate `delta.content` in memory (typically <100KB)
- Count tokens after stream completes
- Log asynchronously to avoid blocking response

---

## Implementation Order

| Phase | Name               | Status       | Dependencies      |
|-------|--------------------|--------------| ------------------|
| 1     | Storage Layer      | ✅ COMPLETED | None              |
| 2     | Data Directory     | ✅ COMPLETED | Phase 1           |
| 3     | Admin API          | ✅ COMPLETED | Phase 1, 2        |
| 4     | Web UI             | ✅ COMPLETED | Phase 3           |
| 5     | Type Definitions   | ✅ COMPLETED | None (parallel)   |
| 6     | Token Counting     | ✅ COMPLETED | Phase 5           |
| 7     | Enhanced Proxy     | ✅ COMPLETED | Phase 1, 2, 5, 6  |
| 8     | Provider Updates   | ✅ COMPLETED | Phase 1, 7        |
| 9     | OpenAI Endpoints   | ✅ COMPLETED | Phase 5           |
| 10    | Middleware         | ✅ COMPLETED | None (parallel)   |
| 11    | CLI & Distribution | ✅ COMPLETED | All               |

**Recommended Parallel Tracks:**

```text
Track A (Core):     Phase 2 → Phase 3 → Phase 4
Track B (Types):    Phase 5 → Phase 6 → Phase 7 → Phase 8
Track C (Extras):   Phase 9, Phase 10 (can be done anytime)
Track D (Final):    Phase 11 (after all others)
```

**Recommended Commit Sequence:**
1. `feat(storage): add SQLite storage layer with encryption`
2. `feat(config): add data directory and YAML config support`
3. `feat(admin): add credential management API`
4. `feat(webui): add embedded web dashboard`
5. `feat(types): add OpenAI request/response types`
6. `feat(tokenizer): add token counting module`
7. `feat(proxy): integrate storage and token counting`
8. `feat(provider): add credential injection support`
9. `feat(handler): add /v1/models endpoints`
10. `feat(middleware): add request ID and logging`
11. `feat(cli): add flags and installation script`
12. `chore(release): add cross-platform build targets`

---

## Success Criteria

1. **Local-First Experience**
   - Single binary, no external dependencies
   - Works offline after initial setup
   - All data stored locally in `~/.goatway/`

2. **Easy Installation**
   - One command install: `curl ... | bash` or `go install`
   - Works on macOS, Linux, Windows
   - No configuration required to start

3. **Web UI Functionality**
   - Add/edit/delete API credentials
   - View usage statistics and charts
   - Browse request logs
   - All via browser at `http://localhost:8080`

4. **Full OpenAI Compatibility**
   - Drop-in replacement for OpenAI API
   - Streaming works with all clients
   - Accurate token counting

5. **Persistent Logging**
   - All requests logged to SQLite
   - Token usage tracked per model
   - Historical data queryable via API

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| SQLite performance at scale | Medium | Add indexes, log rotation, archival |
| CGO dependency for SQLite | Medium | Use pure-Go `modernc.org/sqlite` |
| Web UI complexity | Low | Keep minimal, vanilla JS only |
| Encryption key management | Medium | Document clearly, support password override |
| Cross-platform paths | Low | Use `filepath` package, test on all OS |
| Binary size with embedded UI | Low | Minimize assets, compress if needed |

---

## Implementation Progress

### Phase 1: SQLite Storage Layer ✅ COMPLETED

**Date**: 2026-02-07

**Files Created:**

- `internal/storage/models.go` (~110 lines) - Data models for credentials, request logs, usage stats
- `internal/storage/storage.go` (~45 lines) - Storage interface definition
- `internal/storage/encryption.go` (~100 lines) - AES-256-GCM encryption for API keys
- `internal/storage/sqlite.go` (~660 lines) - Full SQLite implementation
- `internal/storage/sqlite_test.go` (~320 lines) - Comprehensive test suite
- `internal/storage/encryption_test.go` (~100 lines) - Encryption tests

**Dependencies Added:**

- `modernc.org/sqlite` - Pure Go SQLite driver (no CGO required)
- `github.com/google/uuid` - UUID generation for IDs

**Features Implemented:**

- [x] Credential CRUD operations with encryption at rest
- [x] Default credential management per provider
- [x] Request logging with filtering
- [x] Daily usage aggregation with upsert support
- [x] Usage statistics with model breakdown
- [x] API key masking for safe display
- [x] Thread-safe operations with RWMutex
- [x] WAL mode for better concurrency
- [x] Machine-specific key derivation (or env var override)

**Test Coverage:**

- All 14 tests passing
- Covers CRUD operations, encryption, edge cases
- Tests for storage closed state

**Notes:**

- Used `modernc.org/sqlite` (pure Go) instead of CGO-based driver for easier cross-compilation
- Encryption uses AES-256-GCM with machine-derived key or `GOATWAY_ENCRYPTION_KEY` env var
- Daily usage uses empty string instead of NULL for credential_id to support ON CONFLICT upserts

### Phase 2: Data Directory & Configuration ✅ COMPLETED

**Date**: 2026-02-07

**Files Created:**

- `internal/config/paths.go` (~55 lines) - Data directory resolution with cross-platform support

**Files Modified:**

- `internal/config/config.go` - Added DataDir, EncryptionKey, LogFormat, EnableWebUI fields

**Features Implemented:**

- [x] Cross-platform data directory resolution (Linux XDG, Windows APPDATA, macOS)
- [x] GOATWAY_DATA_DIR environment variable override
- [x] Helper functions: DataDir(), DBPath(), ConfigPath(), LogDir(), EnsureDataDir()
- [x] New config fields for storage, logging, and feature flags
- [x] Boolean environment variable parsing (getEnvBool)

**Notes:**

- YAML config file support deferred to future phase (env vars sufficient for now)
- Data directory defaults to `~/.goatway/` on Linux/macOS, `%APPDATA%\goatway` on Windows

### Phase 3: Admin API for Credential Management ✅ COMPLETED

**Date**: 2026-02-07

**Files Created:**

- `internal/transport/http/handler/admin_credentials.go` (~175 lines) - Credential CRUD endpoints
- `internal/transport/http/handler/admin_usage.go` (~145 lines) - Usage and logs endpoints
- `internal/transport/http/handler/admin_system.go` (~70 lines) - Health and info endpoints

**Files Modified:**

- `internal/transport/http/handler/repo.go` - Added Storage dependency and StartTime
- `internal/app/router.go` - Added admin route registration
- `cmd/api/main.go` - Wired storage initialization, added startup banner

**Endpoints Implemented:**

Credentials:
- `POST   /api/admin/credentials` - Create credential
- `GET    /api/admin/credentials` - List all credentials (masked)
- `GET    /api/admin/credentials/{id}` - Get credential (masked)
- `PUT    /api/admin/credentials/{id}` - Update credential
- `DELETE /api/admin/credentials/{id}` - Delete credential
- `POST   /api/admin/credentials/{id}/default` - Set as default

Usage & Logs:
- `GET    /api/admin/usage` - Get usage statistics with filtering
- `GET    /api/admin/usage/daily` - Get daily breakdown
- `GET    /api/admin/logs` - Get request logs (paginated)
- `DELETE /api/admin/logs?before_date=YYYY-MM-DD` - Clear old logs

System:
- `GET    /api/admin/health` - Health check with DB status
- `GET    /api/admin/info` - Version, uptime, stats

**Notes:**

- Admin handlers split into 3 files to stay under 200-line guideline
- Authentication middleware deferred (localhost access open by default)
- DELETE /api/admin/logs requires `before_date` parameter for safety
- All credential responses use masked API keys (CredentialPreview)

### Phase 4: Embedded Web UI ✅ COMPLETED

**Date**: 2026-02-07

**Files Created:**

- `web/index.html` (~35 lines) - SPA shell with navigation and Chart.js CDN
- `web/static/css/styles.css` (~350 lines) - Minimal CSS with cards, tables, forms, modals
- `web/static/js/api.js` (~95 lines) - API helper functions for admin endpoints
- `web/static/js/pages.js` (~195 lines) - Page rendering functions for all views
- `web/static/js/app.js` (~195 lines) - Router, charts, modals, and actions
- `web/embed.go` (~10 lines) - Go embed directive for web assets
- `internal/transport/http/handler/webui.go` (~50 lines) - Static file serving with SPA fallback

**Files Modified:**

- `internal/app/router.go` - Added RouterOptions, registerWebUIRoutes function
- `cmd/api/main.go` - Pass EnableWebUI option to router, updated startup banner

**Features Implemented:**

- [x] Dashboard page with today's stats and recent requests
- [x] Credentials page with CRUD operations (add/edit/delete)
- [x] Usage analytics page with Chart.js charts (bar chart, doughnut chart)
- [x] Request logs page with pagination
- [x] Settings page with system info and log cleanup
- [x] History API routing for clean URLs
- [x] SPA fallback (all routes return index.html)
- [x] Embedded assets via go:embed
- [x] ENABLE_WEB_UI config flag (default: true)

**UI Pages:**

- `/` - Dashboard with usage stats and recent requests
- `/credentials` - API key management (add, edit, delete, set default)
- `/usage` - Analytics with token usage charts and model breakdown
- `/logs` - Paginated request logs with status badges
- `/settings` - System info and danger zone (clear old logs)

**Technical Decisions:**

- **Chart.js** from CDN for usage visualizations
- **History API routing** with server-side fallback to index.html
- **Light mode only** for simpler CSS (no dark mode)
- **Vanilla JS** - no frameworks, minimal dependencies
- **JavaScript split into 3 files** to stay under 200-line guideline

**Notes:**

- Web UI enabled by default, disable with `ENABLE_WEB_UI=false`
- Static files embedded in binary for single-file distribution
- All API calls use the admin endpoints implemented in Phase 3

### Phase 5: Request/Response Type Definitions ✅ COMPLETED

**Date**: 2026-02-07

**Files Created:**

- `internal/types/message.go` (~100 lines) - Message types with multimodal content support
- `internal/types/request.go` (~95 lines) - ChatCompletionRequest with all OpenAI fields
- `internal/types/response.go` (~80 lines) - ChatCompletionResponse, Choice, Usage, Logprobs
- `internal/types/stream.go` (~60 lines) - ChatCompletionChunk for streaming responses
- `internal/types/tools.go` (~90 lines) - Tool/function calling types
- `internal/types/errors.go` (~90 lines) - OpenAI error format with helper functions
- `internal/types/json.go` (~20 lines) - JSON marshaling helper functions

**Features Implemented:**

- [x] Full OpenAI Chat Completions API type coverage
- [x] Multimodal content support (text + image_url)
- [x] Custom JSON marshaling for polymorphic fields (Content, Stop, ToolChoice)
- [x] Tool/function calling types (Tool, ToolCall, ToolChoice)
- [x] Streaming chunk types with delta content
- [x] Response format types (JSON mode, structured outputs)
- [x] Logprobs support (token-level log probabilities)
- [x] Usage statistics with token details breakdown
- [x] OpenAI-compatible error response format
- [x] Helper constructors (NewTextMessage, NewImageMessage, NewTool, etc.)

**Type Coverage:**

Request types:
- `ChatCompletionRequest` - All fields including tools, response_format, seed, logprobs
- `Message` - Polymorphic content (string or array of ContentPart)
- `ContentPart` - Text and image_url support
- `StreamOptions` - Include usage in streaming
- `ResponseFormat` - JSON mode and structured outputs
- `Stop` - String or array of strings

Response types:
- `ChatCompletionResponse` - Non-streaming response
- `ChatCompletionChunk` - Streaming chunk
- `Choice` / `ChunkChoice` - Completion choices
- `Delta` - Incremental content in streaming
- `Usage` - Token usage with details breakdown
- `ChoiceLogprobs` / `TokenLogprob` - Log probability info

Tool types:
- `Tool` - Function tool definition
- `Function` - Function name, description, parameters
- `ToolCall` - Model's call to a tool
- `ToolChoice` - none/auto/required or specific function

Error types:
- `APIError` - OpenAI error response format
- Helper functions for common error types

**Technical Decisions:**

- **Split into 7 files** to stay under 200-line guideline per file
- **Custom JSON marshaling** for polymorphic fields (Content, Stop, ToolChoice)
- **Pointer fields** for optional values to distinguish unset from zero
- **Constants defined** for roles, content types, finish reasons, error types
- **Helper constructors** for common message and tool creation patterns

### Phase 6: Token Counting Module ✅ COMPLETED

**Date**: 2026-02-07

**Files Created:**

- `internal/tokenizer/tokenizer.go` (~110 lines) - Tokenizer interface and encoding management
- `internal/tokenizer/counter.go` (~200 lines) - Token counting for messages, images, and tools
- `internal/tokenizer/tokenizer_test.go` (~260 lines) - Comprehensive test suite

**Dependencies Added:**

- `github.com/pkoukk/tiktoken-go` - Accurate token counting using OpenAI's tiktoken

**Features Implemented:**

- [x] Tokenizer interface with CountTokens, CountMessages, CountRequest methods
- [x] Model-to-encoding resolution (cl100k_base for GPT-4/3.5, o200k_base for GPT-4o/o1)
- [x] Fallback to cl100k_base for unknown models (Claude, Mistral, etc.)
- [x] Thread-safe encoding caching with RWMutex
- [x] Message token counting with per-message overhead
- [x] Multimodal content support with image token calculation
- [x] Image token counting using OpenAI's rules (base + tiles based on detail)
- [x] Tool definition token counting (name, description, parameters schema)
- [x] Tool call token counting (ID, function name, arguments)
- [x] Reply priming overhead included in message counts

**Token Counting Details:**

Model encodings:

- `gpt-4o`, `o1`, `o3`, `chatgpt` → o200k_base
- `gpt-4`, `gpt-3.5`, `text-embedding` → cl100k_base
- All other models → cl100k_base (default)

Image token calculation:

- Base cost: 85 tokens per image
- Low detail: 85 + 170 = 255 tokens
- High detail: 85 + (4 × 170) = 765 tokens (simplified estimate)

Message overhead:

- GPT-4 family: 3 tokens per message
- GPT-3.5 family: 4 tokens per message
- Reply priming: 3 tokens

**Technical Decisions:**

- **Ordered prefix matching** - Longer prefixes checked first to avoid "gpt-4" matching "gpt-4o"
- **Encoding caching** - Tiktoken encodings cached per encoding name, not per model
- **Simplified image tokens** - Without actual image dimensions, uses reasonable estimates
- **Tool overhead constants** - Approximate JSON structure overhead for tool definitions

### Phase 7 & 8: Enhanced Proxy with Logging & Provider Updates ✅ COMPLETED

**Date**: 2026-02-07

**Note**: Phases 7 and 8 were combined into a single implementation as they are tightly coupled.

**Files Created:**

- `internal/provider/stream.go` (~110 lines) - SSE stream processor for parsing chunks

**Files Modified:**

- `internal/provider/provider.go` - Added ProxyOptions and ProxyResult types to interface
- `internal/provider/openrouter.go` (~225 lines) - Full rewrite with credential injection and result tracking
- `internal/transport/http/handler/repo.go` - Added Tokenizer dependency
- `internal/transport/http/handler/proxy.go` (~155 lines) - Complete rewrite with logging flow
- `cmd/api/main.go` - Wired tokenizer initialization

**Provider Interface Changes:**

```go
type ProxyOptions struct {
    APIKey       string     // From credential or header
    RequestID    string     // For tracing
    PromptTokens int        // Pre-calculated by handler
    Model        string     // From parsed request
    IsStreaming  bool       // Streaming request flag
    Body         io.Reader  // Buffered request body
}

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

**Features Implemented:**

- [x] Provider interface enhanced with ProxyOptions/ProxyResult for result tracking
- [x] SSE stream processor that parses chunks while forwarding to client
- [x] Upstream usage extraction from streaming responses (when available)
- [x] Non-streaming response parsing for usage extraction
- [x] Credential resolution: Bearer token from Authorization header or default from DB
- [x] Request body buffering for token counting before proxy
- [x] Prompt token counting using tiktoken before proxy
- [x] Asynchronous request logging to SQLite
- [x] Daily usage aggregate updates
- [x] OpenAI-compatible error responses

**Request Flow:**

```text
1. Parse request body → extract model, messages
2. Resolve API key (Authorization header or default credential)
3. Count prompt tokens using tokenizer
4. Build ProxyOptions with all context
5. Proxy to upstream provider
6. For streaming: use StreamProcessor to extract usage/content
7. For non-streaming: parse JSON response for usage
8. Return ProxyResult with all metadata
9. Log to SQLite asynchronously (request_logs + usage_daily)
```

**Token Counting Strategy:**

- Prefer upstream usage when available (more accurate for billing)
- Fall back to local tiktoken counting when upstream doesn't provide usage
- Streaming responses: extract from final chunk if `stream_options.include_usage=true`
- Non-streaming responses: parse from response JSON

**Technical Decisions:**

- **Combined Phases 7+8** - Provider interface changes needed for proper logging integration
- **Request body buffering** - Required to both parse for token counting and forward to upstream
- **Async logging** - Uses goroutine to avoid blocking response to client
- **Credential passthrough** - Authorization header takes precedence over stored credentials
- **Stream processor** - Parses SSE lines while forwarding, accumulates content for backup token counting

### Phase 9: Additional OpenAI Endpoints ✅ COMPLETED

**Date**: 2026-02-07

**Files Created:**

- `internal/transport/http/handler/models.go` (~120 lines) - Models listing endpoint

**Files Modified:**

- `internal/app/router.go` - Added /v1/models and /v1/models/{model} routes

**Endpoints Implemented:**

- `GET /v1/models` - Proxies to OpenRouter's models endpoint
- `GET /v1/models/{model}` - Fetches all models and filters for specific model ID

**Features Implemented:**

- [x] Proxy to OpenRouter's /v1/models endpoint for real-time model list
- [x] Model detail endpoint that filters from the full list (OpenRouter has no single-model endpoint)
- [x] Credential resolution reusing existing `resolveAPIKey` method from proxy.go
- [x] OpenAI-compatible error responses using types package
- [x] Proper header forwarding from upstream response

**Technical Decisions:**

- **Proxy approach** - Forwards to OpenRouter's model list rather than maintaining a static list
- **Single-model fallback** - OpenRouter doesn't have a GET /models/{id} endpoint, so we fetch all and filter
- **Credential reuse** - Uses the same `resolveAPIKey` method as the proxy handler for consistency
- **Error handling** - Uses the existing `types.WriteError` helpers for consistent error format

### Phase 10: Request ID & Middleware ✅ COMPLETED

**Date**: 2026-02-07

**Files Created:**

- `internal/transport/http/middleware/middleware.go` (~130 lines) - HTTP middleware stack
- `internal/transport/http/middleware/middleware_test.go` (~200 lines) - Comprehensive middleware tests

**Files Modified:**

- `internal/config/config.go` - Added AdminPassword field
- `internal/app/router.go` - Added middleware chain, RouterOptions extended with Logger and AdminPassword
- `cmd/api/main.go` - Added setupLogger function, wired admin password and logger to router

**Middleware Implemented:**

- `RequestID` - Generates unique request ID (16-char hex) or uses existing X-Request-ID header
- `RequestLogger` - Logs requests with method, path, status, duration, and request_id using slog
- `CORS` - Adds Cross-Origin Resource Sharing headers for Web UI compatibility
- `AdminAuth` - Optional password authentication for admin routes (no-op if password not set)

**Features Implemented:**

- [x] Request ID generation with context propagation
- [x] Request logging with configurable slog logger
- [x] CORS headers for cross-origin requests (Web UI)
- [x] OPTIONS preflight handling
- [x] Optional admin authentication via GOATWAY_ADMIN_PASSWORD
- [x] Bearer token authentication for admin routes
- [x] Response writer wrapper that preserves http.Flusher for streaming
- [x] Helper function GetRequestID(ctx) for accessing request ID

**Test Coverage:**

- RequestID generation and passthrough
- CORS headers and OPTIONS handling
- AdminAuth with password/no password scenarios
- RequestLogger execution
- Response writer flush support

**Configuration:**

```bash
# Optional: Set admin password for admin API protection
export GOATWAY_ADMIN_PASSWORD="your-secret-password"

# Optional: Configure log level and format
export LOG_LEVEL="info"    # debug, info, warn, error
export LOG_FORMAT="text"   # text or json
```

**Technical Decisions:**

- **Middleware order** - CORS → RequestID → RequestLogger (outer to inner)
- **Admin auth on routes** - Applied per-route to admin handlers, not globally
- **Localhost-first design** - No auth required if GOATWAY_ADMIN_PASSWORD not set
- **slog integration** - Uses Go 1.21+ structured logging with configurable handler

### Phase 11: CLI & Distribution ✅ COMPLETED

**Date**: 2026-02-08

**Files Created:**

- `install.sh` (~100 lines) - One-line installation script with OS/arch detection

**Files Modified:**

- `cmd/api/main.go` - Added CLI flags (-addr, -data-dir, -version, -v), version variables with ldflags
- `Makefile` - Added VERSION/COMMIT/BUILD_TIME variables, LDFLAGS, build-all, install, release targets

**CLI Flags Implemented:**

```bash
goatway -h
Usage of goatway:
  -addr string
        Server address (overrides SERVER_ADDR)
  -data-dir string
        Data directory path (overrides GOATWAY_DATA_DIR)
  -v    Print version and exit (shorthand)
  -version
        Print version and exit
```

**Makefile Targets Added:**

- `make build` - Build with version info via ldflags
- `make install` - Install locally via go install with version info
- `make build-all` - Cross-compile for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- `make release` - Create release archives (.tar.gz for Unix, .zip for Windows)
- `make clean-dist` - Clean dist directory

**Version Information:**

Version info is injected at build time via ldflags:

```bash
# Build with automatic version detection
make build

# Build with specific version
make build VERSION=1.0.0

# Check version
./bin/goatway -version
goatway 1.0.0
  commit:  abc1234
  built:   2026-02-08T12:00:00Z
```

**Installation Script Features:**

- OS detection (darwin, linux, windows)
- Architecture detection (amd64, arm64)
- Latest version detection from GitHub releases
- Automatic sudo elevation when needed
- Colored output for better UX
- Environment variable overrides (INSTALL_DIR, VERSION)

**Technical Decisions:**

- **ldflags injection** - Version, commit, and build time set at compile time for zero runtime overhead
- **CLI flag precedence** - CLI flags override environment variables via os.Setenv before config.Load()
- **Cross-compilation** - All platforms built from single machine using GOOS/GOARCH
- **Release archives** - tar.gz for Unix systems, zip for Windows
- **Install script** - Detects GitHub releases, no build tools required on target machine
