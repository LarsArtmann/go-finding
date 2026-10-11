# Benchmark Baseline

Captured: 2026-09-23 (regenerated for the multi-edit `Finding.Edits` field, issue #36)
Environment: AMD RYZEN AI MAX+ 395 (32 threads), Linux, Go 1.27.1 with GOEXPERIMENT=jsonv2
Command (multi-module — plain `go test ./...` from the root covers only the core module):

```bash
export GOEXPERIMENT=jsonv2
go test -run='^$' -bench=. -benchmem -count=10 ./... > /tmp/current.txt
(cd pipeline && go test -run='^$' -bench=. -benchmem -count=10 ./... >> /tmp/current.txt)
bash scripts/bench-check.sh benchmarks/baseline.txt /tmp/current.txt 250 10
```

## Usage

Thresholds: allocation regressions (B/op or allocs/op) beyond **+10%** fail the
check; time regressions beyond **+250%** (3.5x) fail it. See the header of
`scripts/bench-check.sh` for why time is thermal-tolerant on this machine.

## Notes

- This baseline is environment-dependent. Use it for relative comparison on the same machine.
- 10 iterations per benchmark for statistical significance (benchstat requires >= 6 for p-values).
- Regenerate after intentional performance changes: `cp /tmp/current.txt benchmarks/baseline.txt`
- **2026-09-08 regeneration rationale (v1.7.0):** v1.7.0 added `GroupID` to the `Finding`
  struct and enriched the LSP round-trip (tag re-emission). Finding-copy benchmarks
  (`Clone`, `Correlate`, `MergeNoDedup`, `LSPRoundTrip`) pay a real CPU cost for the
  larger struct — with **identical allocation profiles** (B/op and allocs/op unchanged).
- **2026-09-23 regeneration rationale (issue #36, multi-edit fixes):** `Finding` gained the
  `Edits []TextEdit` field (+24-byte slice header per copy). Unlike `GroupID`, it does not
  fit existing padding, so Finding-copy benchmarks pay a real **B/op** cost (+11%..+25% on
  `GroupByFile`, `Correlate`, `MergeIter`/`Combine`) with **flat allocs/op** (+0..+3%).
  Time is unchanged except noise. An intermediate full-equality O(n²) applied-findings
  dedup briefly regressed `GoAST_1000` +706% and applier benches +288..332% — replaced by
  O(1) owner-index tracking (`editOwner`) before this baseline was captured; the engine's
  `Applied` stays per finding with zero measurable cost.
- **2026-09-08 evening finding (thermal noise, gate redesign):** on identical code,
  `IntervalIndex_Query/1000` measured 9.3µ in a fresh filtered run but 20.1µ late in a
  sustained full-suite run (n=10, p=0.000 both) — late-suite benchmarks are thermally
  depressed on this APU, and cross-session swings of -9%..+150% were observed (worse than
  the -9%/+86% same-day swings documented at v1.7.0). A 25% time threshold is therefore a
  random-failure gate here. The gate now treats deterministic allocation metrics as the
  primary signal (+10% hard fail) and only fails time beyond 3.5x. When you need
  fine-grained CPU comparisons, run a fresh filtered benchmark (`-bench='Name'`) against
  the suspect change instead of comparing full-suite captures across sessions.
- `bench-check.sh` resolves benchstat via the PATH, else the pinned Go tool directive
  (`go.mod: tool golang.org/x/perf/cmd/benchstat`) — no `@latest` installs.

- 2026-09-23: added `BenchmarkFixEngine_EditListProvider_{1,10,100,1000}`
  (direct byte offsets) and `BenchmarkFixEngine_EditListProviderLineCol_{...}`
  (line/col resolution via the shared line index) for the v1.14.0
  EditListProvider. These are NOT in `baseline.txt` yet — benchstat gates
  only benchmarks present in both files, so they run ungated until the next
  baseline regeneration (which still requires a failed-gate justification
  per the Verschlimmbesser rules).
