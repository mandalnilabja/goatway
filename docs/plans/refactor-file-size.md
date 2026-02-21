# Refactor Plan: File Size Guidelines

## Context

The CLAUDE.md specifies file size limits:

- Target: under 120 lines
- Worst-case: never exceed 150 lines

Current violations require splitting for human readability and simplicity.

## Summary

| Priority               | Files    | Action       |
| ---------------------- | -------- | ------------ |
| Critical (>150 lines)  | 11 files | Must split   |
| Moderate (120-150 lines) | 8 files  | Should review |

---

## Critical: Files Exceeding 150 Lines

### 1. `internal/tokenizer/tokenizer_test.go` (310 lines)

**Split into:**

- `tokenizer_test.go` - TestNew, TestResolveEncoding, TestEncodingCaching (~80 lines)
- `counter_test.go` - TestCountTokens, TestCountMessages, TestCountRequest (~120 lines)
- `counter_image_test.go` - TestCountImageTokens (~80 lines)

### 2. `internal/transport/http/handler/admin/apikeys.go` (297 lines)

**Current:** 3 request/response structs + 6 handler methods

**Split into:**

- `apikeys_types.go` - CreateAPIKeyRequest, CreateAPIKeyResponse, UpdateAPIKeyRequest (~50 lines)
- `apikeys_crud.go` - CreateAPIKey, GetAPIKeyByID, ListAPIKeys, UpdateAPIKey, DeleteAPIKey (~140 lines)
- `apikeys_rotate.go` - RotateAPIKey (~60 lines)

### 3. `internal/storage/sqlite/credentials.go` (263 lines)

**Current:** 7 methods for credential CRUD

**Split into:**

- `credentials_read.go` - GetCredential, GetDefaultCredential, ListCredentials (~90 lines)
- `credentials_write.go` - CreateCredential, UpdateCredential, DeleteCredential, SetDefaultCredential (~120 lines)

### 4. `internal/storage/sqlite/apikeys.go` (232 lines)

**Current:** 7 methods + scanAPIKeys helper

**Split into:**

- `apikeys_read.go` - GetAPIKey, GetAPIKeyByPrefix, ListAPIKeys, scanAPIKeys (~100 lines)
- `apikeys_write.go` - CreateAPIKey, UpdateAPIKey, DeleteAPIKey, UpdateAPIKeyLastUsed (~100 lines)

### 5. `cmd/api/main.go` (224 lines)

**Current:** main() + helper functions mixed together

**Split into:**

- `main.go` - main(), printVersion() (~80 lines)
- `setup.go` - ensureAdminPassword(), isValidAdminPassword() (~70 lines)
- `logger.go` - setupLogger(), printStartupBanner() (~40 lines)

### 6. `internal/storage/argon2_test.go` (194 lines)

**Split into:**

- `argon2_test.go` - TestHashPassword, TestVerifyPassword (~100 lines)
- `argon2_bench_test.go` - Benchmarks and edge cases (~90 lines)

### 7. `internal/transport/http/handler/admin/credentials.go` (191 lines)

**Current:** 2 request structs + 6 handler methods + 1 helper

**Split into:**

- `credentials_types.go` - CreateCredentialRequest, UpdateCredentialRequest (~30 lines)
- `credentials.go` - All handler methods + extractCredentialID (~130 lines)

### 8. `internal/transport/http/handler/proxy/images.go` (172 lines)

**Current:** 3 methods (ImageGeneration, ImageEdit, ImageVariation)

**Split into:**

- `images_generate.go` - ImageGeneration (~60 lines)
- `images_edit.go` - ImageEdit, ImageVariation (~90 lines)

### 9. `internal/transport/http/handler/proxy/audio.go` (170 lines)

**Current:** 3 methods (TextToSpeech, Transcription, Translation)

**Split into:**

- `audio_tts.go` - TextToSpeech (~60 lines)
- `audio_stt.go` - Transcription, Translation (~90 lines)

### 10. `internal/transport/http/handler/admin/usage.go` (164 lines)

**Current:** 4 handler methods + 2 helper functions

**Split into:**

- `usage.go` - GetUsageStats, GetDailyUsage, parseStatsFilter (~80 lines)
- `usage_logs.go` - GetRequestLogs, DeleteRequestLogs, parseLogFilter (~70 lines)

### 11. `internal/storage/sqlite/usage.go` (155 lines)

**Current:** 3 methods with complex queries

**Split into:**

- `usage_stats.go` - GetUsageStats, GetDailyUsage (~100 lines)
- `usage_update.go` - UpdateDailyUsage (~50 lines)

---

## Moderate: Files 120-150 Lines (Review)

These files are near the limit. Consider splitting if adding features:

| File                       | Lines | Notes                          |
| -------------------------- | ----- | ------------------------------ |
| `storage/argon2.go`        | 147   | OK - cohesive password hashing |
| `storage/sqlite/sqlite.go` | 146   | OK - DB init/migrations        |
| `handler/proxy/proxy.go`   | 144   | Consider splitting if grows    |
| `middleware/auth/session.go` | 141   | OK - single responsibility     |
| `app/router.go`            | 139   | Consider: extract route groups |
| `types/completions.go`     | 138   | OK - type definitions          |
| `middleware/auth/apikey.go` | 131   | OK - single middleware         |
| `storage/sqlite/logs.go`   | 128   | OK - log CRUD                  |

---

## Execution Order

1. **cmd/api/main.go** - Entry point, highest impact
2. **handler/admin/apikeys.go** - Largest handler file
3. **storage/sqlite/credentials.go** - Storage layer
4. **storage/sqlite/apikeys.go** - Storage layer
5. **handler/admin/credentials.go** - Handler layer
6. **handler/proxy/images.go** - Proxy handlers
7. **handler/proxy/audio.go** - Proxy handlers
8. **handler/admin/usage.go** - Admin handlers
9. **storage/sqlite/usage.go** - Storage layer
10. **tokenizer/tokenizer_test.go** - Tests
11. **storage/argon2_test.go** - Tests

---

## Verification

After each file split:

1. Run `make test` - all tests pass
2. Run `make lint` - no linting errors
3. Run `make build` - binary compiles
4. Verify line counts with `wc -l <files>`
