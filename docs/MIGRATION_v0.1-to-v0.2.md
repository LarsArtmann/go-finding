# Migration Guide: v0.1.3 → v0.2.0

## Breaking Changes

### 1. `NewFinding` — New `confidence` parameter

The most common breaking change. `NewFinding` now requires a `confidence` value as the last parameter.

```go
// v0.1.3
f := finding.NewFinding("SA1000", "staticcheck", "unused variable", finding.SeverityWarning, pos)

// v0.2.0
f := finding.NewFinding("SA1000", "staticcheck", "unused variable", finding.SeverityWarning, pos, 0.8)
```

If you don't have a confidence value, use `0.0`. The value is clamped to `[0.0, 1.0]`.

### 2. `Builder.Build()` — Now returns error

```go
// v0.1.3 — panics on invalid builder
f := finding.Builder{}.WithRule("R").Build()

// v0.2.0 — returns error
f, err := finding.Builder{}.WithRule("R").Build()
if err != nil {
    // handle ErrInvalidBuilder
}

// v0.2.0 — panic equivalent for tests/examples
f := finding.Builder{}.WithRule("R").MustBuild()
```

### 3. `WithTags` — Parameter type changed from `string` to `Tag`

```go
// v0.1.3
b := finding.Builder{}.WithTags("security", "performance")

// v0.2.0
b := finding.Builder{}.WithTags(finding.TagSecurity, finding.TagPerformance)

// Custom tags
b := finding.Builder{}.WithTags(finding.Tag("custom-tag"))
```

### 4. `Finding.Tag string` — Deprecated

Use `Finding.Tags []Tag` instead:

```go
// v0.1.3
f.Tag = "security"

// v0.2.0
f.Tags = []finding.Tag{finding.TagSecurity}
```

The `Tag` field still works but will be removed in v1.0.

### 5. `FixApplier` — Deterministic file ordering

Fixes are now applied in sorted file-path order. Previously, map iteration order was non-deterministic. This is a behavioral change, not an API change. If your tests depended on a specific fix application order, they may need updating.

---

## New Features (Additive, No Migration Needed)

- `Finding.Validate() error` — Structural validation
- `Finding.Preview() string` — Unified-diff preview of BeforeCode/AfterCode
- `Finding.HasCategory() bool` — Convenience check
- `Finding.Tags []Tag` — Multi-tag support
- `Report.Filter(predicates ...FilterFunc) *Report` — Returns filtered copy
- `Report.Map(fn func(Finding) Finding) *Report` — Returns transformed copy
- `Report.WriteJSON(w io.Writer) error` — Streaming JSON output
- `Finding.WriteJSON(w io.Writer) error` — Single-finding JSON output
- `Builder.MustBuild() Finding` — Panic on invalid builder
- `Pipeline.CorrelateFindings` config option — Post-detection correlation
- CLI `-output` flag — Write to file
- CLI `-version` flag — Print version
- `finding.Version` constant — Programmatic version string

---

## Dependency Changes

No new production dependencies. `golang.org/x/tools` (12MB) still required by `diagnostic.go` for `go/analysis` integration.

## Minimum Go Version

Go 1.21+ (for `slices`, `maps` stdlib packages used internally).
