# Benchmark Baseline

Captured: 2026-07-24
Environment: AMD RYZEN AI MAX+ 395 (32 threads), Linux, Go 1.26.4 with GOEXPERIMENT=jsonv2
Command: `go test -run=^$ -bench=. -benchmem -count=10 ./...`

## Usage

```bash
# Run benchmarks and compare against this baseline
go test -run=^$ -bench=. -benchmem -count=10 ./... > /tmp/current.txt
bash scripts/bench-check.sh benchmarks/baseline.txt /tmp/current.txt 25
```

The threshold (25%) is the maximum allowed regression per benchmark before the script exits non-zero.

## Notes

- This baseline is environment-dependent. Use it for relative comparison on the same machine.
- 10 iterations per benchmark for statistical significance (benchstat requires >= 6 for p-values).
- Regenerate after intentional performance changes: `cp /tmp/current.txt benchmarks/baseline.txt`
