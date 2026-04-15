# Contributing to go-finding

Thank you for your interest in contributing to `go-finding`. This document outlines the development setup, coding standards, and PR process.

## Development Setup

### Prerequisites

- Go 1.22 or later
- `golang.org/x/tools` (for `go/analysis` integration)
- `golang.org/x/sync` (for errgroup)

### Getting Started

```bash
git clone https://github.com/larsartmann/go-finding.git
cd go-finding
go mod download
```

### Build & Test

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Build the CLI
go build ./cmd/go-finding

# Build all examples
go build ./examples/...
```

### Using Just (optional)

If you have [just](https://github.com/casey/just) installed:

```bash
just test
just bench
just lint
```

## Project Structure

```
go-finding/
├── finding.go          # Core Finding type
├── severity.go         # Severity enum
├── category.go         # Category constants
├── fix_strategy.go     # FixStrategy enum
├── position.go         # Position and Range types
├── report.go           # Report container
├── filter.go           # Filtering and grouping
├── merge.go            # Merge, dedup, correlation
├── sarif.go            # SARIF 2.1.0 output
├── id.go               # ID generation and parsing
├── errors.go           # Structured error types
├── suppression.go      # Suppression handling
├── json.go             # JSON marshaling
├── diagnostic.go       # go/analysis integration
├── lsp.go              # LSP diagnostic conversion
├── pipeline/           # Pipeline package
│   ├── pipeline.go     # Pipeline orchestrator
│   ├── conflict.go     # Fix conflict detection
│   ├── astfix.go       # AST-aware fix application
│   ├── verify.go       # Verification stage
│   ├── metrics.go      # Metrics collection
│   ├── retry.go        # Retry with backoff
│   └── partial.go      # Partial success
├── cmd/go-finding/     # CLI tool
└── examples/           # Detector examples
```

## Coding Standards

### Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Prefer composition over inheritance
- Use functional options for configuration
- Early returns over nested conditionals
- Descriptive names over comments
- No one-letter variable names except in tight loops

### Types

- All exported enums are string types (`Severity`, `Category`, `FixStrategy`, `ErrorCategory`)
- Use `IsValid()` methods for validation
- Make impossible states unrepresentable
- Immutable data: `Finding` structs are data, not state machines

### Error Handling

Use structured `FindingError` with categories:

```go
return finding.NewValidationError("invalid severity", nil)
return finding.NewIOError("read file", err).WithPosition(pos)
```

Never return bare `fmt.Errorf` from library code. Wrap external errors:

```go
return fmt.Errorf("detector %s: %w", d.Name(), err)
```

### Testing Requirements

- All new code must have tests
- Test behavior, not implementation
- Integration tests over unit tests where practical
- Real implementations over mocks
- Use `Example*()` functions for GoDoc examples
- Use `testing.F` for fuzz tests when testing parsing/serialization

### Naming Conventions

- Filter predicates: `BySeverity`, `ByCategory`, `ByFile`
- Configuration options: `WithDeduplication`, `WithDeduplicateBy`
- Constructor functions: `NewReport`, `NewMetrics`, `NewFixApplier`
- Boolean getters: `IsValid`, `HasFix`, `HasEnd`, `IsSuppressed`

## PR Process

### Before Submitting

1. Run the full test suite: `go test ./...`
2. Run vet: `go vet ./...`
3. Verify all examples build: `go build ./...`
4. Ensure GoDoc examples pass: `go test -run Example ./...`
5. Check test coverage has not decreased

### PR Guidelines

- One logical change per PR
- Write clear commit messages explaining why, not what
- Include tests for all new functionality
- Update documentation if adding public API surface
- Keep PRs focused and reviewable

### Commit Messages

Use conventional commit format:

```
feat: add support for custom deduplication strategies
fix: handle Windows paths in ID generation correctly
docs: add usage guide for pipeline configuration
test: add fuzz tests for ID parsing
refactor: extract severity comparison to standalone method
```

## Architecture Decisions

### Zero Dependencies for Core Types

The root `finding` package uses only standard library types. This makes it safe to import in any Go project without dependency conflicts.

### Lossless Conversions

All conversions (Finding → SARIF, Finding → LSP Diagnostic) preserve the original data. Round-trip conversions should be lossless.

### String-Based Enums

`Severity`, `Category`, `FixStrategy` are string types, not int-based enums. This ensures JSON serialization is human-readable and avoids int↔string conversion errors.

### Pipeline as Orchestrator

The pipeline uses a functional stage pattern. Each stage is a pure function that transforms input to output. The pipeline manages iteration, context, and error propagation.
