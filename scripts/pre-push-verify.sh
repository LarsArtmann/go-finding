#!/usr/bin/env bash
# Pre-push verification: everything that failed CI in the past, in one
# command, before the push instead of after. Fast gates only — the heavy
# stress/bench gates stay release-time (scripts/release-preflight.sh).
#
# Rules honored (AGENTS.md dead-gate lesson): every check prints an explicit
# verdict line; this script exits non-zero on the first red so a filter can
# never hide a verdict.
#
# Usage: ./scripts/pre-push-verify.sh [--quick]
#   --quick skips race (-race) and uses -count=1 without the race detector.

set -euo pipefail

cd "$(dirname "$0")/.."

QUICK=0
[ "${1:-}" = "--quick" ] && QUICK=1

RACE="-race"
[ "$QUICK" -eq 1 ] && RACE=""

FAILED=0
step() { echo ""; echo "=== $1 ==="; }
check() {
	local label="$1"
	shift
	if "$@" > /tmp/pre-push-last.log 2>&1; then
		echo "OK: $label"
	else
		echo "FAIL: $label"
		tail -30 /tmp/pre-push-last.log
		exit 1
	fi
}

step "disk space on shared build cache"
BUILDCACHE_FREE_KB=$(df -k /mnt/buildcache 2>/dev/null | awk 'NR==2 {print $4}')
if [ -z "$BUILDCACHE_FREE_KB" ]; then
	echo "SKIP: /mnt/buildcache not mounted"
elif [ "$BUILDCACHE_FREE_KB" -lt $((5 * 1024 * 1024)) ]; then
	echo "FAIL: /mnt/buildcache has less than 5 GiB free ($((BUILDCACHE_FREE_KB / 1024 / 1024)) GiB)"
	echo "      builds will die with misleading 'no space left' errors; recover with: go clean -cache"
	exit 1
else
	echo "OK: /mnt/buildcache has $((BUILDCACHE_FREE_KB / 1024 / 1024)) GiB free"
fi

step "formatting (treefmt via nix fmt --fail-on-change)"
check "nix fmt" nix fmt

step "workflow lint"
command -v actionlint > /dev/null && check "actionlint" actionlint .github/workflows/ci.yml .github/workflows/release.yml ||
	echo "SKIP: actionlint not installed"

step "race tests per module"
for mod in . pipeline analysis toolsdk cmd/go-finding; do
	if [ "$QUICK" -eq 1 ]; then
		check "test $mod" bash -c "cd '$mod' && GOWORK=off go test -count=1 ./..."
	else
		check "test -race $mod" bash -c "cd '$mod' && GOWORK=off go test -race -count=1 ./..."
	fi
done

step "structural gates"
check "version-drift" bash scripts/version-drift.sh
check "replace-audit" bash scripts/replace-audit.sh
check "go-work-sync" bash scripts/go-work-sync.sh
check "test-naming" bash scripts/test-naming.sh
check "json-deterministic-check" bash scripts/json-deterministic-check.sh
check "changelog-drift" bash scripts/changelog-drift.sh
check "docs-api-check" bash scripts/docs-api-check.sh
check "docs-freshness" bash scripts/docs-freshness.sh
check "link-check" bash scripts/link-check.sh
check "version-check" bash scripts/version-check.sh
check "preflight-selftest" bash scripts/release-preflight-selftest.sh

step "lint"
check "golangci-lint (root)" golangci-lint run ./...

echo ""
echo "PRE-PUSH VERIFY PASS: formatting, tests, structural gates, and lint green."
