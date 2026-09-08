#!/usr/bin/env bash
# bench-check.sh — Compare benchmark results against baseline and fail on regressions.
#
# Usage: bench-check.sh <baseline.txt> <current.txt> [threshold_pct]
#
# threshold_pct: maximum allowed regression per benchmark (default: 25).
# Exits 0 if no significant regressions, 1 if any benchmark regressed beyond threshold.
set -euo pipefail

baseline="${1:?usage: bench-check.sh <baseline.txt> <current.txt> [threshold]}"
current="${2:?usage: bench-check.sh <baseline.txt> <current.txt> [threshold]}"
threshold="${3:-25}"

if [ ! -f "$baseline" ]; then
	echo "⚠ No baseline found at $baseline — skipping regression check."
	echo "  To create a baseline: cp $current $baseline"
	exit 0
fi

echo "=== Benchmark comparison (threshold: +${threshold}%) ==="
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

# benchstat lines with a regression look like:
#   Clone-32   65.59n ±  3%   95.12n ± 23%  +45.03%  (p=0.000 n=10+10)
# We look for lines with a positive delta and statistical significance (no "~").
# The delta is a percentage like +30.00% — extract and compare to threshold.
# NOTE: benchstat strips the "Benchmark" prefix, so comparison lines start with
# the bare benchmark name. Match on the significance annotation "(p=" that every
# comparison line carries instead of a name prefix.
regressions=$(awk '
    /\(p=/ && /%/ {
        # Find the percentage field (contains + and %)
        for (i=1; i<=NF; i++) {
            if ($i ~ /^\+[0-9.]+%$/) {
                # Strip + and % to get the numeric value
                val = $i; gsub(/[+%]/, "", val)
                if (val + 0 > '"$threshold"' + 0) {
                    print $1 ": " $i " (line: " $0 ")"
                }
            }
        }
    }
' /tmp/benchstat-output.txt)

if [ -n "$regressions" ]; then
	echo ""
	echo "❌ FAIL: Benchmarks regressed beyond ${threshold}%:"
	echo "$regressions"
	exit 1
fi

echo ""
echo "✅ PASS: No benchmark regressed beyond ${threshold}%."
