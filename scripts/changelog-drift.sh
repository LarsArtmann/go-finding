#!/usr/bin/env bash
# CHANGELOG-drift gate: every release tag must have a matching section in its
# module's CHANGELOG.md. Born from the v1.10/v1.11 incident: sub-module
# CHANGELOGs silently drifted (no sections for two releases) until
# hand-backfilled, and again at v1.13.0 (3 of 5 modules shipped tag+release
# with no section). Nothing gated it — this script is the gate.
#
# Mapping (docs/release-procedure.md "Multi-Module Tagging"):
#   v<ver>              -> ./CHANGELOG.md
#   pipeline/v<ver>     -> pipeline/CHANGELOG.md
#   analysis/v<ver>     -> analysis/CHANGELOG.md
#   toolsdk/v<ver>      -> toolsdk/CHANGELOG.md
#   cmd/go-finding/v<ver> -> cmd/go-finding/CHANGELOG.md
#
# A tag is expected to have a "## [<ver>]" section (Keep a Changelog).
# Ancient tags that predate a module's CHANGELOG discipline are listed in
# KNOWN_GAPS below — explicit, reviewable, and frozen: any tag NOT in that
# list without a section fails. New entries in KNOWN_GAPS require a written
# justification in this file.
#
# Usage: ./scripts/changelog-drift.sh
set -euo pipefail

cd "$(dirname "$0")/.."

# Tag -> CHANGELOG file mapping (prefix = tag glob for git tag --list).
MODULES=(
	"v[0-9]*:CHANGELOG.md"
	"pipeline/v*:pipeline/CHANGELOG.md"
	"analysis/v*:analysis/CHANGELOG.md"
	"toolsdk/v*:toolsdk/CHANGELOG.md"
	"cmd/go-finding/v*:cmd/go-finding/CHANGELOG.md"
)

# Tags accepted WITHOUT a CHANGELOG section. Ancient history predating a
# module's changelog discipline; frozen list, do not grow it.
# Justifications: v0.2.0/v0.3.0 + */v0.1.0 predate any CHANGELOG in the repo;
# 1.3.x-1.9.0 sub-module gaps were discovered (and accepted as history) when
# this gate was introduced — they predate the 2026-09-17 multi-module
# release-hygiene reset.
KNOWN_GAPS=(
	"v0.2.0" "v0.3.0"
	"pipeline/v0.1.0" "pipeline/v1.3.0" "pipeline/v1.4.0" "pipeline/v1.4.1"
	"analysis/v0.1.0" "analysis/v1.3.0" "analysis/v1.4.0" "analysis/v1.6.0" "analysis/v1.9.0"
	"cmd/go-finding/v0.1.0" "cmd/go-finding/v1.3.0" "cmd/go-finding/v1.4.0" "cmd/go-finding/v1.4.1"
)

FAILURES=0
CHECKED=0
GAP_HITS=0

in_known_gaps() {
	local tag="$1" gap
	for gap in "${KNOWN_GAPS[@]}"; do
		[ "$gap" = "$tag" ] && return 0
	done
	return 1
}

echo "Checking CHANGELOG sections for release tags..."

for spec in "${MODULES[@]}"; do
	pattern="${spec%%:*}"
	file="${spec#*:}"
	if [ ! -f "$file" ]; then
		echo "FAIL: $file does not exist (module CHANGELOG missing)"
		FAILURES=$((FAILURES + 1))
		continue
	fi
	while IFS= read -r tag; do
		CHECKED=$((CHECKED + 1))
		# Strip the directory prefix and the leading v: "pipeline/v1.2.0" -> "1.2.0".
		ver="${tag##*/}"
		ver="${ver#v}"
		if grep -q "^## \[${ver}\]" "$file"; then
			echo "OK: $tag -> $file has section [$ver]"
		elif in_known_gaps "$tag"; then
			echo "GAP (known): $tag has no [$ver] section in $file (accepted history)"
			GAP_HITS=$((GAP_HITS + 1))
		else
			echo "FAIL: $tag has NO [$ver] section in $file — every release tag needs a CHANGELOG section"
			FAILURES=$((FAILURES + 1))
		fi
	done < <(git tag --list "$pattern" | sort -V)
done

echo ""
echo "Checked $CHECKED tag(s): $((CHECKED - FAILURES - GAP_HITS)) sectioned, $GAP_HITS known-gap(s)."

if [ "$FAILURES" -gt 0 ]; then
	echo "FAIL: CHANGELOG drift — $FAILURES tag(s) without sections. Backfill or justify before releasing."
	exit 1
fi

echo "OK: CHANGELOG drift check passed."
