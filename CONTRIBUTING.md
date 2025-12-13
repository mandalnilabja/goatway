# Contributing to Goatway

Thank you for your interest in contributing to **Goatway**. The project is intentionally designed to be **lightweight, correct, and operationally predictable**. Contributions are welcome, provided they align with these goals.

Goatway is licensed under the **MIT License**. By contributing, you agree that your contributions will be licensed under the same terms.

---

## Guiding Principles

All contributions are evaluated against the following principles:

1. **Correctness Over Features**  
   Protocol fidelity, streaming correctness, and OpenAI compatibility take precedence over feature expansion.

2. **Minimalism as a Design Goal**  
   Goatway is not a framework, platform, or control plane. Avoid unnecessary abstractions and feature creep.

3. **Strict OpenAI Compatibility**  
   Request and response semantics must remain OpenAI-compatible. Breaking compatibility is unacceptable.

4. **Stateless by Default**  
   The system should avoid hidden persistence, implicit state, or side effects unless explicitly justified.

---

## Ways to Contribute

Contributions may include, but are not limited to:

- Bug fixes
- Performance or memory usage improvements
- Test coverage additions
- Documentation improvements
- New LLM provider implementations (within the planned provider architecture)
- Refactoring that improves clarity without altering behavior

If you are unsure whether a change is appropriate, please open an issue for discussion before submitting a pull request.

---

## Development Setup

### Prerequisites

- Go **1.25.3** or later
- `make` (optional but recommended)

### Clone the Repository

```bash
git clone https://github.com/mandalnilabja/goatway.git
cd goatway
```

### Build

```bash
make build
```

### Run

```bash
make run
```

---

## Code Structure

For code organization and maintainability, refer to the discussions on project layout.


### General Rules

* Avoid circular dependencies
* Avoid global state unless strictly necessary
* Do not silently swallow errors
* Do not introduce unnecessary abstractions

---

## Streaming and Proxying Requirements

If your contribution affects request proxying or streaming behavior:

* Do **not** buffer full responses
* Do **not** modify SSE framing
* Do **not** read entire responses into memory
* Stream data incrementally as it is received
* Preserve headers and HTTP status codes
* Treat latency regressions as defects

Changes that break streaming compatibility will not be accepted.

---

## Testing Guidelines

Testing support is a roadmap priority, and contributions in this area are strongly encouraged.

Preferred test types include:

* Unit tests for handlers and utilities
* Integration tests using mocked OpenAI or OpenRouter endpoints
* Streaming tests that validate chunk-by-chunk forwarding behavior

Use Go’s standard `testing` package unless there is a strong reason not to.

---

## Commit Guidelines

* Keep commits small and focused
* Limit each commit to a single logical change
* Write clear, descriptive commit messages

**Example (Good):**

```
proxy: preserve SSE framing for streamed responses
```

**Example (Bad):**

```
fix stuff
```

---

## Pull Request Process

1. Fork the repository
2. Create a feature branch from `main`
3. Implement your changes
4. Ensure `go build ./...` completes successfully
5. Open a pull request that clearly describes:

   * What was changed
   * Why the change was made
   * Any tradeoffs or known limitations

Large or architectural changes must be explicitly justified.

---

## Non-Goals

Pull requests will not be accepted for the following:

* Frontend or UI development (out of scope for this repository)
* Vendor lock-in abstractions
* Features that turn Goatway into a SaaS product

---

## Reporting Bugs

When reporting a bug, please open an issue and include:

* Expected behavior
* Actual behavior
* Steps to reproduce
* Relevant logs, if applicable
* Go version and operating system

Issues with minimal, reproducible examples will be prioritized.

---

## Code of Conduct

Maintain a professional and respectful tone. Technical disagreements should be resolved using data, benchmarks, and code—not personal opinion.

---

Contributions that make Goatway **simpler, faster, or more correct** are likely to be accepted. Contributions that add unnecessary complexity or enterprise-oriented overhead are not.

