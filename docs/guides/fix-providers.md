# FixProvider Authoring Guide

This guide explains how to write custom FixProviders for domain-specific
fix application, and when to implement the optional `lineIndexAware` interface.

## Overview

A `FixProvider` converts a `finding.Finding` into one or more `pipeline.FixEdit`
operations (byte-level edits). The FixEngine tries providers in order; the first
whose `CanHandle` returns `true` wins.

```
Finding → FixEngine → [Provider1, Provider2, ...] → FixEdit[] → applyEditsToContent → Modified Content
```

## Built-in Providers

| Provider            | Name          | Handles                             | Accuracy |
| ------------------- | ------------- | ----------------------------------- | -------- |
| `OffsetProvider`    | `byte-offset` | Findings with byte-offset Range     | Exact    |
| `LineProvider`      | `line-column` | Findings with line/column Position  | High     |
| `SubstringProvider` | `substring`   | Findings with BeforeCode (fallback) | Medium   |
| `goast.Provider`    | `go-ast`      | `.go` files with any code change    | Highest  |

## Writing a Custom Provider

### Interface

```go
type FixProvider interface {
    Name() string
    CanHandle(f finding.Finding) bool
    Edits(content []byte, f finding.Finding) ([]FixEdit, error)
}
```

### Minimal Example: Rust Syntax Provider

```go
type RustProvider struct{}

func (*RustProvider) Name() string { return "rust-syntax" }

func (*RustProvider) CanHandle(f finding.Finding) bool {
    return f.HasCodeChange() && strings.HasSuffix(f.Position.File, ".rs")
}

func (p *RustProvider) Edits(content []byte, f finding.Finding) ([]pipeline.FixEdit, error) {
    // Parse Rust source via syn equivalent...
    // Produce precise byte-level edits from the parsed IR.
    return []pipeline.FixEdit{{
        Offset:      offset,
        Length:      length,
        Replacement: []byte(f.AfterCode),
        Source:      f,
    }}, nil
}
```

### Registration

```go
cfg := pipeline.Config{
    FixProviders: []pipeline.FixProvider{
        &pipeline.OffsetProvider{},    // byte-offset findings (no parse)
        &goast.Provider{},             // AST-aware for .go files
        &pipeline.LineProvider{},      // line/column fallback
        &pipeline.SubstringProvider{}, // last resort
    },
}
```

### CLI Registration

Use the `-fix-provider` flag or YAML config:

```bash
go-finding -fix-provider go-ast -dir ./...
```

```yaml
# go-finding.yaml
fixProviders:
  - go-ast
```

## The `lineIndexAware` Interface (Optional)

Providers that need line-to-offset conversion can implement this optional
interface to receive a pre-built line offset index:

```go
type lineIndexAware interface {
    EditsWithLineIndex(content []byte, lineIndex []int, f finding.Finding) ([]FixEdit, error)
}
```

### When to Implement It

Implement `lineIndexAware` when your provider:

1. **Converts line/column to byte offsets** — The index provides O(1) lookup
   instead of O(n) newline scanning per finding.
2. **Processes many findings in the same file** — The index is built once per
   file and reused across all findings, saving O(n × m) where n = file size
   and m = finding count.

### When NOT to Implement It

Skip `lineIndexAware` when your provider:

1. **Uses its own parser** (e.g., `go/parser`, tree-sitter) — Your parser's
   `token.FileSet` already provides line-to-offset conversion.
2. **Only uses byte offsets** — No line conversion needed; `OffsetProvider`
   pattern.
3. **Handles very few findings per file** — The overhead of building the index
   may exceed the savings.

### Implementation Pattern

```go
func (p *MyProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
    return p.EditsWithLineIndex(content, buildLineOffsetIndex(content), f)
}

func (p *MyProvider) EditsWithLineIndex(content []byte, idx []int, f finding.Finding) ([]FixEdit, error) {
    offset, err := indexLineColToOffset(idx, len(content), f.Position.Line, f.Position.Column)
    if err != nil {
        return nil, nil
    }
    // Use offset to produce edits...
}
```

The `Edits` method builds a fresh index for backward compatibility. The
FixEngine detects `lineIndexAware` and passes a shared index instead.

### How the FixEngine Uses It

```go
// In FixEngine.resolveEdits:
if la, ok := p.(lineIndexAware); ok {
    if *lineIndex == nil {
        *lineIndex = buildLineOffsetIndex(content)  // built ONCE
    }
    edits, err = la.EditsWithLineIndex(content, *lineIndex, f)
} else {
    edits, err = p.Edits(content, f)
}
```

The index is **lazily built** — only when the first `lineIndexAware` provider
actually handles a finding. If all findings are handled by `OffsetProvider`
(byte offsets), the index is never built.

## FixEdit Construction

Always set the `Source` field to the originating finding:

```go
return []pipeline.FixEdit{{
    Offset:      startByte,
    Length:      bytesToRemove,
    Replacement: []byte(f.AfterCode),
    Source:      f,
}}, nil
```

### Edit Types

| Type    | Length | Replacement | Description              |
| ------- | ------ | ----------- | ------------------------ |
| Replace | > 0    | non-empty   | Remove bytes, insert new |
| Delete  | > 0    | empty/nil   | Remove bytes             |
| Insert  | 0      | non-empty   | Insert at offset         |

## Error Handling

- Return `(nil, nil)` if the finding cannot be resolved — the engine falls
  through to the next provider.
- Return `(nil, err)` for unexpected failures — the error is collected and
  surfaced if no other provider succeeds.
- Parse failures should return `(nil, nil)`, not an error, so text-based
  providers get a chance to handle the finding.

## Caching

If your provider parses content (AST, IR), cache the result per unique content:

```go
type MyProvider struct {
    mu    sync.Mutex
    cache struct {
        hash  uint64
        parse *MyParseResult
    }
}

func (p *MyProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
    h := fnv.New64a()
    h.Write(content)
    hash := h.Sum64()

    p.mu.Lock()
    defer p.mu.Unlock()

    if p.cache.hash == hash && p.cache.parse != nil {
        return p.editsFromParse(p.cache.parse, f)
    }

    parsed, err := parseContent(content)
    if err != nil {
        return nil, nil // fall through to next provider
    }

    p.cache.hash = hash
    p.cache.parse = parsed

    return p.editsFromParse(parsed, f)
}
```

## Testing

Test your provider in isolation and via the FixEngine:

```go
func TestMyProvider(t *testing.T) {
    p := &MyProvider{}

    // Unit test: direct Edits call
    edits, err := p.Edits(content, finding)
    assert edits and err

    // Integration: through FixEngine
    engine := pipeline.NewFixEngineWithProviders(p, &pipeline.SubstringProvider{})
    result, applied, count := engine.Apply(content, findings)
    assert result content
}
```
