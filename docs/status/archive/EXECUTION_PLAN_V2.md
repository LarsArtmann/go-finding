# go-finding Comprehensive Execution Plan V2

**Date:** 2026-04-13
**Status:** Phase 1–3 Complete (Tasks 1–13) — Phase 4 Remaining
**Goal:** Production-ready library with real-world tool integrations

---

## Executive Summary

This plan addresses the gaps identified in the status report and prioritizes work by impact vs effort. Each task is designed to be completed in ≤12 minutes, enabling rapid iteration.

---

## Prioritization Matrix

| Priority    | Impact | Effort | Tasks |
| ----------- | ------ | ------ | ----- |
| 🔴 Critical | High   | Low    | 1-5   |
| 🟠 High     | High   | Medium | 6-10  |
| 🟡 Medium   | Medium | Medium | 11-14 |
| 🟢 Low      | Low    | High   | 15-20 |

---

## Phase 1: Immediate Fixes (Critical)

### Task 1: Fix Pipeline Linting Issues

**Time:** ~8 min\
**Impact:** Code quality, prevents technical debt\
**Why:** Current code has linting hints that should be addressed

**Issues to fix:**

- `pipeline/pipeline.go:214` - Unnecessary loop variable copy (gopls forvar)
- `pipeline/pipeline.go:435` - Inefficient string concatenation (use strings.Builder)

**Verification:**

```bash
just lint
```

---

### Task 2: Create Structured Error Types

**Time:** ~10 min\
**Impact:** Better error handling for consumers\
**Why:** Current errors are plain fmt.Errorf, hard to programmatically handle

**Deliverables:**

```go
// errors.go
package finding

type ErrorCategory string

const (
    ErrCategoryValidation ErrorCategory = "validation"
    ErrCategoryIO         ErrorCategory = "io"
    ErrCategoryParse      ErrorCategory = "parse"
    ErrCategoryConflict   ErrorCategory = "conflict"
)

type FindingError struct {
    Category ErrorCategory
    Finding  Finding
    Message  string
    Cause    error
}

func (e *FindingError) Error() string
func (e *FindingError) Unwrap() error
```

**Verification:** Add test in `errors_test.go`

---

### Task 3: Implement Fix Conflict Detection

**Time:** ~12 min\
**Impact:** Prevents broken code from overlapping fixes\
**Why:** The #1 question in status report - how to handle overlapping fixes

**Approach:** Group by location, detect overlaps using Position/Range

```go
// conflict.go
package pipeline

// FixGroup represents fixes that can be safely applied together
type FixGroup struct {
    File   string
    Fixes  []finding.Finding
    Bounds finding.Range
}

// DetectConflicts returns groups of non-conflicting fixes
func DetectConflicts(fixes []finding.Finding) ([]FixGroup, []finding.Finding)
```

**Strategy:** Group fixes by file, sort by position, detect overlaps

**Verification:** Unit tests with overlapping and non-overlapping scenarios

---

### Task 4: Add Position/Range Overlap Methods

**Time:** ~8 min\
**Impact:** Enables conflict detection, improves type model\
**Why:** Current Range only has Contains(), needs Overlaps(), Intersects()

```go
// position.go additions

// Overlaps returns true if this range overlaps with other
func (r Range) Overlaps(other Range) bool

// Intersection returns the overlapping region, or nil if none
func (r Range) Intersection(other Range) *Range

// Adjacent returns true if ranges are adjacent (touching but not overlapping)
func (r Range) Adjacent(other Range) bool
```

**Verification:** Add tests in `position_test.go`

---

### Task 5: Fix Applier Error Structuring

**Time:** ~10 min\
**Impact:** Better error reporting from FixApplier\
**Why:** Currently logs failures but doesn't provide structured errors

```go
// pipeline/fixerros.go

type FixResult struct {
    Finding finding.Finding
    Applied bool
    Error   error
}

type ApplyResult struct {
    Total    int
    Applied  int
    Failed   int
    Skipped  int // Conflicts detected
    Results  []FixResult
}

func (a *FixApplier) ApplyWithResult(ctx context.Context, fixes []finding.Finding) (*ApplyResult, error)
```

---

## Phase 2: Core Improvements (High Priority)

### Task 6: Implement AST-Aware Fix Application

**Time:** ~12 min\
**Impact:** Robust code transformations\
**Why:** Current text replacement is fragile

**Use standard library:** `go/parser`, `go/ast`, `go/printer`

```go
// astfix.go
package pipeline

import (
    "go/ast"
    "go/parser"
    "go/printer"
    "go/token"
)

// ASTFixer applies fixes using AST transformations
type ASTFixer struct {
    fset *token.FileSet
}

func NewASTFixer() *ASTFixer

// Apply attempts AST-aware fix, falls back to text replacement
func (a *ASTFixer) Apply(file string, fix finding.Finding) error
```

**Fallback strategy:** Try AST first, fall back to text replacement if parsing fails

---

### Task 7: Create Go Vet Converter

**Time:** ~12 min\
**Impact:** Real tool integration\
**Why:** Need detector implementations that work with actual tools

**Deliverable:** Moved to `internal/detectors/govet.go` — exports `NewGoVetDetector(dir string) pipeline.Detector`

**Status:** ✅ Done (extracted to `internal/detectors/`)

---

### Task 8: Evaluate go-sarif Library

**Time:** ~10 min\
**Impact:** Decision on SARIF implementation\
**Why:** EXECUTION_PLAN.md mentions evaluating github.com/owenrumney/go-sarif

**Research:**

1. Check if go-sarif provides value over custom implementation
2. Compare features: validation, schema compliance, round-tripping
3. Check dependency cost

**Decision criteria:**

- Only adopt if it provides significant value
- Prefer keeping zero external deps for core types
- Consider as optional integration

---

### Task 9: Add Pipeline Metrics Collection

**Time:** ~12 min\
**Impact:** Performance insights, debugging\
**Why:** Track timing per stage, memory, detector performance

```go
// pipeline/metrics.go

type Metrics struct {
    StageDurations map[string]time.Duration
    DetectorTimes  map[string]time.Duration
    FindingsFound  map[string]int
    MemoryMB       float64
}

func (p *Pipeline) RunWithMetrics(ctx context.Context) (*Result, *Metrics, error)
```

---

### Task 10: Implement Verification Stage

**Time:** ~12 min\
**Impact:** Confirms fixes actually worked\
**Why:** Pipeline has verify step in theory but no implementation

```go
// pipeline/verify.go

// Verifier checks that fixes were applied correctly
type Verifier struct {
    detectors []Detector
}

func (v *Verifier) Verify(ctx context.Context, original []finding.Finding) (*VerifyResult, error)

type VerifyResult struct {
    Fixed     []finding.Finding // Confirmed fixed
    Remaining []finding.Finding // Still present
    New       []finding.Finding // New issues introduced
}
```

**Strategy:** Re-run detectors, compare findings

---

## Phase 3: Robustness (Medium Priority)

### Task 11: Add Fuzz Tests for Filter/Merge

**Time:** ~10 min\
**Impact:** Catches edge cases\
**Why:** Property-based testing for core operations

```go
// filter_fuzz_test.go
// merge_fuzz_test.go

func FuzzFilter(f *testing.F) {
    // Generate random findings
    // Apply random filters
    // Verify properties
}
```

---

### Task 12: Add Retry Logic with Exponential Backoff

**Time:** ~10 min\
**Impact:** Handles transient detector failures\
**Why:** External tools may fail intermittently

```go
// pipeline/retry.go

func detectWithRetry(d Detector, maxRetries int, baseDelay time.Duration) ([]finding.Finding, error)
```

**Use:** `golang.org/x/sync` (already in go.mod)

---

### Task 13: Implement Partial Success Handling

**Time:** ~12 min\
**Impact:** More resilient pipelines\
**Why:** Currently fails entire iteration if one detector errors

```go
// PartialResult contains findings and errors separately
type PartialResult struct {
    Findings []finding.Finding
    Errors   map[string]error // detector name -> error
}

func (p *Pipeline) detectPartial(ctx context.Context) *PartialResult
```

---

### Task 14: Create CLI Tool

**Time:** ~12 min\
**Impact:** Usable standalone tool\
**Why:** Make the library accessible without code

```go
// cmd/go-finding/main.go

// Commands:
//   go-finding run [path]      # Run pipeline
//   go-finding convert [file]  # Convert SARIF/other formats
//   go-finding list            # List available detectors
```

---

## Phase 4: Advanced Features (Low Priority)

### Task 15: Configuration File Support

**Time:** ~12 min\
**Impact:** Flexible configuration\
**Why:** YAML/JSON config for pipelines

```go
// config.go

type PipelineConfig struct {
    MaxIterations     int              `yaml:"maxIterations"`
    Parallel          bool             `yaml:"parallel"`
    Detectors         []DetectorConfig `yaml:"detectors"`
    SeverityThreshold Severity         `yaml:"severityThreshold"`
}
```

---

### Task 16: Add Watch Mode

**Time:** ~12 min\
**Impact:** Continuous analysis during development\
**Why:** File system monitoring for continuous analysis

Use `fsnotify` or poll-based approach

---

### Task 17: Create Web UI Prototype

**Time:** >12 min\
**Impact:** Visual pipeline monitoring\
**Why:** Progress visualization

**Defer:** Larger task, create separate project

---

### Task 18: Add Property-Based Tests

**Time:** ~12 min\
**Impact:** Mathematical confidence\
**Why:** Quick-check style testing

Use `testing/quick` from stdlib or `gopter`

---

## Summary Table

| #  | Task                    | Time | Impact | Effort | Status      |
| -- | ----------------------- | ---- | ------ | ------ | ----------- |
| 1  | Fix linting issues      | 8m   | High   | Low    | ✅ Done     |
| 2  | Structured error types  | 10m  | High   | Low    | ✅ Done     |
| 3  | Fix conflict detection  | 12m  | High   | Medium | ✅ Done     |
| 4  | Position/Range methods  | 8m   | High   | Low    | ✅ Done     |
| 5  | FixApplier error struct | 10m  | High   | Low    | ✅ Done     |
| 6  | AST-aware fixes         | 12m  | High   | Medium | ✅ Done     |
| 7  | Go vet converter        | 12m  | High   | Medium | ✅ Done     |
| 8  | Evaluate go-sarif       | 10m  | Medium | Low    | ⬜ Deferred |
| 9  | Metrics collection      | 12m  | Medium | Medium | ✅ Done     |
| 10 | Verification stage      | 12m  | High   | Medium | ✅ Done     |
| 11 | Fuzz tests              | 10m  | Medium | Low    | ✅ Done     |
| 12 | Retry logic             | 10m  | Medium | Low    | ✅ Done     |
| 13 | Partial success         | 12m  | Medium | Medium | ✅ Done     |
| 14 | CLI tool                | 12m  | High   | Medium | ⬜ Pending  |
| 15 | Config file support     | 12m  | Low    | Medium | ⬜ Pending  |
| 16 | Watch mode              | 12m  | Low    | Medium | ⬜ Pending  |
| 17 | Web UI                  | >12m | Low    | High   | ⬜ Deferred |
| 18 | Property tests          | 12m  | Low    | Medium | ⬜ Pending  |

---

## Dependencies to Evaluate

| Library                        | Purpose        | Current Status    |
| ------------------------------ | -------------- | ----------------- |
| github.com/owenrumney/go-sarif | SARIF parsing  | Evaluate Task 8   |
| github.com/fsnotify/fsnotify   | File watching  | Task 16           |
| github.com/spf13/cobra         | CLI framework  | Task 14           |
| github.com/BurntSushi/toml     | Config parsing | Task 15           |
| gopkg.in/yaml.v3               | Config parsing | Already available |

---

## Success Criteria

1. ✅ All linting issues resolved
2. ✅ Pipeline can detect and handle fix conflicts
3. ✅ Real tool integration example (go vet)
4. ✅ Structured error handling throughout
5. ✅ Verification stage implemented
6. ✅ Metrics collection available
7. ⬜ CLI tool usable (Phase 4)
8. ✅ High test coverage with fuzz tests

---

## Completion Notes (2026-04-13)

**Completed: 13/18 tasks** (Phase 1–3 fully done)

Key deliverables shipped:

- Structured errors with 5 categories (validation, io, parse, conflict, internal)
- Conflict detection with range-based overlap analysis
- Verification stage with ID-based finding diffing
- Metrics collection with nil-safe stage timing
- RetryDetector with configurable exponential backoff
- Partial success detection (DetectPartial)
- Fuzz tests for Filter, Merge, Correlate, DedupKey
- Go vet detector in `internal/detectors/govet.go`

Remaining work (Phase 4): CLI tool, config files, watch mode, Web UI, property tests.
These are lower priority and can be addressed as the project evolves.

---

_Plan created: 2026-04-13_
_Last updated: 2026-04-13 — Phase 1–3 complete_
