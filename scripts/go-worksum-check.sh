#!/usr/bin/env bash
# go.work.sum freshness gate.
#
# Legs:
#   1. WORKSPACE GRAPH — `go list -m all` in workspace mode. Missing or
#      corrupted checksums die here with the checksum-mismatch error class
#      the 09-27 dependency bumps produced in cold-cache CI.
#   2. LINE HYGIENE — every go.work.sum line parses as `module version hash`.
#   3. SYNC IDEMPOTENCE (--sync, opt-in because it mutates the tree) —
#      `go work sync` must leave go.work and go.work.sum unchanged. A diff
#      proves the files were stale or incomplete; the diff IS the fix.
#
# Deliberately NOT a gate: entries for module versions outside the current
# workspace graph. go.work.sum legitimately accumulates supersets (test
# deps of dependencies, tool graphs, pruned-history residue); pruning them
# is not required for correctness and flagging them is a false-positive
# machine (proven 2026-11-08: ~140 such entries on a healthy tree).
#
# Usage: ./scripts/go-worksum-check.sh [--sync]
# Exit codes: 0 = fresh, 1 = findings, 2 = setup error.
set -euo pipefail

cd "$(dirname "$0")/.."

FAILURES=0

step() { printf '\n== %s ==\n' "$1"; }

SYNC=0
for arg in "$@"; do
	case "$arg" in
	--sync) SYNC=1 ;;
	*)
		echo "usage: $0 [--sync]"
		exit 2
		;;
	esac
done

step "workspace graph resolves (missing/broken checksums class)"
if ! graph=$(go list -m all 2>&1); then
	echo "FAIL: go list -m all failed in workspace mode:"
	echo "$graph" | head -5
	exit 1
fi
echo "OK: workspace graph resolves ($(echo "$graph" | wc -l | tr -d ' ') modules)"

step "go.work.sum lines parse"
if [ ! -f go.work.sum ]; then
	echo "FAIL: go.work.sum missing"
	exit 1
fi
bad_lines=$(grep -vcE '^\S+ \S+ h1:\S+$' go.work.sum || true)
if [ "${bad_lines:-0}" -gt 0 ]; then
	echo "FAIL: $bad_lines unparseable line(s) in go.work.sum"
	FAILURES=$((FAILURES + 1))
else
	echo "OK: all $(grep -c '' go.work.sum | tr -d ' ') lines parse"
fi

if [ "$SYNC" -eq 1 ]; then
	step "go work sync idempotence (stale/incomplete class)"
	git diff --quiet -- go.work go.work.sum ||
		{
			echo "FAIL: pre-existing uncommitted changes to go.work/go.work.sum — commit or restore first"
			exit 1
		}
	go work sync >/dev/null 2>&1 || {
		echo "FAIL: go work sync errored"
		exit 1
	}
	if git diff --quiet -- go.work go.work.sum; then
		echo "OK: go work sync changed nothing"
	else
		echo "FAIL: go work sync rewrote go.work/go.work.sum — the files were stale."
		echo "The working tree now holds the CORRECTED files; review and commit them:"
		git diff --stat -- go.work go.work.sum
		FAILURES=$((FAILURES + 1))
	fi
fi

echo "GO.WORK.SUM VERDICT: $([ "$FAILURES" -eq 0 ] && echo PASS || echo "FAIL ($FAILURES class(es))")"
exit "$FAILURES"
