#!/usr/bin/env bash
# docs-api-check.sh — Documentation API drift guard.
#
# Extracts backtick code spans that look like exported Go identifiers from the
# living docs (FEATURES.md, API_STABILITY.md) and verifies each one appears in
# the Go source. A documented API name that no longer exists in code is a doc
# bug; this catches it mechanically (it would have caught most of the ~26
# stale claims found in the 2026-09-08 docs audit).
#
# Usage: ./scripts/docs-api-check.sh
# Exits 0 if every documented identifier resolves, 1 with a list otherwise.
set -euo pipefail

cd "$(dirname "$0")/.."

DOCS=(FEATURES.md docs/API_STABILITY.md docs/DOMAIN_LANGUAGE.md docs/USAGE_GUIDE.md)

# Identifier-shaped spans only: exported Go names (CamelCase, digits allowed).
# Lowercase/mixed spans (flags, paths, versions, prose code words) are skipped;
# the allowlist below covers intentionally non-code uppercase spans.
ALLOWLIST=(
	"AI" # concept, not a symbol
	"CPU"
	"CSV"
	"IO"
	"JSON"
	"LSP"
	"SARIF"
	"TSV"
	"ID" # branded type exists, listed to keep the allowlist explicit
	"OK"
	"CLI"
	"TOML"
	"YAML"
)

# Names that appear in docs ONLY as historical migration references
# ("replaces X", "X -> Use Y"). Adding a name here must include where the
# rename is documented.
HISTORICAL=(
	"GetCategory" # renamed to CategoryOf (API_STABILITY.md migration table)
)

is_allowed() {
	local want="$1" a
	for a in "${ALLOWLIST[@]}" "${HISTORICAL[@]}"; do
		[ "$a" = "$want" ] && return 0
	done
	return 1
}

# Collect the candidate identifiers.
mapfile -t candidates < <(
	cat "${DOCS[@]}" |
		grep -o '`[^`]*`' |
		tr -d '`' |
		sort -u |
		grep -E '^[A-Z][A-Za-z0-9_]*$' || true
)

missing=()
for id in "${candidates[@]}"; do
	is_allowed "$id" && continue

	# Exists if any Go file mentions the identifier outside a comment-only line.
	if ! rg -q --glob '*.go' --glob '!**/vendor/**' -e "(\\b|\\.)${id}([ (<.,:)\[]|\$)" -e "\"${id}\"" -e "^// .*${id}" .; then
		missing+=("$id")
	fi
done

if [ "${#missing[@]}" -gt 0 ]; then
	echo "ERROR: documented identifiers not found in any .go file:"
	for id in "${missing[@]}"; do
		echo "  - $id"
	done
	echo ""
	echo "Fix the docs (or rename the code and update docs together)."
	exit 1
fi

echo "OK: ${#candidates[@]} documented identifiers all resolve in code."

# --- Version-claim guard (FEATURES.md must not claim unreleased versions) ---
# FEATURES.md is the honest inventory of what EXISTS. Any vX.Y.Z mentioned
# there must be <= version.go's version. Future work belongs in ROADMAP.md
# (or a PLANNED row without a version stamp).
MAJOR=$(sed -n 's/^const VersionMajor = //p' version.go)
MINOR=$(sed -n 's/^const VersionMinor = //p' version.go)
PATCH=$(sed -n 's/^const VersionPatch = //p' version.go)

version_gt() { # $1 > $2 in semver (M.m.p, dot-separated)
	local a b IFS=.
	read -r a1 a2 a3 <<<"$1"
	read -r b1 b2 b3 <<<"$2"
	[ "$a1" -ne "$b1" ] && { [ "$a1" -gt "$b1" ]; return; }
	[ "$a2" -ne "$b2" ] && { [ "$a2" -gt "$b2" ]; return; }
	[ "$a3" -gt "$b3" ]
}

declare -A overclaims=()
while read -r v; do
	[ -z "$v" ] && continue
	if version_gt "${v#v}" "$MAJOR.$MINOR.$PATCH"; then
		overclaims["$v"]=1
	fi
done < <(grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' FEATURES.md | sort -u)

if [ "${#overclaims[@]}" -gt 0 ]; then
	echo "ERROR: FEATURES.md claims version(s) beyond version.go ($MAJOR.$MINOR.$PATCH):"
	for v in "${!overclaims[@]}"; do
		echo "  - $v"
	done
	echo ""
	echo "FEATURES documents what exists. Move future-version claims to ROADMAP.md,"
	echo "or bump version.go first (release train order)."
	exit 1
fi

UNRELEASED_COUNT=$(grep -ci 'unreleased' FEATURES.md || true)
echo "OK: no version claims beyond v$MAJOR.$MINOR.$PATCH in FEATURES.md"
if [ "$UNRELEASED_COUNT" -gt 0 ]; then
	echo "NOTE: FEATURES.md mentions 'unreleased' ${UNRELEASED_COUNT}x — fine mid-cycle, must be 0 at tag time (release-procedure step 7)."
fi
