# Handler Package Reorganization Plan

## Context

The handler package has 16 files with all handlers as methods on a single `Repo` struct. This plan reorganizes into domain-specific subpackages for better testability, discoverability, and maintainability.

**No backward compatibility needed** - no existing users.

## Target Directory Structure

```text
internal/transport/http/handler/
├── handler.go              # Main Repo type (composes domain handlers)
├── shared.go               # writeJSON, writeError utilities
│
├── admin/
│   ├── admin.go            # Handlers type + constructor
│   ├── system.go           # Health, Info, ChangePassword
│   ├── credentials.go      # Credential CRUD
│   ├── apikeys.go          # API key CRUD
│   └── usage.go            # Usage stats, logs
│
├── webui/
│   ├── webui.go            # Handlers type + constructor
│   ├── serve.go            # Static file serving
│   └── auth.go             # Login, Logout, LoginPage
│
├── proxy/
│   ├── proxy.go            # Handlers type + shared logic
│   ├── chat.go             # Chat completions
│   ├── completions.go      # Legacy completions
│   ├── models.go           # ListModels, GetModel
│   ├── embeddings.go       # Embeddings
│   ├── audio.go            # Audio endpoints
│   ├── images.go           # Image endpoints
│   └── moderations.go      # Moderation
│
└── infra/
    ├── health.go           # RootStatus, HealthCheck
    └── cache.go            # Cache handler
```

## Design

### Domain Handler Pattern

Each subpackage has a simple `Handlers` struct with only needed dependencies:

```go
// admin/admin.go
package admin

type Handlers struct {
    storage   storage.Storage
    startTime time.Time
}

func New(storage storage.Storage, startTime time.Time) *Handlers {
    return &Handlers{storage: storage, startTime: startTime}
}
```

### Main Repo Composes Domains

```go
// handler/handler.go
type Repo struct {
    Admin  *admin.Handlers
    WebUI  *webui.Handlers
    Proxy  *proxy.Handlers
    Infra  *infra.Handlers
}

func NewRepo(cache *ristretto.Cache[string, any], prov provider.Provider,
             store storage.Storage, tok tokenizer.Tokenizer) *Repo {
    startTime := time.Now()
    return &Repo{
        Admin:  admin.New(store, startTime),
        WebUI:  webui.New(store, nil), // SessionStore set later
        Proxy:  proxy.New(prov, store, tok, cache),
        Infra:  infra.New(cache, startTime),
    }
}
```

### Router Uses Domain Paths

```go
// router.go - simple, no conditional auth paths
mux.Handle("POST /v1/chat/completions", apiKeyAuth(http.HandlerFunc(repo.Proxy.ChatCompletions)))
mux.Handle("GET /v1/models", apiKeyAuth(http.HandlerFunc(repo.Proxy.ListModels)))
mux.Handle("POST /api/admin/credentials", adminAuth(repo.Admin.CreateCredential))
```

## Implementation Steps

### Step 1: Create Subpackages

Create all 4 subpackages with their handler types and constructors.

### Step 2: Move Handlers

Move handlers to appropriate subpackages, updating method receivers.

### Step 3: Create Shared Utilities

Extract `writeJSON`, `writeError`, `IsValidAdminPassword` to `shared.go`.

### Step 4: Update Main Repo

Replace old Repo with composed version.

### Step 5: Update Router

Change all routes from `repo.Method` to `repo.Domain.Method`.

### Step 6: Delete Old Files

Remove all original handler files from root.

### Step 7: Remove Backward Compat Code

Remove any conditional paths in router that exist for backward compatibility:

- Remove "no auth configured" fallback paths in router.go
- Always require API key auth for proxy routes
- Always require admin auth for admin routes

## Files to Modify

| Action | Files |
| ------ | ----- |
| Create | `admin/admin.go`, `admin/system.go`, `admin/credentials.go`, `admin/apikeys.go`, `admin/usage.go` |
| Create | `webui/webui.go`, `webui/serve.go`, `webui/auth.go` |
| Create | `proxy/proxy.go`, `proxy/chat.go`, `proxy/completions.go`, `proxy/models.go`, `proxy/embeddings.go`, `proxy/audio.go`, `proxy/images.go`, `proxy/moderations.go` |
| Create | `infra/health.go`, `infra/cache.go` |
| Create | `handler.go`, `shared.go` |
| Modify | `internal/app/router.go` - update all route registrations |
| Delete | All 16 original handler files in root |

## Simplifications (No Backward Compat)

1. **Router**: Remove duplicate route blocks for "no auth" mode - always require auth
2. **WebUI**: Remove legacy "no session auth" path - always use session auth
3. **Dependencies**: Pass concrete types directly, no interface indirection needed for now

## Verification

```bash
make build          # Must compile
make test           # Must pass
make lint           # No issues
# Manual: test streaming endpoint with actual request
```

## Out of Scope

- **Tokenizer**: Current organization is fine
- **Types**: Endpoint-based organization matches OpenAI patterns
