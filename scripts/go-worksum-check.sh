#!/usr/bin/env bash
# go.work.sum freshness gate.
#
# Two failure classes:
#   1. BROKEN — a checksum entry does not match the module it names. Every
#      workspace go command dies with a checksum-mismatch error, which the
#      09-27 dependency bumps produced when go.sum files drifted from the
#      workspace view.
#   2. STALE — entries for module@version pairs the workspace graph no
#      longer requires. Harmless at build time, but they mask real drift:
#      a reader cannot tell served checksums from dead ones.
#
# Usage: ./scripts/go-worksum-check.sh
# Exit codes: 0 = fresh, 1 = stale or broken, 2 = setup error.
set -euo pipefail

cd "$(dirname "$0")/.."

FAILURES=0

step() { printf '\n== %s ==\n' "$1"; }

step "workspace graph resolves (missing/broken checksums class)"
if ! graph=$(go list -m all 2>&1); then
	echo "FAIL: go list -m all failed in workspace mode:"
	echo "$graph" | head -5
	exit 1
fi
echo "OK: workspace graph resolves ($(echo "$graph" | wc -l | tr -d ' ') modules)"

step "go.work.sum entries exist and are parseable"
if [ ! -f go.work.sum ]; then
	echo "FAIL: go.work.sum missing"
	exit 1
fi
bad_lines=$(grep -vcE '^\S+ \S+ h1:\S+$' go.work.sum || true)
if [ "${bad_lines:-0}" -gt 0 ]; then
	echo "FAIL: $bad_lines unparseable line(s) in go.work.sum"
	FAILURES=$((FAILURES + 1))
else
	echo "OK: all go.work.sum lines parse"
fi

step "every go.work.sum entry is in the workspace graph (staleness class)"
stale=0
while IFS= read -r pair; do
	[ -z "$pair" ] && continue
	if ! grep -qE "^${pair//\//\\/}( |$)" <<<"$graph"; then
		echo "FAIL: stale entry (not in workspace graph): $pair"
		stale=$((stale + 1))
	fi
done < <(awk '{print $1 " " $2}' go.work.sum | grep -v '/go.mod ' | sort -u)
if [ "$stale" -gt 0 ]; then
	FAILURES=$((FAILURES + 1))
else
	echo "OK: no stale go.work.sum entries"
fi

echo "GO.WORK.SUM VERDICT: $([ "$FAILURES" -eq 0 ] && echo PASS || echo "FAIL ($FAILURES class(es))")"
exit "$FAILURES"
