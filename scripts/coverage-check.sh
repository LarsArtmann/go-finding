#!/usr/bin/env bash
# Per-package coverage threshold checker.
# Usage: ./scripts/coverage-check.sh
# Fails with exit code 1 if any package is below its threshold.

set -euo pipefail

# Package thresholds (percentage)
declare -A THRESHOLDS=(
    ["github.com/larsartmann/go-finding"]=98.0
    ["github.com/larsartmann/go-finding/pipeline"]=95.0
    ["github.com/larsartmann/go-finding/cmd/go-finding"]=90.0
    ["github.com/larsartmann/go-finding/internal/detectors"]=90.0
)

TOTAL_THRESHOLD=93.0
FAILED=0

TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT

echo "Running tests with coverage..."
go test -cover ./... > "$TMPFILE" 2>&1

# Check per-package coverage
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
done < "$TMPFILE"

# Check total coverage using the profile
echo ""
echo "Checking total coverage..."
COVER_OUT=$(mktemp)
trap 'rm -f "$TMPFILE" "$COVER_OUT"' EXIT

go test -coverprofile="$COVER_OUT" ./... > /dev/null 2>&1 || true
TOTAL_COV=$(go tool cover -func="$COVER_OUT" | awk '/^total:/{print $3}' | sed 's/%//')

if awk "BEGIN { exit (!($TOTAL_COV < $TOTAL_THRESHOLD)) }"; then
    echo "FAIL: total coverage ${TOTAL_COV}% is below ${TOTAL_THRESHOLD}% threshold"
    FAILED=1
else
    echo "PASS: total coverage ${TOTAL_COV}% (threshold: ${TOTAL_THRESHOLD}%)"
fi

if [[ "$FAILED" -ne 0 ]]; then
    exit 1
fi
