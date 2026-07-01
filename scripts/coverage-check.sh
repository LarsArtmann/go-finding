#!/usr/bin/env bash
# Per-package coverage threshold checker.
# Usage: ./scripts/coverage-check.sh [coverage.out]
#   If a coverage profile argument is provided, it's used for total coverage.
#   Per-package coverage is always read from go test output.
#   Fails with exit code 1 if any package is below its threshold.

set -euo pipefail

# Package thresholds (percentage)
declare -A THRESHOLDS=(
    ["github.com/larsartmann/go-finding"]=98.0
    ["github.com/larsartmann/go-finding/pipeline"]=95.0
    ["github.com/larsartmann/go-finding/cmd/go-finding"]=90.0
    ["github.com/larsartmann/go-finding/cmd/go-finding/internal/detectors"]=90.0
)

TOTAL_THRESHOLD=93.0
FAILED=0

COVERPROFILE=""
if [[ $# -ge 1 && -f "$1" ]]; then
    COVERPROFILE="$1"
fi

# Check per-package coverage from go test output
echo "Running tests with coverage..."
TESTFILE=$(mktemp)
trap 'rm -f "$TESTFILE"' EXIT

go test -cover ./... > "$TESTFILE" 2>&1 || true

while IFS= read -r line; do
    if [[ "$line" =~ coverage:\ ([0-9]+\.[0-9]+)%\ of\ statements ]]; then
        pct="${BASH_REMATCH[1]}"
        pkg=$(echo "$line" | awk '{print $2}')
        [[ -z "${THRESHOLDS[$pkg]:-}" ]] && continue

        thresh="${THRESHOLDS[$pkg]}"
        if awk "BEGIN { exit (!($pct < $thresh)) }"; then
            echo "FAIL: $pkg coverage ${pct}% is below ${thresh}% threshold"
            FAILED=1
        else
            echo "PASS: $pkg coverage ${pct}% (threshold: ${thresh}%)"
        fi
    fi
done < "$TESTFILE"

# Check total coverage
echo ""

if [[ -n "$COVERPROFILE" ]]; then
    echo "Checking total coverage from profile: $COVERPROFILE"
    TOTAL_COV=$(go tool cover -func="$COVERPROFILE" | awk '/^total:/{print $3}' | sed 's/%//')
else
    echo "Checking total coverage..."
    COVER_OUT=$(mktemp)
    trap 'rm -f "$TESTFILE" "$COVER_OUT"' EXIT
    go test -coverprofile="$COVER_OUT" ./... > /dev/null 2>&1 || true
    TOTAL_COV=$(go tool cover -func="$COVER_OUT" | awk '/^total:/{print $3}' | sed 's/%//')
fi

if awk "BEGIN { exit (!($TOTAL_COV < $TOTAL_THRESHOLD)) }"; then
    echo "FAIL: total coverage ${TOTAL_COV}% is below ${TOTAL_THRESHOLD}% threshold"
    FAILED=1
else
    echo "PASS: total coverage ${TOTAL_COV}% (threshold: ${TOTAL_THRESHOLD}%)"
fi

if [[ "$FAILED" -ne 0 ]]; then
    exit 1
fi
