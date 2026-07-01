# Contributing to go-finding

Thank you for your interest in contributing to `go-finding`. This document outlines the development setup, coding standards, and PR process.

## Development Setup

### Prerequisites

- Go 1.26 or later
- `golangci-lint` (for linting)

### Nix Setup (optional)

If you use [Nix](https://nixos.org/), you can get a reproducible development environment:

```bash
# Enter dev shell (if flake.nix is configured)
nix develop

# Or with direnv (auto-loads on cd)
echo "use flake" > .envrc
direnv allow
```

### Getting Started

```bash
git clone https://github.com/larsartmann/go-finding.git
cd go-finding
go mod download
```

### Build & Test

```bash
# Run all tests (must use -race)
go test -race -count=1 ./...

# Run tests with coverage
go test -race -count=1 -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...

# Build all packages
go build ./...

# Run linter
golangci-lint run ./...

# Run vet
go vet ./...
```

### Using Nix (recommended)

If you have [Nix](https://nixos.org/) with flakes enabled:

```bash
nix run .#test                              # Run all tests with -race
nix run .#bench                             # Run benchmarks
nix run .#lint                              # golangci-lint
nix run .#check                             # All checks (fmt + lint + test)
nix build                                   # Build the CLI binary
```

### Using Go directly

```bash
go test -race -count=1 ./...
go test -bench=. -benchmem ./...
golangci-lint run ./...
go test -coverprofile=coverage.out ./...
go vet ./...
```

## Project Structure

```
go-finding/
├── finding.go           # Core Finding type
├── finding_builder.go   # Fluent builder API
├── severity.go          # Severity enum with comparison
├── confidence.go        # Confidence named type (0.0–1.0)
├── category.go          # Category constants
├── fix_strategy.go      # FixStrategy enum
├── position.go          # Position and Range types
├── report.go            # Report container (thread-safe)
├── filter.go            # Filtering and grouping
├── merge.go             # Merge, dedup, correlation
├── diff.go              # Diff (before/after finding sets)
├── format.go            # FormatText/FormatMarkdown output
├── sarif_export.go      # SARIF 2.1.0 export
├── sarif_import.go      # SARIF 2.1.0 import
├── sarif_types.go       # SARIF struct types and constants
├── id.go                # ID generation and parsing
├── errors.go            # Structured error types
├── tag.go               # Tag type with standard constants
├── suppression.go       # Suppression handling with TTL
├── json.go              # JSON marshaling/unmarshaling
├── lsp.go               # LSP diagnostic conversion
├── version.go           # Version constants
├── analysis/            # go/analysis.Diagnostic integration
│   └── analysis.go
├── pipeline/            # Pipeline package
│   ├── pipeline.go      # Pipeline orchestrator
│   ├── adapters.go      # Detector/Processor interfaces
│   ├── config.go        # Config with validation
│   ├── conflict.go      # Fix conflict detection
│   ├── fix_engine.go    # Byte-level edit engine
│   ├── fix_provider.go  # Composable fix providers
│   ├── fix_applier.go   # Filesystem fix application
│   ├── verify.go        # Verification stage
│   ├── metrics.go       # Metrics collection
│   ├── retry.go         # Retry with exponential backoff
│   └── partial.go       # Partial success
├── cmd/go-finding/      # CLI tool (own module: yaml, go-output deps)
│   └── internal/detectors/  # Built-in detectors (govet, staticcheck)
├── examples/            # Standalone examples (basic, builder)
├── docs/                # Documentation and schemas
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

1. Run the full test suite: `go test -race -count=1 ./...`
2. Run vet: `go vet ./...`
3. Run linter: `golangci-lint run ./...`
4. Verify all examples build: `go build ./...`
5. Ensure GoDoc examples pass: `go test -run Example ./...`
6. Check test coverage has not decreased
7. Verify benchmarks still work: `go test -bench=. -benchmem ./...`

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
