# Credential Name Binding Design

## Overview

Replace provider-based default credential resolution with explicit name-based credential binding per model alias.

## Problem

Current behavior requires credentials to have `is_default = 1` per provider. This is confusing because:
- Users add credentials but forget to set as default
- No way to use different credentials for different model aliases of the same provider
- Implicit behavior leads to unexpected 401 errors

## Solution

Each model alias explicitly references a credential by name via `credential_name` field.

## Config Changes

### config.toml structure

```toml
[default]
provider = "openrouter"
credential_name = "my-openrouter-key"  # Required for unaliased model fallback

[[models]]
slug = "gpt5.2"
provider = "openrouter"
model = "openai/gpt-5.2"
credential_name = "my-openrouter-key"  # Optional, 401 if missing at request time

[[models]]
slug = "premium-model"
provider = "openrouter"
model = "anthropic/claude-3-opus"
credential_name = "expensive-key"  # Can use different credential
```

### Code changes

- Add `CredentialName string` field to `config.ModelAlias` struct
- Add `CredentialName string` field to `config.DefaultRoute` struct
- No validation at startup (missing credential_name fails at request time)

## Database/Storage Changes

### Schema migration

- Remove `is_default` column from `credentials` table
- Add unique constraint on `name` column

### Storage interface

```go
// Remove:
GetDefaultCredential(provider string) (*Credential, error)
SetDefaultCredential(id string) error

// Add:
GetCredentialByName(name string) (*Credential, error)
```

### Credential model

- Remove `IsDefault bool` field from `models.Credential`

### Web UI

- Remove "Set as default" checkbox from credential form
- Remove "Set Default" action from credentials list

## Router Changes

### CredentialResolver

```go
// Change from provider-based to name-based resolution
func (r *CredentialResolver) Resolve(credentialName string) (*Credential, error)

// Cache key: credential_name instead of provider
cache map[string]*cachedCredential
```

### resolvedRoute struct

```go
type resolvedRoute struct {
    provider       types.Provider
    model          string
    credentialName string  // NEW: from config alias or [default]
}
```

### ProxyRequest flow

1. `resolveModel(slug)` returns provider, model, credentialName
2. If `credentialName == ""` → return 401 "No credential configured for model"
3. `credResolver.Resolve(credentialName)` → fetch credential by name
4. If credential not found → return 401 "Credential not found: {name}"
5. Proceed with request

## Error Handling

| Scenario | HTTP Status | Message |
|----------|-------------|---------|
| Model slug not in config, no `[default]` | 400 | "Model not found: {slug}" |
| Model slug not in config, `[default]` has no `credential_name` | 401 | "No credential configured for model: {slug}" |
| Model alias has no `credential_name` | 401 | "No credential configured for model: {slug}" |
| `credential_name` references non-existent credential | 401 | "Credential not found: {name}" |

## Migration Path

1. Existing credentials remain (just lose `is_default` flag)
2. Users must update config.toml to add `credential_name` fields
3. First request to an alias without `credential_name` returns 401
