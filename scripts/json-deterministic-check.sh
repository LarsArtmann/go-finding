#!/usr/bin/env bash
# json-deterministic-check.sh — Verify all production json.Marshal/json.MarshalWrite
# calls include the Deterministic option or use shared marshalOpts/prettyMarshalOpts.
#
# encoding/json/v2 serializes Go map keys in unspecified order by default.
# Without json.Deterministic(true), every marshal call with maps produces
# non-reproducible output. This script prevents future regressions at code level.
#
# Exits 1 if violations found, 0 otherwise.
#
# Usage: bash scripts/json-deterministic-check.sh

set -euo pipefail

VIOLATIONS=0

# Find all production .go files (exclude test files, vendor, generated, doc.go).
mapfile -t FILES < <({
    find . -name "*.go" \
        ! -name "*_test.go" \
        ! -name "*.gen.go" \
        ! -name "doc.go" \
        ! -path "./vendor/*" \
        ! -path "./.git/*" \
        2>/dev/null
} | sort -u)

for file in "${FILES[@]}"; do
    [ -f "$file" ] || continue

    # Find lines containing json.Marshal( / json.MarshalWrite( / json.MarshalEncode(
    while IFS=: read -r linenum line; do
        [ -z "$linenum" ] && continue

        # Skip comment lines.
        trimmed="${line#"${line%%[![:space:]]*}"}"
        case "$trimmed" in
            \/\/*) continue ;;
        esac

        # Extract the call statement: from this line to the matching ')'
        # (tracking paren depth to handle nested function calls correctly).
        call=$(sed -n "${linenum},$((linenum + 20))p" "$file" |
            awk '
                {
                    for (i = 1; i <= length($0); i++) {
                        c = substr($0, i, 1)
                        if (c == "(") depth++
                        else if (c == ")") depth--
                    }
                    print
                    if (depth <= 0) exit
                }
            ')

        # The call is compliant if it contains Deterministic,
        # marshalOpts, or prettyMarshalOpts.
        if ! echo "$call" | grep -qE 'Deterministic|marshalOpts|prettyMarshalOpts'; then
            echo "::error file=$file,line=$linenum::json.Marshal call without json.Deterministic(true)"
            echo "  ${file}:${linenum}: ${trimmed}"
            VIOLATIONS=$((VIOLATIONS + 1))
        fi
    done < <(grep -nE 'json\.Marshal(Write|Encode)?\(' "$file" 2>/dev/null || true)
done

echo ""
if [ "$VIOLATIONS" -gt 0 ]; then
    echo "Found $VIOLATIONS production marshal call(s) missing json.Deterministic(true)."
    echo "All production json.Marshal / json.MarshalWrite calls must include Deterministic"
    echo "or use the shared marshalOpts / prettyMarshalOpts variables."
    exit 1
fi

echo "json-deterministic-check: all production marshal calls include Deterministic."
exit 0
