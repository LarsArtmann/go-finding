# Status Report — go-finding

**Date:** 2026-05-03 | **Version:** 0.2.1 | **Branch:** master

---

## 1. Fully Done

| Area | What | Evidence |
|------|------|----------|
| Critical bug fix | Report mutex was a no-op (`lock()` with `defer Unlock()` inside) | Commit `bda5ae1` — split into `lock()`/`unlock()`, added lock to `ComputeSummary()` |
| FindingProcessor | Interface, ProcessorFunc adapter, NamedProcessorFunc, Config.Processors, wired into Run loop | Commit `e2ee28b` — interface + pipeline integration |
| FindingProcessor tests | 3 BDD specs: adapter, named adapter, chaining | Commit `91bf42c` |
| Type model consistency | Tag now has `IsStandard()`, `IsValid()`, `String()` mirroring Category | Commit `aa2f05c` |
| Suppression.IsActive | Convenience combining IsValid + !IsExpired | Commit `aa2f05c` |
| Deprecated marker | `WithTag()` has `// Deprecated: Use WithTags instead.` godoc | Commit `d2f014b` |
| BDD test framework | ginkgo/v2 + gomega added, 29 total BDD specs (23 root + 6 pipeline) | Commit `5d669cb` |
| Doc bugs fixed | integration-guide compile-breaking example, USAGE_GUIDE wrong field names | Commits `a73358a`, `4978bf7` |
| FEATURES.md | FindingProcessor section 16.4, Tag/Suppression method docs, renumbered subsections | Commit `b849a77` |
| AGENTS.md | FindingProcessor, tag.go, Suppression.IsActive | Commit `3f15d38` |
| Architecture diagrams | Current + ideal mermaid diagrams in docs/architecture-understanding/ | Commit `5d669cb` |
| Depguard config | ginkgo/gomega in test allowlist | Commit `5d669cb` |
| go.work version | Fixed from 1.26.0 to 1.26.2 | Commit `5d669cb` |

## 2. Partially Done

| Area | Status | Gap |
|------|--------|-----|
| Tag method tests | 0% coverage | `IsStandard()`, `IsValid()`, `String()` have no direct tests |
| BDD test coverage | 29 specs exist | Some edge cases not covered (e.g., Suppression.IsActive) |

## 3. Not Started

| Area | Priority | Notes |
|------|----------|-------|
| Confidence strong type | Low | Currently `float64`; breaking API change |
| FixApplier interface extraction | Medium | Would enable custom fix strategies; breaking |
| DetectorRegistry | Medium | Dynamic detector registration in pipeline package |
| go/analysis in-process detectors | High (v0.3) | Avoid subprocess overhead for govet |
| slog migration for CLI | Low | Would break stderr-capturing tests |
| go-sarif library adoption | Deferred | Hand-rolled SARIF is better integrated |

## 4. Issues Found

| Issue | Severity | Status |
|-------|----------|--------|
| Report mutex was a no-op | CRITICAL | Fixed (`bda5ae1`) |
| Tag methods at 0% coverage | Medium | Needs tests |
| LSP typecheck warnings in bdd_test.go | Low | Stale LSP cache; tests pass, golangci-lint clean |
| integration-guide had compile-breaking code | Medium | Fixed (`a73358a`) |
| USAGE_GUIDE had wrong MetricsSnapshot fields | Medium | Fixed (`a73358a`) |

## 5. Metrics

| Metric | Value |
|--------|-------|
| Total coverage | 95.5% |
| Core (finding) | 98.9% |
| Pipeline | 98.0% |
| CLI | 95.4% |
| Detectors | 96.1% |
| Source lines | ~5,989 |
| Test lines | ~14,675 |
| Test-to-code ratio | 2.45:1 |
| BDD specs | 29 (23 root + 6 pipeline, now 9) |
| Lint issues | 0 |
| Race detector | Clean |
| Dependencies | golang.org/x/tools, golang.org/x/sync, gopkg.in/yaml.v3, ginkgo/v2 (test), gomega (test) |

## 6. Session Commits (10)

```
3f15d38 docs(AGENTS): add FindingProcessor, Tag, Suppression updates
b849a77 docs(FEATURES): add FindingProcessor section, Tag/Suppression methods
91bf42c test(pipeline): add FindingProcessor BDD specs, fix lint warning
aa2f05c feat(finding): add Tag.IsStandard/IsValid/String and Suppression.IsActive
a73358a fix(docs): compile-breaking examples in integration-guide and USAGE_GUIDE
4978bf7 docs: update TODO_LIST, CHANGELOG, architecture-decisions, USAGE_GUIDE
e2ee28b feat(pipeline): add FindingProcessor interface for composable finding transforms
d2f014b docs(finding): add Deprecated godoc marker to WithTag builder method
5d669cb feat: add BDD tests (ginkgo), fix stale docs, update architecture diagrams
bda5ae1 fix(report): critical — lock() was releasing mutex immediately via defer
```

## 7. Top 25 Next Items (by impact)

### High Impact, Low Effort

1. **Add Tag method tests** — IsStandard/IsValid/String at 0% coverage
2. **Add Suppression.IsActive tests** — convenience method untested
3. **BDD tests for root package Tag/Suppression** — extend existing ginkgo suite
4. **FixApplier test coverage** — verify edge cases in string replacement
5. **CLI integration test for SARIF output** — validate end-to-end SARIF

### High Impact, Medium Effort

6. **go/analysis in-process detectors** — avoid subprocess overhead
7. **FixApplier interface extraction** — enable custom fix strategies
8. **DetectorRegistry in pipeline** — dynamic registration
9. **Confidence strong type** — replace float64 with type-safe wrapper
10. **Cross-tool correlation improvements** — semantic analysis beyond heuristics

### Medium Impact, Low Effort

11. **Benchmark FindingProcessor** — measure processor chain overhead
12. **Example_test.go for FindingProcessor** — runnable example in docs
13. **FixEngine fuzz tests** — stress test line-based replacement
14. **Pipeline config validation tests** — cover all invalid config paths
15. **Tag constants audit** — ensure all standard tags have test coverage

### Medium Impact, Medium Effort

16. **Stream-based SARIF output** — lazy generation for large reports
17. **Pipeline iteration history** — store finding diffs per iteration
18. **Report.Merge benchmarks** — performance with 10K+ findings
19. **CLI progress output** — real-time detection progress
20. **Config file schema validation** — strict YAML/JSON schema

### Lower Priority

21. **slog migration for CLI** — modern logging, but breaks tests
22. **go-sarif library evaluation** — revisit if it improves
23. **OpenTelemetry integration** — pipeline tracing
24. **Plugin system for fix strategies** — extensible fix engines
25. **WASM compilation target** — browser-based analysis

## 8. Top Question

**Should FindingProcessor be promoted from EXPERIMENTAL to STABLE before v0.3.0?**

The interface is simple, well-tested (BDD specs), and follows established patterns (DetectorFunc adapter). However, it has zero external consumers yet. Recommendation: keep EXPERIMENTAL until at least one real-world processor is built against it.

---

_Assisted-by: Crush <crush@charm.land>_
