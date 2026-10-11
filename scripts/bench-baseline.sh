#!/usr/bin/env bash
# Capture benchmark baseline for regression comparison.
#
# Usage:
#   ./scripts/bench-baseline.sh          # capture new baseline
#   ./scripts/bench-baseline.sh compare   # compare current vs baseline
#
# The baseline is stored in benchmarks/baseline.txt and should be committed
# when you're satisfied with current performance.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

BASELINE="benchmarks/baseline.txt"
CURRENT=$(mktemp)

trap 'rm -f "$CURRENT"' EXIT

case "${1:-capture}" in
capture)
	echo "Running benchmarks (count=10)..."
	go test -bench=. -benchmem -count=10 -run='^$' ./... 2>&1 | tee "$BASELINE"
	echo ""
	echo "Baseline saved to $BASELINE"
	echo "Commit this file to enable regression detection in CI."
	;;
compare)
	if [ ! -f "$BASELINE" ]; then
		echo "Error: No baseline found at $BASELINE"
		echo "Run: ./scripts/bench-baseline.sh capture"
		exit 1
	fi
	echo "Running benchmarks (count=10)..."
	go test -bench=. -benchmem -count=10 -run='^$' ./... 2>&1 | tee "$CURRENT"
	echo ""
	echo "=== benchstat comparison ==="
	benchstat "$BASELINE" "$CURRENT" || true
	;;
*)
	echo "Usage: $0 [capture|compare]"
	exit 1
	;;
esac
