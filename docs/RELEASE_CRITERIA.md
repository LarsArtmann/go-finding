# v1.0.0 Release Criteria

## Must Have

- [x] Core types stable: `Finding`, `Report`, `Position`, `Range`, `Severity`, `Confidence`, `Category`, `Tag`, `FixStrategy`, `Suppression`
- [x] No `Properties map[string]any` — `Metadata map[string]string` only
- [x] SARIF 2.1.0 full round-trip (export + import)
- [x] LSP Diagnostic conversion
- [x] go/analysis integration
- [x] Pipeline: detect → triage → fix → verify loop
- [x] Byte-level FixEngine with provider chain
- [x] Conflict detection (position + byte-level)
- [x] Report merging with deduplication (by ID, position, rule)
- [x] Cross-tool correlation
- [x] Diff (before/after comparison)
- [x] 95%+ test coverage on core packages
- [x] Zero lint warnings
- [x] Race detector clean
- [x] Comprehensive godoc examples (22+)
- [x] Structured errors with category classification
- [x] Builder pattern for Finding construction
- [x] Context cancellation throughout pipeline
- [x] Per-detector timeouts
- [x] Graceful degradation on detector failures
- [x] Generated file filtering (gogenfilter)
- [x] Customizable TriageFunc
- [x] ByteLevelConflictDetection opt-in

## Should Have

- [ ] CI/CD pipeline (GitHub Actions)
- [ ] API stability guarantee document
- [ ] Real-world tool integration guide
- [ ] Finding JSON schema
- [ ] USAGE_GUIDE.md complete for v0.3.0+
- [ ] README.md with pipeline examples

## Nice to Have

- [ ] Spatial index for Correlate (performance)
- [ ] Streaming merge (memory optimization)
- [ ] Plugin architecture for external detectors
- [ ] FixEngine line-offset tracking

## Release Blockers

None currently. Code is stable, tests pass, lint is clean.

## Minimum Bar for v1.0.0

1. All "Must Have" items complete
2. No breaking changes after release without major version bump
3. CI/CD running green
4. API stability guarantee published
