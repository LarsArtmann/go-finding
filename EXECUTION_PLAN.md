# go-finding Execution Plan

**Last Updated:** 2026-04-12  
**Status:** Phase 1 Complete, Phase 2 Ready

---

## ✅ COMPLETED

### Phase 1: Foundation
- [x] Core types (Finding, Severity, FixStrategy, Position, Range, Suppression, Report)
- [x] Utilities (filter, merge, id, json)
- [x] Integrations (SARIF, LSP, go/analysis)
- [x] CI/CD (GitHub Actions)
- [x] Documentation (README, CHANGELOG, AGENTS.md, doc.go)
- [x] Result[T] type for ergonomic error handling
- [x] Justfile for common tasks
- **Tests:** 34 passing

---

## 🔄 NEXT: Phase 2 - Pipeline Engine

**Status:** ⏸️ **BLOCKED - Need user input on architecture**

### The Question

Pipeline stages have different input/output types:

| Stage | Input | Output |
|-------|-------|--------|
| Detect | `[]Detector` | `[]Finding` |
| Triage | `[]Finding` | `TriageResult{Direct, Suggest, None}` |
| Fix | `TriageResult` | `FixResult{Applied, Failed}` |
| Verify | `[]string` (modified files) | `VerificationResult` |

**Which design?**

| Option | Approach | Pros | Cons |
|--------|----------|------|------|
| **A** | `Stage[In, Out any]` interface | Type-safe | Complex |
| **B** | Separate concrete stages | Clear, simple | Rigid |
| **C** | Functional: `func(ctx, input) (output, error)` | Simple | Less structure |
| **D** | `any` with type assertions | Flexible | Runtime errors |

**My recommendation: Option C (Functional)**

```go
// Simple, composable, clear
func Detect(ctx context.Context, detectors []Detector) ([]Finding, error)
func Triage(ctx context.Context, findings []Finding) (*TriageResult, error)
func Fix(ctx context.Context, triage *TriageResult) (*FixResult, error)
func Verify(ctx context.Context, files []string) (*VerifyResult, error)

// Pipeline orchestrates
pipeline.Run(ctx) // Calls stages in sequence
```

**Waiting for your decision before proceeding.**

---

## 📋 Phase 3: Established Libraries (Pending)

- [ ] Evaluate `github.com/owenrumney/go-sarif`
- [ ] SARIF schema validation tests

## 📋 Phase 4: Testing & Examples (Pending)

- [ ] Benchmark tests
- [ ] Fuzz tests
- [ ] Tool converter examples

## 📋 Phase 5: Integration (Pending)

- [ ] Real project validation
- [ ] Performance tuning

---

## Immediate Actions (No Pipeline)

If you want me to continue without the pipeline:

1. ✅ Add benchmarks (next)
2. ✅ Add fuzz tests
3. ✅ Fix Range type (pointer → value)
4. ✅ Create simple CLI example
5. ✅ Push current state

**Current commits ahead of origin: 3**
