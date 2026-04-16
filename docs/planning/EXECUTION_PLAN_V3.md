# go-finding Execution Plan V3

**Date:** 2026-04-15
**Status:** Planning — Ready for Execution
**Predecessor:** EXECUTION_PLAN_V2.md (13/18 tasks completed)

---

## Part 1: Audit — What Was Forgotten, What Could Improve

### What Was Done Well
- Clean architecture with zero external deps for core types
- 96% test coverage in root package, 87.5% in pipeline
- Fuzz tests, property tests, benchmarks, integration tests
- SARIF 2.1.0 output, LSP diagnostic conversion
- Pipeline with retry, partial success, conflict detection, verification
- Multiple real detector examples (govet, staticcheck, artdupl, branching)

### What Was Forgotten / Could Improve

| # | Finding | Severity | Location |
|---|---------|----------|----------|
| 1 | CI matrix tests Go 1.21/1.22/1.23 but go.mod requires 1.26.0 | Critical | `.github/workflows/ci.yml` |
| 2 | `registerDetector` in CLI is dead code | Low | `cmd/go-finding/main.go` |
| 3 | `sevFromInt` duplicated in `bench_test.go` and `merge_fuzz_test.go` | Low | Test files |
| 4 | No `Equal` method on Finding, Position, Range | Medium | Core types |
| 5 | No `fmt.Stringer` on Severity, Category, FixStrategy | Low | Core types |
| 6 | `sha1Sum` for backup paths — should use sha256 | Low | `pipeline/pipeline.go` |
| 7 | No SARIF parsing (`FromSARIF`) — only generation exists | High | `sarif.go` |
| 8 | `Suppression.IsExpired` calls `time.Now()` — not deterministic | Medium | `suppression.go` |
| 9 | No configuration validation in CLI | Medium | `cmd/go-finding/main.go` |
| 10 | LSP conversion is lossy (loses FixStrategy, Confidence) | Medium | `lsp.go` |
| 11 | `IsHashID` checks `len(parts[2]) == 16` — fragile | Low | `id.go` |
| 12 | No streaming/iterator for large finding sets | Low | Core types |
| 13 | `Correlate` only works cross-tool, same-tool ignored | Medium | `merge.go` |
| 14 | FixApplier uses naive `strings.Replace` | Medium | `pipeline/pipeline.go` |
| 15 | `FindingError.WithFinding`/`WithPosition` clone, not mutate | Low | `errors.go` |
| 16 | 12+ gopls hints (rangeint, mapsloop, stringsseq, newexpr) | Low | Various |
| 17 | No `CategoryUnused` constant | Low | `category.go` |
| 18 | CLI flags hardcoded — would benefit from cobra | Medium | `cmd/go-finding/main.go` |
| 19 | No `go generate` support | Low | Build |
| 20 | Pipeline `detectBranchingIssues` is hardcoded stub | Low | `examples/branching/` |

### What Could Still Improve — Architecture

| # | Topic | Current State | Improvement |
|---|-------|---------------|-------------|
| A | Type safety | `Severity string`, `Category string` | Consider typed enums with compile-time guarantees |
| B | Value objects | Position/Range are structs with methods | Add `Equal`, `String`, `Compare` methods |
| C | Immutability | Finding is mutable struct | Consider builder pattern or functional options |
| D | Error wrapping | `FindingError` with categories | Add `errors.Is`/`errors.As` support via sentinel values |
| E | Collections | `[]Finding` everywhere | Consider `Findings` collection type with methods |
| F | Serialization | Manual JSON marshal | Consider `encoding/json` custom marshalers |
| G | Pipeline API | `*Pipeline` with methods | Consider functional options for configuration |

### Library Recommendations

| Library | Purpose | Assessment | Recommendation |
|---------|---------|------------|----------------|
| `github.com/owenrumney/go-sarif` | SARIF parsing + generation | Mature, well-maintained, SARIF schema compliance | Adopt for `FromSARIF`; keep custom `ToSARIF` for zero-dep core |
| `github.com/spf13/cobra` | CLI framework | Industry standard for Go CLIs | Adopt for `cmd/go-finding` |
| `github.com/stretchr/testify` | Test assertions | Already in scope | Consider for cleaner test code |
| `github.com/google/uuid` | ID generation | RFC 4122 UUIDs | Evaluate for `GenerateID` alternative |
| `github.com/bmatcuk/doublestar` | Glob pattern matching | For file filtering in pipeline | Low priority |
| `github.com/fsnotify/fsnotify` | File watching | For watch mode | Low priority |
| `golang.org/x/exp/slices` | Slice utilities | `slices.Equal`, `slices.Sort` | Use for Finding comparison |

---

## Part 2: Execution Plan

### Prioritization

Tasks sorted by: **(Impact × Urgency) / Effort** — highest ROI first.

Each task is self-contained and commit-ready.

---

### Tier 1: Quick Wins (Low Effort, High Impact)

---

#### T1-01: Fix CI Go Version Matrix

**Why:** CI is broken — tests Go 1.21-1.23 but module requires 1.26.0.

**Files:** `.github/workflows/ci.yml`

**Change:** Update matrix from `[1.21, 1.22, 1.23]` to `[1.24, 1.25, 1.26]`

**Verify:** `cat .github/workflows/ci.yml` (CI will verify on push)

---

#### T1-02: Fix gopls Hints (rangeint, mapsloop, stringsseq, newexpr)

**Why:** 12+ linting hints indicate code that can be modernized for Go 1.22+.

**Files:**
- `bench_test.go` — range over int
- `pipeline/metrics.go` — maps.Copy
- `examples/staticcheck/main.go` — stringsseq
- `coverage_test.go` — newexpr

**Change:** Apply gopls-suggested modernizations.

**Verify:** `just lint && go build ./...`

---

#### T1-03: Extract Duplicated `sevFromInt` to testutil

**Why:** Same helper defined in two test files.

**Files:** `bench_test.go`, `merge_fuzz_test.go`, `testutil.go`

**Change:** Move `sevFromInt` to `testutil.go`, reference from both files.

**Verify:** `go test ./...`

---

#### T1-04: Add `Equal` Methods to Position, Range, Finding

**Why:** No way to compare value equality. Consumers need this for assertions.

**Files:** `position.go`, `finding.go`, plus new test files

**Change:**
```go
func (p Position) Equal(other Position) bool
func (r Range) Equal(other Range) bool
func (f Finding) Equal(other Finding) bool
```

**Verify:** `go test ./...`

---

#### T1-05: Add `fmt.Stringer` to Severity, Category, FixStrategy

**Why:** Debugging and logging is harder without string representation.

**Files:** `severity.go`, `category.go`, `fix_strategy.go`, plus tests

**Change:** Add `String() string` methods that return the string value.

**Verify:** `go test ./...`

---

#### T1-06: Replace `sha1Sum` with `sha256` for Backup Paths

**Why:** SHA-1 is deprecated for cryptographic use. Even for non-crypto paths, consistency matters.

**Files:** `pipeline/pipeline.go`

**Change:** Replace `crypto/sha1` with `crypto/sha256`, update `sha1Sum` → `sha256Sum`.

**Verify:** `go test ./pipeline/...`

---

#### T1-07: Make `Suppression.IsExpired` Accept `now time.Time` Parameter

**Why:** `time.Now()` makes testing non-deterministic. Accept `now` parameter, default to `time.Now()`.

**Files:** `suppression.go`, `suppression_test.go` (if exists)

**Change:**
```go
func (s Suppression) IsExpired(now time.Time) bool
```

**Verify:** `go test ./...`

---

#### T1-08: Clean Up Dead `registerDetector` Code

**Why:** Dead code in CLI. Either remove or document intent.

**Files:** `cmd/go-finding/main.go`

**Change:** Remove unused `registerDetector` and related scaffolding, or convert to documented placeholder with TODO.

**Verify:** `go build ./cmd/go-finding/...`

---

### Tier 2: Medium Effort, High Impact

---

#### T2-01: Add SARIF Parsing (`FromSARIF`)

**Why:** Can generate SARIF but can't read it. Interoperability requires both.

**Files:** `sarif.go`, `sarif_test.go`

**Approach:** Evaluate `github.com/owenrumney/go-sarif/v2` for parsing. If adopted, keep `ToSARIF` in core (zero deps) but add `FromSARIF` that optionally uses the library.

**Change:**
```go
func FromSARIF(data []byte) (*Report, error)
func ReportFromSARIF(data []byte) (*Report, error)
```

**Verify:** `go test ./...` with real SARIF files as test fixtures.

---

#### T2-02: Add Configuration Validation

**Why:** YAML config parsed but never validated. Invalid configs fail silently.

**Files:** `cmd/go-finding/main.go` (or new `config.go`)

**Change:** Add `Validate()` method on config struct. Validate: max iterations > 0, detector names non-empty, severity thresholds valid.

**Verify:** `go test ./cmd/...`

---

#### T2-03: Improve LSP Conversion — Preserve Lost Data

**Why:** `FromLSP` loses FixStrategy, Confidence, Category, Related.

**Files:** `lsp.go`

**Change:** Store lost data in `Properties` map (SARIF-style metadata). Add `LSPMetadata` struct.

**Verify:** `go test ./...`

---

#### T2-04: Add `Compare` Method to Position and Range

**Why:** Enables sorting findings by location.

**Files:** `position.go`

**Change:**
```go
func (p Position) Compare(other Position) int  // -1, 0, +1
func (r Range) Compare(other Range) int
```

**Verify:** `go test ./...`

---

#### T2-05: Adopt `cobra` for CLI

**Why:** Flag-based CLI doesn't scale. Need subcommands, help, completion.

**Files:** `cmd/go-finding/main.go`

**Change:** Refactor to cobra with `run`, `convert`, `list` subcommands.

**Verify:** `go build ./cmd/go-finding/... && ./go-finding --help`

---

#### T2-06: Harden `IsHashID` — Remove Magic Number

**Why:** `len(parts[2]) == 16` is fragile. Hash length is an implementation detail.

**Files:** `id.go`

**Change:** Extract hash length as named constant. Add `hashPrefix` marker to distinguish hash IDs.

**Verify:** `go test ./...`

---

#### T2-07: Improve FixApplier — Line-Based Replacement

**Why:** `strings.Replace` is fragile. Replace with line-range-based replacement.

**Files:** `pipeline/pipeline.go`, `pipeline/astfix.go`

**Change:** Use finding Range to extract line range, replace lines instead of substring.

**Verify:** `go test ./pipeline/...`

---

### Tier 3: Higher Effort, Medium-High Impact

---

#### T3-01: Add `Findings` Collection Type

**Why:** `[]Finding` everywhere with repeated filter/sort/group logic. A collection type encapsulates this.

**Files:** New `findings.go`, `findings_test.go`

**Change:**
```go
type Findings []Finding

func (fs Findings) Filter(pred func(Finding) bool) Findings
func (fs Findings) SortBy(severity) Findings
func (fs Findings) GroupByFile() map[string][]Finding
func (fs Findings) Contains(f Finding) bool
```

**Verify:** `go test ./...`

**Note:** Evaluate if this replaces `filter.go` or wraps it.

---

#### T3-02: Add Sentinel Errors with `errors.Is`/`errors.As` Support

**Why:** `FindingError` categories are strings, not usable with `errors.Is`.

**Files:** `errors.go`

**Change:** Add sentinel errors:
```go
var (
    ErrValidation = errors.New("finding: validation error")
    ErrIO         = errors.New("finding: I/O error")
    ErrParse      = errors.New("finding: parse error")
    ErrConflict   = errors.New("finding: conflict error")
    ErrInternal   = errors.New("finding: internal error")
)
```

Make `FindingError` wrap these for `errors.Is` support.

**Verify:** `go test ./...`

---

#### T3-03: Add Streaming/Iterator for Large Finding Sets

**Why:** Processing millions of findings loads them all into memory.

**Files:** New `iterator.go`, `iterator_test.go`

**Change:**
```go
type FindingIterator interface {
    Next() bool
    Finding() Finding
    Err() error
}

func NewFileIterator(path string) FindingIterator
func NewSARIFIterator(r io.Reader) FindingIterator
```

**Verify:** `go test ./...`

---

#### T3-04: Improve Correlate — Same-Tool Same-File Support

**Why:** Correlate currently only works cross-tool. Same-tool findings in same file are ignored.

**Files:** `merge.go`

**Change:** Add option to correlate within same tool. Update `CorrelateOptions` struct.

**Verify:** `go test ./...`

---

#### T3-05: Add `CategoryUnused` Constant

**Why:** Category has 13 constants but no `CategoryUnused`. Symmetry with `FixStrategyNone`.

**Files:** `category.go`

**Change:** Add `CategoryUnused Category = "unused"`.

**Verify:** `go test ./...`

---

### Tier 4: Lower Priority / Future Work

---

#### T4-01: Evaluate `go-sarif` Library Integration

**Why:** EXECUTION_PLAN_V2 deferred this. Re-evaluate for T2-01.

**Action:** Research `github.com/owenrumney/go-sarif/v2`. If adopted, scope to `internal/sarif` package.

---

#### T4-02: Add Watch Mode

**Why:** Continuous analysis during development.

**Dependencies:** T2-05 (cobra CLI), `fsnotify`

---

#### T4-03: Add `go generate` Support

**Why:** Boilerplate like String() methods could be generated.

**Action:** Evaluate `stringer` for enum types, `enumerable` for JSON methods.

---

#### T4-04: Property-Based Tests for New Methods

**Why:** New `Equal`, `Compare` methods should have property tests.

**Dependencies:** T1-04, T2-04

---

#### T4-05: Builder Pattern for Finding

**Why:** Finding construction is verbose with many fields.

**Change:**
```go
func NewFinding(tool, rule string) *FindingBuilder
// builder.WithFile(f).WithSeverity(s).WithMessage(m).Build()
```

---

## Part 3: Execution Order

Recommended execution sequence (approximately):

| Order | Task | Tier | Rationale |
|-------|------|------|-----------|
| 1 | T1-01: Fix CI matrix | Quick Win | CI is currently broken |
| 2 | T1-02: Fix gopls hints | Quick Win | Code modernization |
| 3 | T1-03: Extract sevFromInt | Quick Win | DRY |
| 4 | T1-04: Add Equal methods | Quick Win | Core API completeness |
| 5 | T1-05: Add Stringer methods | Quick Win | Debugging support |
| 6 | T1-06: sha1→sha256 | Quick Win | Security consistency |
| 7 | T1-07: IsExpired determinism | Quick Win | Testability |
| 8 | T1-08: Clean dead code | Quick Win | Maintenance |
| 9 | T2-06: Harden IsHashID | Medium | Robustness |
| 10 | T3-05: Add CategoryUnused | Medium | API completeness |
| 11 | T2-04: Add Compare methods | Medium | Sorting support |
| 12 | T3-02: Sentinel errors | Medium | Error handling |
| 13 | T2-02: Config validation | Medium | Robustness |
| 14 | T2-03: Improve LSP conversion | Medium | Data loss prevention |
| 15 | T2-07: Line-based FixApplier | Medium | Correctness |
| 16 | T2-01: SARIF parsing | Medium | Interoperability |
| 17 | T2-05: Adopt cobra CLI | Medium | UX |
| 18 | T3-01: Findings collection | Higher | Architecture |
| 19 | T3-04: Improve Correlate | Higher | Feature completeness |
| 20 | T3-03: Streaming iterator | Higher | Scalability |
| 21 | T4-03: go generate | Lower | Automation |
| 22 | T4-04: Property tests | Lower | Coverage |
| 23 | T4-05: Builder pattern | Lower | Ergonomics |
| 24 | T4-01: go-sarif evaluation | Lower | Dependencies |
| 25 | T4-02: Watch mode | Lower | Features |

---

## Success Criteria

- [ ] CI passes with correct Go versions
- [ ] All gopls hints resolved
- [ ] Core types have `Equal`, `String`, `Compare` methods
- [ ] Pipeline backup uses sha256
- [ ] Suppression expiry is testable
- [ ] Dead code removed
- [ ] SARIF parsing available
- [ ] Configuration validated
- [ ] CLI uses cobra
- [ ] All tests pass, coverage ≥ 95%

---

*Plan created: 2026-04-15*
*Based on: Full codebase audit of all source, test, config, and documentation files*
