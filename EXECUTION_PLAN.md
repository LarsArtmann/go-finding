# go-finding Execution Plan

## Phase 1: Foundation (Immediate - High Impact, Low Effort)

### 1.1 Add Missing Standard Files
- [ ] CHANGELOG.md
- [ ] .github/workflows/ci.yml
- [ ] Package documentation (doc.go)

### 1.2 Fix go.mod
- [ ] Lower Go version to 1.21 for broader compatibility
- [ ] Add module documentation

### 1.3 Improve Error Handling
- [ ] Add error returns to ID generation functions
- [ ] Add validation errors for Finding construction
- [ ] Add Result[T] type using samber/mo

**Effort:** 2 hours | **Impact:** High | **Priority:** P0

---

## Phase 2: Pipeline Engine (High Impact, Medium Effort)

### 2.1 Core Pipeline Types
- [ ] Detector interface (tools implement this)
- [ ] Stage interface (detect, triage, fix, verify)
- [ ] Pipeline struct with configuration
- [ ] Context cancellation support

### 2.2 Detect Stage
- [ ] Run multiple detectors in parallel
- [ ] Collect results with timeout
- [ ] Error aggregation

### 2.3 Triage Stage
- [ ] Route by FixStrategy
- [ ] Batch by file for efficiency
- [ ] Priority queue for AI fixes

### 2.4 Fix Stage
- [ ] Direct fix application (file writer with backup)
- [ ] AI fix routing (prompt builder)
- [ ] Fix result tracking

### 2.5 Verify Stage
- [ ] Re-run detection on modified files
- [ ] Compare before/after
- [ ] Detect new issues introduced

### 2.6 Pipeline Loop
- [ ] Orchestrate detect → triage → fix → verify
- [ ] Configurable max iterations
- [ ] Progress reporting

**Effort:** 2-3 days | **Impact:** Very High | **Priority:** P0

---

## Phase 3: Established Libraries (Medium Impact, Low Effort)

### 3.1 SARIF
- [ ] Evaluate `github.com/owenrumney/go-sarif`
- [ ] Replace custom SARIF types or use alongside
- [ ] Add schema validation tests

### 3.2 LSP
- [ ] Evaluate `github.com/sourcegraph/go-lsp`
- [ ] Decide: use library types or keep custom

**Effort:** 1 day | **Impact:** Medium | **Priority:** P1

---

## Phase 4: Testing & Quality (High Impact, Medium Effort)

### 4.1 Test Improvements
- [ ] Add SARIF schema validation tests
- [ ] Add fuzz tests for ID parsing
- [ ] Add benchmark tests for filtering/merging
- [ ] Add integration test with real go/analysis analyzer

### 4.2 Examples
- [ ] Complete tool integration example (art-dupl-style)
- [ ] CLI tool example using the library
- [ ] Pipeline usage example

**Effort:** 1-2 days | **Impact:** High | **Priority:** P1

---

## Phase 5: Architecture Improvements (Medium Impact, High Effort)

### 5.1 Type System Enhancements
- [ ] FindingBuilder fluent API
- [ ] Generic Report[T] for tool-specific extensions
- [ ] Event system for pipeline observability

### 5.2 Performance
- [ ] Streaming Report processing for large codebases
- [ ] Parallel filtering
- [ ] Memory pooling for high-throughput scenarios

**Effort:** 2-3 days | **Impact:** Medium | **Priority:** P2

---

## Phase 6: Integration (Very High Impact, Medium Effort)

### 6.1 Create Converters for Existing Tools
- [ ] art-dupl → Finding converter
- [ ] branching-flow → Finding converter
- [ ] go-auto-upgrade → Finding converter

### 6.2 Validate with Real Projects
- [ ] Test on go-business-rules
- [ ] Test on index project
- [ ] Test on ActaFlow

**Effort:** 2-3 days | **Impact:** Very High | **Priority:** P1

---

## Sorted by Impact vs Effort

| Priority | Task | Effort | Impact |
|----------|------|--------|--------|
| P0 | Pipeline Engine | 2-3d | Very High |
| P0 | Missing Standard Files | 2h | High |
| P1 | Integration Examples | 2-3d | Very High |
| P1 | Established Libraries | 1d | Medium |
| P1 | Test Improvements | 1-2d | High |
| P2 | Architecture Improvements | 2-3d | Medium |

---

## Immediate Next Steps

1. Create CHANGELOG.md and CI workflow
2. Start Pipeline Engine implementation
3. Research go-sarif library
