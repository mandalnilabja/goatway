# Config File & Model Routing Implementation Plan

## Context

Goatway currently loads configuration from environment variables only. Users need:
1. **Persistent configuration** via `~/.goatway/config.toml`
2. **Custom model aliases** for routing (e.g., `opnr-gpt5` → `openrouter/openai/gpt-5`)

This enables users to define short, memorable model slugs that route to specific provider+model combinations.

## Design Decisions

- **Config format:** TOML with array of tables for model aliases
- **Priority:** CLI flags → Env vars → config.toml → defaults
- **Routing behavior:** Only explicitly aliased models work by default; optional `[default]` section for fallback
- **Bootstrap:** Auto-create config.toml with commented examples on first run

## Config Structure

```toml
# ~/.goatway/config.toml
server_port = ":8080"
enable_web_ui = true

# Optional default routing
[default]
provider = "openrouter"
model = "openai/gpt-4o"

# Model aliases
[[models]]
slug = "opnr-gpt5"
provider = "openrouter"
model = "openai/gpt-5"

[[models]]
slug = "claude"
provider = "openrouter"
model = "anthropic/claude-3.5-sonnet"
```

## Implementation Steps

### Step 1: Add TOML parsing to config package

**File:** `internal/config/file.go` (new, ~80 lines)

```go
type FileConfig struct {
    ServerPort  string       `toml:"server_port"`
    EnableWebUI bool         `toml:"enable_web_ui"`
    Default     *DefaultRoute `toml:"default"`
    Models      []ModelAlias `toml:"models"`
}

type DefaultRoute struct {
    Provider string `toml:"provider"`
    Model    string `toml:"model"`
}

type ModelAlias struct {
    Slug     string `toml:"slug"`
    Provider string `toml:"provider"`
    Model    string `toml:"model"`
}

func LoadFile() (*FileConfig, error)
func ConfigPath() string  // ~/.goatway/config.toml
func EnsureConfigFile() error  // creates default if missing
```

**Dependency:** Add `github.com/BurntSushi/toml` to go.mod

### Step 2: Update Config struct to include models

**File:** `internal/config/config.go` (modify)

```go
type Config struct {
    ServerPort  string
    EnableWebUI bool
    Default     *DefaultRoute  // NEW
    Models      []ModelAlias   // NEW
}

func Load() *Config {
    fileConfig, _ := LoadFile()

    cfg := &Config{
        ServerPort:  getEnv("SERVER_PORT", fileConfig.ServerPort, ":8080"),
        EnableWebUI: getEnvBool("ENABLE_WEB_UI", fileConfig.EnableWebUI, true),
        Default:     fileConfig.Default,
        Models:      fileConfig.Models,
    }
    return cfg
}
```

### Step 3: Create Provider Registry

**File:** `internal/provider/registry.go` (new, ~25 lines)

```go
func NewProviders() map[string]Provider {
    return map[string]Provider{
        "openrouter": openrouter.New(),
        // Future: "openai": openai.New(),
        // Future: "ollama": ollama.New(),
    }
}
```

### Step 4: Create Model Router

**File:** `internal/provider/router.go` (new, ~90 lines)

```go
type Router struct {
    providers map[string]Provider
    slugMap   map[string]*resolvedRoute  // CACHED: slug -> resolved route (O(1) lookup)
    default_  *config.DefaultRoute
}

// Pre-resolved at startup for fast lookups
type resolvedRoute struct {
    provider Provider
    model    string
}

func NewRouter(providers map[string]Provider, cfg *config.Config) *Router {
    r := &Router{
        providers: providers,
        slugMap:   make(map[string]*resolvedRoute),
        default_:  cfg.Default,
    }

    // Build slug map at startup (not per-request)
    for _, alias := range cfg.Models {
        if p, ok := providers[alias.Provider]; ok {
            r.slugMap[alias.Slug] = &resolvedRoute{
                provider: p,
                model:    alias.Model,
            }
        }
    }
    return r
}

// Implements Provider interface
func (r *Router) Name() string
func (r *Router) BaseURL() string
func (r *Router) PrepareRequest(ctx, req) error
func (r *Router) ProxyRequest(ctx, w, req, opts) (*ProxyResult, error) {
    resolved, err := r.resolveModel(opts.Model)
    if err != nil {
        return nil, err  // model not found, no default
    }
    opts.Model = resolved.model
    return resolved.provider.ProxyRequest(ctx, w, req, opts)
}

// O(1) map lookup - no iteration over config.Models
func (r *Router) resolveModel(slug string) (*resolvedRoute, error) {
    if route, ok := r.slugMap[slug]; ok {
        return route, nil
    }
    if r.default_ != nil {
        // use default provider with original model name
        ...
    }
    return nil, ErrModelNotFound
}
```

### Step 5: Update main.go to use Router

**File:** `cmd/api/main.go` (modify ~5 lines)

```go
// Before:
llmProvider := openrouter.New()

// After:
providers := provider.NewProviders()
llmProvider := provider.NewRouter(providers, cfg)
```

## Files Summary

| File | Action | Est. Lines |
|------|--------|------------|
| `internal/config/file.go` | Create | ~80 |
| `internal/config/config.go` | Modify | ~15 changed |
| `internal/provider/registry.go` | Create | ~25 |
| `internal/provider/router.go` | Create | ~90 |
| `internal/provider/router_test.go` | Create | ~60 |
| `cmd/api/main.go` | Modify | ~5 changed |

## Verification

### Unit Tests
1. **Config parsing:** Load TOML, verify fields parsed correctly
2. **Router resolution:**
   - Known alias → returns correct provider + model
   - Unknown alias with default → returns default
   - Unknown alias without default → returns error

### Integration Test
```bash
# Create test config
cat > ~/.goatway/config.toml << 'EOF'
[[models]]
slug = "test-gpt"
provider = "openrouter"
model = "openai/gpt-4o"
EOF

# Start server
make run

# Test aliased model
curl -X POST localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "test-gpt", "messages": [{"role": "user", "content": "hi"}]}'

# Verify request went to openrouter with model "openai/gpt-4o"
```

### Manual Verification
1. `make test` - all tests pass
2. `make lint` - no lint errors
3. `make build` - builds successfully
