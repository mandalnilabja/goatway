# Goatway Codebase Refactoring Plan

> **Status: COMPLETED** - All source files are now under 200 lines.

## Context

This document outlines a comprehensive audit of the Goatway codebase to ensure all code files adhere to the 200-line limit and maintain single-responsibility principles as specified in the project guidelines.

## Refactoring Results

### Files That Were Refactored

| Original File | Lines | Action | New Files Created |
|---------------|-------|--------|-------------------|
| `internal/storage/sqlite.go` | 666 → 124 | Split into 6 files | `sqlite_credentials.go`, `sqlite_credentials_read.go`, `sqlite_logs.go`, `sqlite_usage.go`, `sqlite_helpers.go` |
| `internal/tokenizer/counter.go` | 251 → 126 | Split into 3 files | `counter_content.go`, `counter_tools.go` |
| `internal/provider/openrouter.go` | 224 → 121 | Split into 2 files | `openrouter_response.go` |

### Current File Line Counts (Source Files)

All source files are now under 200 lines:

| File | Lines |
|------|-------|
| `internal/transport/http/handler/admin_credentials.go` | 190 |
| `cmd/api/main.go` | 168 |
| `internal/transport/http/handler/admin_usage.go` | 163 |
| `internal/transport/http/handler/proxy.go` | 156 |
| `internal/storage/sqlite_usage.go` | 153 |
| `internal/storage/sqlite_credentials.go` | 145 |
| `internal/transport/http/middleware/middleware.go` | 140 |
| `internal/tokenizer/counter.go` | 126 |
| `internal/storage/sqlite_logs.go` | 126 |
| `internal/storage/sqlite.go` | 124 |
| `internal/storage/sqlite_credentials_read.go` | 122 |
| `internal/provider/openrouter.go` | 121 |
| `internal/provider/openrouter_response.go` | 109 |

### Test Files (Not Refactored - Low Priority)

| File | Lines | Status |
|------|-------|--------|
| `internal/storage/sqlite_test.go` | 381 | LOW (test) |
| `internal/tokenizer/tokenizer_test.go` | 310 | LOW (test) |
| `internal/transport/http/middleware/middleware_test.go` | 214 | LOW (test) |

---

## Original Audit Summary

### Files That Exceeded 200 Lines (Before Refactoring)

| File | Lines | Priority |
|------|-------|----------|
| `internal/storage/sqlite.go` | 666 | CRITICAL |
| `internal/storage/sqlite_test.go` | 381 | LOW (test) |
| `internal/tokenizer/tokenizer_test.go` | 310 | LOW (test) |
| `internal/tokenizer/counter.go` | 251 | HIGH |
| `internal/provider/openrouter.go` | 224 | MEDIUM |
| `internal/transport/http/middleware/middleware_test.go` | 214 | LOW (test) |

### Files Within Acceptable Limits

| File | Lines | Status |
|------|-------|--------|
| `cmd/api/main.go` | 168 | OK |
| `internal/app/router.go` | 111 | OK |
| `internal/app/server.go` | 43 | OK |
| `internal/config/config.go` | 78 | OK |
| `internal/config/paths.go` | 61 | OK |
| `internal/provider/provider.go` | 67 | OK |
| `internal/provider/stream.go` | 118 | OK |
| `internal/storage/storage.go` | 38 | OK |
| `internal/storage/models.go` | 116 | OK |
| `internal/storage/encryption.go` | 122 | OK |
| `internal/tokenizer/tokenizer.go` | 115 | OK |
| `internal/transport/http/handler/repo.go` | 30 | OK |
| `internal/transport/http/handler/proxy.go` | 156 | OK |
| `internal/transport/http/handler/models.go` | 120 | OK |
| `internal/transport/http/handler/admin_credentials.go` | 190 | OK |
| `internal/transport/http/handler/admin_usage.go` | 163 | OK |
| `internal/transport/http/handler/admin_system.go` | 71 | OK |
| `internal/transport/http/handler/health.go` | 21 | OK |
| `internal/transport/http/handler/cache.go` | 39 | OK |
| `internal/transport/http/handler/webui.go` | 51 | OK |
| `internal/transport/http/middleware/middleware.go` | 140 | OK |
| `internal/types/errors.go` | 96 | OK |
| `internal/types/message.go` | 113 | OK |
| `internal/types/request.go` | 102 | OK |
| `internal/types/response.go` | 84 | OK |
| `internal/types/tools.go` | 98 | OK |
| `internal/types/stream.go` | 58 | OK |
| `internal/types/json.go` | 18 | OK |
| `web/embed.go` | 10 | OK |

---

## Refactoring Plan

### 1. CRITICAL: Split `internal/storage/sqlite.go` (666 → ~150 lines each)

**Current structure:**
- Core struct + constructor + migration + close (lines 1-124)
- Credential CRUD operations (lines 125-378)
- Request logging operations (lines 380-498)
- Usage statistics operations (lines 500-650)
- Helper functions (lines 652-666)

**Proposed split:**

#### 1.1 `internal/storage/sqlite.go` (~120 lines)
- `SQLiteStorage` struct definition
- `NewSQLiteStorage()` constructor
- `Migrate()` schema creation
- `Close()` method
- `generateID()` helper

#### 1.2 `internal/storage/sqlite_credentials.go` (~180 lines)
- `CreateCredential()`
- `GetCredential()`
- `GetDefaultCredential()`
- `ListCredentials()`
- `UpdateCredential()`
- `DeleteCredential()`
- `SetDefaultCredential()`

#### 1.3 `internal/storage/sqlite_logs.go` (~120 lines)
- `LogRequest()`
- `GetRequestLogs()`
- `DeleteRequestLogs()`

#### 1.4 `internal/storage/sqlite_usage.go` (~130 lines)
- `GetUsageStats()`
- `GetDailyUsage()`
- `UpdateDailyUsage()`

#### 1.5 `internal/storage/sqlite_helpers.go` (~20 lines)
- `boolToInt()` helper
- `nullString()` helper

---

### 2. HIGH: Split `internal/tokenizer/counter.go` (251 → ~120 lines each)

**Current structure:**
- Constants (lines 1-28)
- Message counting functions (lines 30-118)
- Content counting (lines 120-166)
- Tool counting functions (lines 168-251)

**Proposed split:**

#### 2.1 `internal/tokenizer/counter.go` (~130 lines)
- Keep constants
- `CountMessages()`
- `CountRequest()`
- `countMessage()`
- `getMessageOverhead()`

#### 2.2 `internal/tokenizer/counter_content.go` (~70 lines)
- `countContent()`
- `countImageTokens()`

#### 2.3 `internal/tokenizer/counter_tools.go` (~60 lines)
- `countTools()`
- `countToolCalls()`

---

### 3. MEDIUM: Split `internal/provider/openrouter.go` (224 → ~120 lines each)

**Current structure:**
- Provider struct + basic methods (lines 1-42)
- `ProxyRequest()` main method (lines 44-124)
- Response handlers (lines 126-224)

**Proposed split:**

#### 3.1 `internal/provider/openrouter.go` (~130 lines)
- `OpenRouterProvider` struct
- `NewOpenRouterProvider()`
- `Name()`
- `BaseURL()`
- `PrepareRequest()`
- `ProxyRequest()`

#### 3.2 `internal/provider/openrouter_response.go` (~100 lines)
- `handleStreamingResponse()`
- `handleJSONResponse()`
- `handleErrorResponse()`

---

### 4. LOW: Split Test Files (Optional)

Test files exceeding 200 lines can be split by test category:

#### 4.1 `internal/storage/sqlite_test.go` (381 lines)
Split into:
- `sqlite_credentials_test.go` - Credential tests
- `sqlite_logs_test.go` - Log tests
- `sqlite_usage_test.go` - Usage tests

#### 4.2 `internal/tokenizer/tokenizer_test.go` (310 lines)
Split into:
- `tokenizer_test.go` - Core encoding tests
- `counter_test.go` - Message counting tests
- `counter_content_test.go` - Content counting tests

#### 4.3 `internal/transport/http/middleware/middleware_test.go` (214 lines)
Split into:
- `request_id_test.go` - Request ID middleware tests
- `logging_test.go` - Logging middleware tests
- `cors_test.go` - CORS middleware tests
- `auth_test.go` - Admin auth tests

---

## Function-Level Analysis

### Functions with Multiple Responsibilities (Consider Refactoring)

| Function | File | Issue | Recommendation |
|----------|------|-------|----------------|
| `main()` | `cmd/api/main.go` | 94 lines, handles many setup steps | Consider extracting `initStorage()`, `initProvider()`, `initRouter()` helpers |
| `ProxyRequest()` | `openrouter.go` | 80 lines, handles request creation + execution + routing | Already planning to extract response handlers |
| `GetRequestLogs()` | `sqlite.go` | 72 lines, builds dynamic query | Consider extracting query builder |
| `GetUsageStats()` | `sqlite.go` | 82 lines, two separate queries | Consider splitting aggregate + model breakdown |

### Well-Structured Functions (No Changes Needed)

Most handler functions follow single-responsibility:
- `CreateCredential()` - Single responsibility: create credential
- `ListCredentials()` - Single responsibility: list credentials
- `GetUsageStats()` - Single responsibility: get stats
- All middleware functions - Clear, focused responsibilities

---

## Implementation Order

1. **Phase 1: Critical** - Split `sqlite.go` (highest impact)
2. **Phase 2: High** - Split `counter.go`
3. **Phase 3: Medium** - Split `openrouter.go`
4. **Phase 4: Low** - Split test files (optional, improves maintainability)

---

## Verification

After refactoring, verify:

1. **Build passes:**
   ```bash
   make build
   ```

2. **Tests pass:**
   ```bash
   make test
   ```

3. **Line count check:**
   ```bash
   find . -name "*.go" -exec wc -l {} \; | sort -rn | head -20
   ```
   All files should be under 200 lines.

4. **Lint passes:**
   ```bash
   make lint
   ```

5. **Functional test:**
   ```bash
   make run
   # Test proxy endpoint
   curl -X POST http://localhost:8080/v1/chat/completions \
     -H "Authorization: Bearer $API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "hi"}]}'
   ```

---

## Notes

- The `internal/types/` package is already well-organized with clear separation
- The handler package is well-structured with separate files per domain
- No circular dependencies will be introduced by these splits
- All splits maintain the same package, just different files
