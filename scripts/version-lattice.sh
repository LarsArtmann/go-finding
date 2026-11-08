#!/usr/bin/env bash
# Version-lattice guard: before a train, assert the tag/module lattice is
# coherent — what the PUBLISHED tags require is what consumers can resolve.
#
# The lattice rules:
#   1. cmd/go-finding's require of pipeline == pipeline's latest tag (the
#      CLI must build against the engine the train ships).
#   2. Each sibling's latest tag must require a core version that EXISTS as
#      a tag. This is the true consumer-broken class: the v1.14.0 train
#      tagged a sibling whose require resolved to nothing.
#   3. A sibling whose tagged core require LAGS the working tree's require
#      is EXPECTED under independent versioning (toolsdk ships on its own
#      schedule; its v1.16.0 legitimately requires core v1.14.0 while the
#      tree moved on) — printed as a warning, never a failure.
#
# Usage:
#   ./scripts/version-lattice.sh             # the real lattice
#   ./scripts/version-lattice.sh --selftest  # prove the FAIL branch
#
# Exit codes: 0 = coherent (warnings allowed), 1 = incoherent.
set -euo pipefail

cd "$(dirname "$0")/.."

FAILURES=0
step() { printf '\n== %s ==\n' "$1"; }

tag_exists() { git rev-parse -q --verify "refs/tags/$1" >/dev/null; }

core_require_of() { # module directory, tree | tag:<tag>
	local mod="$1" source="$2"
	if [ "$source" = "tree" ]; then
		sed -n 's|.*github.com/larsartmann/go-finding \(v[0-9.]*\)$|\1|p' "$mod/go.mod" | head -1
	else
		git show "${source#tag:}:$mod/go.mod" 2>/dev/null |
			sed -n 's|.*github.com/larsartmann/go-finding \(v[0-9.]*\)$|\1|p' | head -1
	fi
}

check_module_lattice() { # mod latest_tag — shared by live run and selftest
	local mod="$1" latest="$2" tagged_require
	tagged_require=$(core_require_of "$mod" "tag:$latest")
	if [ -z "$tagged_require" ]; then
		echo "SKIP: $latest carries no core require line (independent module)"
		return
	fi
	if tag_exists "$tagged_require"; then
		local work_require
		work_require=$(core_require_of "$mod" tree)
		if [ "$tagged_require" = "$work_require" ]; then
			echo "OK: $mod latest tag $latest requires core $tagged_require (matches working tree)"
		else
			echo "WARN: $mod latest tag $latest requires core $tagged_require, working tree requires $work_require — expected under independent versioning; the NEXT release of $mod carries the newer require"
		fi
	else
		echo "FAIL: $mod latest tag $latest requires core $tagged_require, which does NOT EXIST as a tag — every consumer resolving that tag gets a build-breaking require (the v1.14.0 incident class)"
		return 1
	fi
}

if [ "${1:-}" = "--selftest" ]; then
	step "selftest: unresolvable require must FAIL"
	# Build a throwaway tree (git plumbing; working tree untouched) whose
	# selftest/go.mod requires core v9.9.9 — a version with no tag — and
	# point a throwaway tag at it. mktree needs a nested subtree for the
	# selftest/ directory.
	blob=$(printf 'module selftest\n\ngo 1.27\n\nrequire github.com/larsartmann/go-finding v9.9.9\n' | git hash-object -w --stdin)
	subtree=$(printf '100644 blob %s\tgo.mod\n' "$blob" | git mktree)
	tree=$(printf '040000 tree %s\tselftest\n' "$subtree" | git mktree)
	commit=$(printf 'lattice selftest' | git commit-tree "$tree")
	git tag -a -m "lattice selftest" "lattice-selftest/v0.0.1" "$commit"
	check_ok=0
	check_module_lattice "selftest" "lattice-selftest/v0.0.1" || check_ok=1
	git tag -d "lattice-selftest/v0.0.1" >/dev/null
	if [ "$check_ok" -eq 0 ]; then
		echo "FAIL: selftest — the unresolvable require was NOT caught (gate is dead)"
		exit 1
	fi
	echo "OK: selftest — unresolvable requires fail the lattice"
	exit 0
fi

step "cmd/go-finding require of pipeline == pipeline's latest tag"
pipeline_latest=$(git tag -l 'pipeline/v*' --sort=-v:refname | head -1)
cmd_require=$(sed -n 's|.*github.com/larsartmann/go-finding/pipeline \(v[0-9.]*\).*|\1|p' cmd/go-finding/go.mod)
if [ -z "$pipeline_latest" ] || [ -z "$cmd_require" ]; then
	echo "FAIL: could not parse pipeline latest ('$pipeline_latest') or cmd require ('$cmd_require')"
	FAILURES=$((FAILURES + 1))
elif [ "$pipeline_latest" = "pipeline/$cmd_require" ]; then
	echo "OK: cmd requires $cmd_require, pipeline latest is $pipeline_latest"
else
	echo "FAIL: cmd requires $cmd_require but pipeline's latest tag is $pipeline_latest — the CLI would build against an older engine than the train ships"
	FAILURES=$((FAILURES + 1))
fi

step "each sibling's latest tag requires an EXISTING core tag version"
for mod in pipeline analysis toolsdk cmd/go-finding; do
	latest=$(git tag -l "$mod/v*" --sort=-v:refname | head -1)
	if [ -z "$latest" ]; then
		echo "SKIP: $mod has no tags yet"
		continue
	fi
	if ! check_module_lattice "$mod" "$latest"; then
		FAILURES=$((FAILURES + 1))
	fi
done

core_stamp="v$(sed -n 's/^const VersionMajor = //p' version.go).$(sed -n 's/^const VersionMinor = //p' version.go).$(sed -n 's/^const VersionPatch = //p' version.go)"
echo ""
echo "LATTICE VERDICT: $([ "$FAILURES" -eq 0 ] && echo COHERENT || echo "INCOHERENT ($FAILURES relation(s) broken)") (core stamp: $core_stamp)"
exit "$FAILURES"
