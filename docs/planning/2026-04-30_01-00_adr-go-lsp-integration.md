# ADR: Should we integrate sourcegraph/go-lsp?

**Date:** 2026-04-30
**Status:** Rejected
**Context:** Evaluating whether to replace hand-rolled LSP types with `github.com/sourcegraph/go-lsp`

---

## Decision

**No.** We will not integrate `sourcegraph/go-lsp`. Our self-contained `lsp.go` is the correct approach.

## Reasoning

### 1. Library is archived and unmaintained

Sourcegraph archived `go-lsp` on **August 16, 2024**. It is read-only. Depending on an archived library is a maintenance liability with no path for bug fixes or spec updates.

### 2. Zero dependency benefit

Our current `lsp.go` is ~170 lines of self-contained code defining exactly the types we need (`LSPDiagnostic`, `LSPRange`, `LSPPosition`, `LSPRelatedInfo`, `LSPLocation`) with no external dependency. Adding `go-lsp` would introduce a dependency for types we already have.

### 3. Type mismatch overhead is identical

Our `Position` is 1-based with `Offset` and `File` fields. LSP's `Position` is 0-based with just `Line`/`Character`. We handle conversion correctly via `toZeroBased()` and the `FromLSP`/`ToLSP` converters. Using `go-lsp` types would require the same conversion layer — we'd map *their* types instead of *our own*.

### 4. Our types are richer

- Our `LSPRelatedInfo` includes related information
- Our severity mapping handles `Critical → Error`
- `FromLSP` preserves raw severity in metadata
- The `go-lsp` `Diagnostic` struct lacks `RelatedInformation` in its base type

### 5. Violates design principle

AGENTS.md states: *"Minimal dependencies — core types depend only on stdlib."* The core package currently depends on zero external libraries. Adding `go-lsp` contradicts this principle for no functional gain.

## What we already do well

- Correct 0-based ↔ 1-based conversion via `toZeroBased()`
- Bidirectional `ToLSP()` / `FromLSP()` with lossless metadata preservation
- SARIF output as an additional interchange format
- `go/analysis` integration via `diagnostic.go`

## When to reconsider

If we ever build a **full language server** (not just outputting diagnostics), we would need the full LSP message types (initialize, completion, hover, text synchronization, etc.). At that point, consider:

- `go-language-server` packages (actively maintained forks)
- Protocol types generated from the official LSP JSON schema

Currently, this library *outputs* LSP diagnostics — it is not an LSP server. The hand-rolled types are sufficient and appropriate.
