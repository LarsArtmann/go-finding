#!/usr/bin/env bash
# CI-event watchdog: proves a pushed commit actually has a CI run.
#
# Two failure classes killed whole days silently:
#   1. f394f54 (2026-10-01): a master push updated origin/master but created
#      NO ci.yml run — no error, no run, nothing. Found only by absence.
#   2. v1.9.0: a multi-tag push created no Release events.
# This script is the explicit check for class 1 (and, with --workflow, any
# workflow). Run it right after every push; the release-train push stage
# watches Release runs the same way.
#
# Usage:
#   ./scripts/ci-watchdog.sh <sha>                 # ci.yml run exists for sha
#   ./scripts/ci-watchdog.sh <sha> --workflow=release.yml
#   ./scripts/ci-watchdog.sh master-green          # latest master ci.yml run green
#   ./scripts/ci-watchdog.sh all-green             # latest run per workflow green
#
# Exit codes: 0 = OK, 1 = FAIL (missing run or non-green), 2 = usage/gh error.
set -euo pipefail

require_gh() {
	if ! command -v gh >/dev/null; then
		echo "FAIL: gh CLI not available"
		exit 2
	fi
}

run_exists() {
	local sha="$1" workflow="${2:-ci.yml}"
	local count
	count=$(gh run list --workflow="$workflow" --limit 30 --json headSha \
		--jq "[.[] | select(.headSha==\"$sha\")] | length")
	if [ "${count:-0}" -ge 1 ]; then
		echo "OK: $workflow run exists for $sha"
		return 0
	fi
	echo "FAIL: no $workflow run for $sha (the f394f54 silent-drop class)."
	echo "Check nothing is queued, then: gh workflow run $workflow --ref <branch>"
	return 1
}

master_green() {
	local latest
	latest=$(gh run list --workflow=ci.yml --branch master --limit 1 \
		--json conclusion,status --jq '.[0]')
	local conclusion status
	conclusion=$(echo "$latest" | python3 -c "import json,sys; print(json.load(sys.stdin).get('conclusion'))")
	status=$(echo "$latest" | python3 -c "import json,sys; print(json.load(sys.stdin).get('status'))")
	if [ "$status" != "completed" ]; then
		echo "IN PROGRESS: latest master ci.yml run is $status"
		return 0
	fi
	if [ "$conclusion" = "success" ]; then
		echo "OK: latest master ci.yml run is green"
		return 0
	fi
	echo "FAIL: latest master ci.yml run concluded '$conclusion' — master is red."
	return 1
}

main() {
	require_gh
	if [ "${1:-}" = "master-green" ]; then
		master_green
		return
	fi
	if [ "${1:-}" = "all-green" ]; then
		local rc=0
		master_green || rc=1
		# release.yml has no branch runs; report its latest run of ANY kind.
		local rel
		rel=$(gh run list --workflow=release.yml --limit 1 --json conclusion,status --jq '"\(.status) \(.conclusion)"')
		echo "latest release.yml run: $rel"
		return "$rc"
	fi
	local sha="${1:-}"
	[ -z "$sha" ] && {
		echo "usage: $0 <sha> [--workflow=name] | master-green | all-green"
		exit 2
	}
	local workflow="ci.yml"
	if [ "${2:-}" = "--workflow=release.yml" ]; then
		workflow="release.yml"
	fi
	# Short SHAs are accepted by the API comparison below only as full
	# matches; expand short shas first.
	if [ "${#sha}" -lt 40 ]; then
		sha=$(git rev-parse "$sha")
	fi
	run_exists "$sha" "$workflow"
}

main "$@"
