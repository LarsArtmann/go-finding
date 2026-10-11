#!/usr/bin/env bash
# Check relative links in markdown files resolve to existing files.
# Extracted from ci.yml (markdown-link-check job) so it is runnable locally
# and testable; the workflow just calls this.
#
# Code is stripped first (fenced blocks, indented blocks, inline spans):
# generic signatures like `NewIntervalIndex[T](intervals)` otherwise parse
# as links. \x60 = backtick (avoids shellcheck SC2016).
#
# Usage: ./scripts/link-check.sh
set -euo pipefail

cd "$(dirname "$0")/.."

rc=0
while IFS= read -r file; do
	while IFS= read -r link; do
		# Skip URLs and anchors
		[[ "$link" =~ ^https?:// ]] && continue
		[[ "$link" =~ ^# ]] && continue
		# Strip anchor from end
		target="${link%%#*}"
		[ -z "$target" ] && continue
		# Skip targets with spaces (code remnants)
		[[ "$target" == *" "* ]] && continue
		# Resolve relative to the markdown file's directory
		dir=$(dirname "$file")
		fullpath="$dir/$target"
		if [ ! -e "$fullpath" ]; then
			echo "::error file=$file::Broken link: $link (resolved: $fullpath)"
			rc=1
		fi
	done < <(awk '/^```/{f=!f;next} f{next} /^ {4}/{next} {print}' "$file" |
		sed 's/\x60[^\x60]*\x60//g' |
		grep -oP '\[.*?\]\(\K[^)]+' 2>/dev/null)
done < <(find . -name "*.md" -not -path "./vendor/*")
exit "$rc"
