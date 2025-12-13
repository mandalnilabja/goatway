# AGENTS.md

Mandatory rules for AI coding agents working on Goatway.

**These are constraints, not suggestions.** Changes that violate these rules will be rejected.

---

## Project Core

Goatway is a **lightweight, streaming-safe, OpenAI-compatible HTTP proxy**.

**Non-negotiable priorities:**
- Streaming correctness (`text/event-stream`) > features
- Low latency > abstraction
- Explicitness > magic
- Stability > refactors

**Any change that risks breaking streaming is out of bounds.**

---

## Rules

### ✅ Allowed

- Fix bugs with minimal, targeted diffs
- Improve streaming correctness/robustness
- Improve error handling for upstream failures
- Add backward-compatible abstractions
- Improve documentation
- Add tests (when infrastructure exists)
- Add observability (logs/metrics) without changing behavior

### ❌ Prohibited

- Breaking or buffering streaming responses
- Changing public API without an issue
- New dependencies without approval
- Large refactors without design discussion
- Repo-wide reformatting
- Frameworks, ORMs, or middleware layers
- Retries, queues, or async workers in proxy path

---

## Streaming Rules (Critical)

For code touching the proxy handler:

- `http.Transport.DisableCompression` MUST be `true`
- `text/event-stream` responses MUST flush immediately
- `http.Flusher.Flush()` MUST be called after each write
- Client context MUST propagate end-to-end
- No buffering, accumulation, or transformation of SSE chunks
- No retries or background goroutines in request path

**If unsure whether a change affects streaming: do not proceed.**

---

## Technical Constraints

**Code Organization:**
- Entry point logic in main command file only
- HTTP handlers as methods on repository struct
- Shared dependencies in repository struct
- No global state outside initialization
- No cross-package imports from internal packages

**Headers:**
- Filter hop-by-hop headers
- Forward all others verbatim
- Provider headers must be additive, not destructive

**Concurrency:**
- Goroutines must be request-scoped or lifecycle-managed
- No global goroutines or unbounded channels
- Prefer synchronous streaming

**Dependencies:**
- Prefer standard library
- No web frameworks, middleware stacks, or observability SDKs that alter flow
- Current surface is intentional and minimal

**Testing:**
- Use Go's `testing` package
- Table-driven tests preferred
- Mock external HTTP calls
- Streaming logic MUST be tested
- **No tests → no new logic**

**Commits:**
- Atomic and scoped
- Conventional commits: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`
- PRs describe **why**, not just **what**

---

## Philosophy

This is infrastructure code. Optimize for:
- Predictability
- Traceability  
- Debuggability

Avoid:
- Over-abstraction
- Premature extensibility
- Speculative design

**If a change is "nice to have" rather than "necessary" — do not make it.**

---

## Conflict Resolution

If this file conflicts with other documentation: **AGENTS.md takes precedence.**
