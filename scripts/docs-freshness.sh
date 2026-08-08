#!/usr/bin/env bash
# docs-freshness.sh — Check documentation freshness against code changes.
#
# Flags docs that haven't been modified in N days AND docs whose referenced
# source files have been modified more recently than the doc itself.
#
# Warnings only — exits 0 always (informational, not a build gate).
#
# Usage: bash scripts/docs-freshness.sh [days]
#   days — staleness threshold in days (default: 180)
#
# Requires: git, bash 4+ (mapfile)

set -euo pipefail

THRESHOLD_DAYS="${1:-180}"
THRESHOLD_SECONDS=$((THRESHOLD_DAYS * 86400))
NOW=$(date +%s)
STALE_COUNT=0
SYNC_COUNT=0

# Find living documentation files (exclude archived, reviews, status reports)
mapfile -t FILES < <({
    find docs/guides -maxdepth 1 -name "*.md" 2>/dev/null || true
    find docs -maxdepth 1 -name "*.md" ! -name "MIGRATION*" 2>/dev/null || true
    echo "README.md"
    echo "TODO_LIST.md"
    echo "FEATURES.md"
} | sort -u)

for file in "${FILES[@]}"; do
    [ -f "$file" ] || continue

    # Get last commit timestamp for this doc file
    doc_ts=$(git log -1 --format="%ct" -- "$file" 2>/dev/null || echo "0")

    if [ "$doc_ts" -eq 0 ]; then
        echo "::warning file=$file::No git history (untracked or new file) — review for accuracy"
        continue
    fi

    doc_age=$((NOW - doc_ts))
    doc_age_days=$((doc_age / 86400))

    # Check 1: Simple staleness — doc not touched in threshold days
    if [ "$doc_age" -gt "$THRESHOLD_SECONDS" ]; then
        echo "::warning file=$file::Stale doc — last modified ${doc_age_days} days ago (threshold: ${THRESHOLD_DAYS}d). Review for accuracy."
        STALE_COUNT=$((STALE_COUNT + 1))
    fi

    # Check 2: Code-doc sync — find referenced .go files modified after the doc
    # Extract Go file references from markdown: `filename.go` or path/to/filename.go
    while IFS= read -r gofile; do
        [ -f "$gofile" ] || continue

        go_ts=$(git log -1 --format="%ct" -- "$gofile" 2>/dev/null || echo "0")
        [ "$go_ts" -eq 0 ] && continue

        if [ "$go_ts" -gt "$doc_ts" ]; then
            days_ahead=$(( (go_ts - doc_ts) / 86400))
            echo "::warning file=$file::Referenced source $gofile modified ${days_ahead}d after this doc. Content may be stale."
            SYNC_COUNT=$((SYNC_COUNT + 1))
            break # one stale reference per doc is enough
        fi
    done < <(grep -oP '`?\K[a-zA-Z0-9_/]+\.go' "$file" 2>/dev/null | sort -u | head -20 || true)
done

echo ""
echo "Docs freshness check: $STALE_COUNT stale doc(s), $SYNC_COUNT out-of-sync reference(s)"
echo "(threshold: ${THRESHOLD_DAYS} days)"

exit 0
