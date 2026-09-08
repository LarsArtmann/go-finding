# Benchmark Baseline

Captured: 2026-09-08 evening (regenerated post-v1.8.0, including the FlightRecorder [Unreleased] tail)
Environment: AMD RYZEN AI MAX+ 395 (32 threads), Linux, Go 1.26.7 with GOEXPERIMENT=jsonv2
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
