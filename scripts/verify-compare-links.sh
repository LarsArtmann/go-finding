#!/usr/bin/env bash
# Verify every GitHub compare link in the markdown docs actually resolves,
# using AUTHENTICATED gh api calls. Anonymous link checkers (lychee on CI
# runners) flake on VALID compare pairs with 429/502 — the v1.15.0 train
# lost a lychee job to a months-old compare link that the API returns 200
# for. This script is the authoritative verdict; lychee reds on github.com
# compare links should be cross-checked here before anyone edits links.
#
# Extracts compare URLs of the shape
#   https://github.com/larsartmann/go-finding/compare/<base>...<head>
# from all tracked *.md files and requires each to return HTTP 200 from
#   gh api repos/larsartmann/go-finding/compare/<base>...<head>
#
# Usage:
#   ./scripts/verify-compare-links.sh
#
# Exit codes: 0 = all pairs resolve, 1 = at least one invalid, 2 = gh
# unavailable / no auth. Requires: gh CLI authenticated against the repo.
set -euo pipefail

cd "$(dirname "$0")/.."

REPO="larsartmann/go-finding"

if ! command -v gh >/dev/null; then
	echo "FAIL: gh CLI not available — cannot verify compare links"
	exit 2
fi

if ! gh auth status >/dev/null 2>&1; then
	echo "FAIL: gh is not authenticated — run: gh auth login"
	exit 2
fi

# Collect pairs from all tracked markdown files (code stripped by link-check
# conventions is unnecessary: compare URLs appear as markdown link targets).
pairs=$(git ls-files '*.md' |
	xargs grep -hoE "https://github\.com/$REPO/compare/[^)\"' ]+" |
	sed "s|https://github\.com/$REPO/compare/||" |
	sort -u)

if [ -z "$pairs" ]; then
	echo "OK: no compare links found in tracked markdown"
	exit 0
fi

FAILURES=0
TOTAL=0

while IFS= read -r pair; do
	TOTAL=$((TOTAL + 1))
	# URL-encode the pair path (tags carry slashes: pipeline/v1.15.0).
	encoded=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$pair")
	if gh api --silent "repos/$REPO/compare/$encoded" >/dev/null 2>&1; then
		echo "OK: $pair"
	else
		echo "FAIL: $pair (gh api non-200)"
		FAILURES=$((FAILURES + 1))
	fi
done <<<"$pairs"

echo "VERDICT: $((TOTAL - FAILURES))/$TOTAL compare pair(s) resolve."

if [ "$FAILURES" -gt 0 ]; then
	exit 1
fi
