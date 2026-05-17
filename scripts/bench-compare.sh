#!/usr/bin/env bash
# bench-compare.sh — Compare Go benchmark results against a baseline.
#
# Usage:
#   ./scripts/bench-compare.sh                    # Record baseline (first run)
#   ./scripts/bench-compare.sh                    # Compare against baseline (subsequent runs)
#   ./scripts/bench-compare.sh --reset            # Discard baseline and record new one
#   ./scripts/bench-compare.sh --bench <regex>    # Run specific benchmarks
#   ./scripts/bench-compare.sh --count <n>        # Run n times (default: 3)
#
# Requires: go (with benchstat), benchstat is built into Go 1.26+
#
# The baseline file is stored at .bench-baseline.txt (gitignored).

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
readonly BASELINE_FILE="$PROJECT_DIR/.bench-baseline.txt"
readonly CURRENT_FILE="$PROJECT_DIR/.bench-current.txt"

COUNT=3
BENCH="."
RESET=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --reset)
            RESET=true
            shift
            ;;
        --bench)
            BENCH="$2"
            shift 2
            ;;
        --count)
            COUNT="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [--reset] [--bench <regex>] [--count <n>]"
            exit 0
            ;;
        *)
            echo "Unknown flag: $1" >&2
            exit 1
            ;;
    esac
done

cd "$PROJECT_DIR"

if [[ "$RESET" == true ]] || [[ ! -f "$BASELINE_FILE" ]]; then
    echo "Recording baseline benchmarks ($COUNT runs, bench=$BENCH)..."
    go test -run='^$' -bench="$BENCH" -count="$COUNT" -benchmem ./... > "$BASELINE_FILE"
    echo "Baseline saved to ${BASELINE_FILE#"$PROJECT_DIR"/}"
    echo ""
    echo "Run $0 again to compare against this baseline."
    exit 0
fi

echo "Running current benchmarks ($COUNT runs, bench=$BENCH)..."
go test -run='^$' -bench="$BENCH" -count="$COUNT" -benchmem ./... > "$CURRENT_FILE"

echo ""
echo "=== Benchmark Comparison (old vs new) ==="
echo ""

if command -v benchstat &>/dev/null; then
    benchstat "$BASELINE_FILE" "$CURRENT_FILE"
else
    echo "benchstat not found. Install with: go install golang.org/x/perf/cmd/benchstat@latest"
    echo ""
    echo "Baseline: $BASELINE_FILE"
    echo "Current:  $CURRENT_FILE"
    echo ""
    echo "To compare manually:"
    echo "  benchstat $BASELINE_FILE $CURRENT_FILE"
fi
