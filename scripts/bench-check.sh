#!/usr/bin/env bash
# bench-check.sh — Compare benchmark results against baseline and fail on regressions.
#
# Usage: bench-check.sh <baseline.txt> <current.txt> [time_threshold_pct] [alloc_threshold_pct]
#
#   time_threshold_pct:   maximum allowed sec/op regression (default: 250)
#   alloc_threshold_pct:  maximum allowed B/op or allocs/op regression (default: 10)
#
# Gate philosophy (2026-09-08, evening session): on this AMD APU, absolute
# sec/op swings of -9%..+150% were measured on IDENTICAL code depending on
# position in the benchmark suite (sustained full-suite runs thermally depress
# later benchmarks: IntervalIndex_Query/1000 measured 9.3µ in a fresh filtered
# run vs 20.1µ late in a full-suite run, n=10, p=0.000 both). Time comparisons
# across sessions are therefore only trustworthy for algorithmic blowouts
# (default 250% = 3.5x), not fine regressions. A 2x pure-CPU regression is
# physically indistinguishable from thermal noise on this hardware — if you
# need that precision, run a fresh filtered benchmark against the suspect
# change instead. Allocation metrics (B/op, allocs/op) are deterministic
# (±0% on unchanged code in every observed run) and are the primary signal.
#
# Exits 0 if no regression beyond thresholds, 1 otherwise.
set -euo pipefail

baseline="${1:?usage: bench-check.sh <baseline.txt> <current.txt> [time_threshold] [alloc_threshold]}"
current="${2:?usage: bench-check.sh <baseline.txt> <current.txt> [time_threshold] [alloc_threshold]}"
time_threshold="${3:-250}"
alloc_threshold="${4:-10}"

if [ ! -f "$baseline" ]; then
	echo "⚠ No baseline found at $baseline — skipping regression check."
	echo "  To create a baseline: cp $current $baseline"
	exit 0
fi
if [ ! -s "$current" ]; then
	echo "❌ FAIL: current benchmark file $current is missing or empty — capture error."
	exit 1
fi

echo "=== Benchmark comparison (time +${time_threshold}%, alloc +${alloc_threshold}%) ==="
# benchstat resolution order: PATH binary, else the pinned Go tool directive
# (go.mod: tool golang.org/x/perf/cmd/benchstat — reproducible comparisons).
BENCHSTAT=benchstat
if ! command -v benchstat >/dev/null 2>&1; then
	BENCHSTAT="go tool benchstat"
fi
export GOEXPERIMENT="${GOEXPERIMENT:-jsonv2}"
$BENCHSTAT "$baseline" "$current" | tee /tmp/benchstat-output.txt

echo ""
echo "=== Regression check ==="

# benchstat emits three tables in order: sec/op, B/op, allocs/op. Each table
# header line contains the unit label ("sec/op", "B/op", "allocs/op"); data
# lines with significance carry "(p=". We track the current table's unit and
# apply the matching threshold: time_threshold for sec/op (thermal-noise
# tolerant), alloc_threshold for B/op and allocs/op (deterministic signal).
regressions=$(awk -v time_thr="$time_threshold" -v alloc_thr="$alloc_threshold" '
    /^(goos|goarch|pkg|cpu):/ { next }
    /sec\/op/  && /\│/ { unit = "time"; next }
    /B\/op/    && /\│/ { unit = "bytes"; next }
    /allocs\/op/ && /\│/ { unit = "allocs"; next }
    /\(p=/ && /%/ {
        thr = (unit == "time") ? time_thr : alloc_thr
        for (i = 1; i <= NF; i++) {
            if ($i ~ /^\+[0-9.]+%$/) {
                val = $i; gsub(/[+%]/, "", val)
                if (val + 0 > thr + 0) {
                    print $1 " [" unit "]: " $i " (line: " $0 ")"
                }
            }
        }
    }
' /tmp/benchstat-output.txt)

if [ -n "$regressions" ]; then
	echo ""
	echo "❌ FAIL: Benchmarks regressed beyond thresholds (time +${time_threshold}%, alloc +${alloc_threshold}%):"
	echo "$regressions"
	exit 1
fi

echo ""
echo "✅ PASS: No benchmark regressed beyond time +${time_threshold}% / alloc +${alloc_threshold}%."
