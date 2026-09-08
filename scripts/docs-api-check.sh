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
	"AI"          # concept, not a symbol
	"CPU"
	"CSV"
	"IO"
	"JSON"
	"LSP"
	"SARIF"
	"TSV"
	"ID"          # branded type exists, listed to keep the allowlist explicit
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
