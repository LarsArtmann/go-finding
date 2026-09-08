# Benchmark Baseline

Captured: 2026-09-08 (regenerated at v1.7.0)
Environment: AMD RYZEN AI MAX+ 395 (32 threads), Linux, Go 1.26.7 with GOEXPERIMENT=jsonv2
Command (multi-module — plain `go test ./...` from the root covers only the core module):

```bash
export GOEXPERIMENT=jsonv2
go test -run='^$' -bench=. -benchmem -count=10 ./... > /tmp/current.txt
(cd pipeline && go test -run='^$' -bench=. -benchmem -count=10 ./... >> /tmp/current.txt)
bash scripts/bench-check.sh benchmarks/baseline.txt /tmp/current.txt 25
```

## Usage

The threshold (25%) is the maximum allowed regression per benchmark before the script exits non-zero.

## Notes

- This baseline is environment-dependent. Use it for relative comparison on the same machine.
- 10 iterations per benchmark for statistical significance (benchstat requires >= 6 for p-values).
- Regenerate after intentional performance changes: `cp /tmp/current.txt benchmarks/baseline.txt`
- **2026-09-08 regeneration rationale (v1.7.0):** v1.7.0 added `GroupID` to the `Finding`
  struct and enriched the LSP round-trip (tag re-emission). Finding-copy benchmarks
  (`Clone`, `Correlate`, `MergeNoDedup`, `LSPRoundTrip`) pay a real CPU cost for the
  larger struct — with **identical allocation profiles** (B/op and allocs/op unchanged).
  Cross-session absolute times on this machine also swing widely for these benchmarks
  (Correlate measured -9% then +86% on unchanged code in two same-day runs), so the
  baseline was recaptured same-session.
- `bench-check.sh` requires `benchstat` on PATH
  (`go install golang.org/x/perf/cmd/benchstat@latest`).
