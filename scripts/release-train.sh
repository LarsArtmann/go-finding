#!/usr/bin/env bash
# Release train driver: one command per train stage, in dependency order.
# Born from five manual trains (v1.13.0 -> v1.15.1) where every stage lived
# in the operator's head and two trains shipped defects the checklist was
# supposed to catch (stale stamp, skipped stress).
#
# Stages (run in this order):
#   prep       structural preflight (clean tree, pushed commits, drift, docs)
#   verify     prep + compare-link verification (network, gh auth)
#   stress     the repeat=20 race stress gate (mandatory unless the train
#              log carries a WRITTEN no-op justification, d/1 rule)
#   tag VER    stamp gate, then create all 5 annotated tags LOCALLY (no push)
#   push VER   push master, then ONE tag per push, each Release run watched
#              to a terminal state (serialized; reruns strictly one at a time)
#   resync VER bump cmd/go-finding requires, tidy, go.sum completion; prints
#              the vendorHash instructions (nix hash is operator-applied)
#   postflight VER  the post-release verifier (proxy x5, clean-dir go get x5,
#              release assets, pkg.go.dev backoff, stamp-vs-tag)
#   dry        print the plan for the current tree state without side effects
#
# The script NEVER pushes without the operator invoking the push stage
# explicitly, and never deletes or retags release tags. Usage:
#
#   ./scripts/release-train.sh dry
#   ./scripts/release-train.sh prep
#   ./scripts/release-train.sh tag 1.16.0          # version WITHOUT the v
#   ./scripts/release-train.sh postflight 1.16.0
set -euo pipefail

cd "$(dirname "$0")/.."

verdict() { printf '\n== %s: %s ==\n' "$1" "$2"; }
die() { echo "TRAIN FAIL: $*" >&2; exit 1; }

usage() {
	sed -n '2,30p' "$0" | grep -E '^#\s+(prep|verify|stress|tag|push|resync|postflight|dry)' |
		sed 's/^#   //'
	exit 2
}

core_version() {
	local major minor patch
	major=$(sed -n 's/^const VersionMajor = //p' version.go)
	minor=$(sed -n 's/^const VersionMinor = //p' version.go)
	patch=$(sed -n 's/^const VersionPatch = //p' version.go)
	echo "${major}.${minor}.${patch}"
}

watch_release_run() {
	local tag="$1"
	local run_id
	run_id=$(gh run list --workflow=release.yml --limit 5 --json databaseId,headBranch,status \
		--jq ".[] | select(.headBranch==\"$tag\") | .databaseId" | head -1)
	if [ -z "$run_id" ]; then
		echo "FAIL: no Release run found for $tag (the v1.9.0 silent-drop class)."
		echo "Check nothing is queued, then: gh workflow run release.yml"
		return 1
	fi
	echo "watching Release run $run_id for $tag"
	gh run watch "$run_id" --exit-status
}

stage_prep() {
	verdict prep "running structural preflight"
	bash scripts/release-preflight.sh
}

stage_verify() {
	stage_prep
	verdict verify "compare links (authenticated gh api)"
	bash scripts/release-preflight.sh --links >/dev/null || {
		echo "FAIL: compare links do not all resolve"
		return 1
	}
	echo "OK: compare links resolve"
}

stage_stress() {
	verdict stress "repeat=20 race gate (mandatory; d/1 rule)"
	echo "core + pipeline: ginkgo -r --race --repeat=20 --skip-package=examples"
	echo "analysis + cmd:  go test -race -count=20 ./..."
	echo "Run bash scripts/release-preflight.sh --stress to execute through the gate."
	echo "A no-op justification MUST already be written into the train log to skip this."
}

stage_tag() {
	local ver="${1:?usage: release-train.sh tag <X.Y.Z>}"
	stage_verify

	verdict tag "stamp gate for all 5 module tags"
	for prefix in "" "pipeline/" "analysis/" "toolsdk/" "cmd/go-finding/"; do
		bash scripts/stamp-gate.sh "${prefix}v${ver}"
	done

	verdict tag "creating 5 annotated tags locally (NO push)"
	git tag -a "v${ver}" -m "Release v${ver}"
	git tag -a "pipeline/v${ver}" -m "pipeline v${ver}"
	git tag -a "analysis/v${ver}" -m "analysis v${ver}"
	git tag -a "toolsdk/v${ver}" -m "toolsdk v${ver}"
	git tag -a "cmd/go-finding/v${ver}" -m "cli v${ver}"
	echo "OK: tags created. NOTE: cmd/go-finding/v${ver} must be RE-CREATED at the"
	echo "resync commit before pushing (stage resync recreates it; stage push enforces order)."
}

stage_push() {
	local ver="${1:?usage: release-train.sh push <X.Y.Z>}"
	verdict push "master first, then ONE tag per push, each watched"
	git push origin master

	# Library tags first (core, pipeline, analysis, toolsdk).
	for tag in "v${ver}" "pipeline/v${ver}" "analysis/v${ver}" "toolsdk/v${ver}"; do
		if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
			echo "pushing $tag"
			git push origin "$tag"
			watch_release_run "$tag"
		else
			echo "skip: tag $tag does not exist locally"
		fi
	done

	die "stop: push the resync commit (stage resync) and re-create cmd/go-finding/v${ver} at it, THEN run: git push origin cmd/go-finding/v${ver} && watch_release_run manually"
}

stage_resync() {
	local ver="${1:?usage: release-train.sh resync <X.Y.Z>}"
	verdict resync "cmd/go-finding requires -> v${ver} siblings"
	(cd cmd/go-finding &&
		go get "github.com/larsartmann/go-finding@v${ver}" &&
		go get "github.com/larsartmann/go-finding/pipeline@v${ver}" &&
		GOWORK=off go mod tidy)
	GOWORK=off go mod tidy
	echo "next: update flake.nix vendorHash (buildflow -s nix-hash-fix --fix, or take the"
	echo "hash the CI nix job prints), commit, push, wait for green CI, then re-create"
	echo "cmd/go-finding/v${ver} at THIS commit before pushing it."
}

stage_postflight() {
	local ver="${1:?usage: release-train.sh postflight <X.Y.Z>}"
	verdict postflight "proxy resolution x5"
	local failed=0
	for mod in "" "pipeline/" "analysis/" "toolsdk/" "cmd/go-finding/"; do
		if go list -m "github.com/larsartmann/go-finding/${mod}@v${ver}" >/dev/null 2>&1; then
			echo "OK: proxy has ${mod:-core} v${ver}"
		else
			echo "FAIL: proxy does not have ${mod:-core} v${ver} (yet)"
			failed=1
		fi
	done

	verdict postflight "clean-dir go get x5"
	local scratch
	scratch=$(mktemp -d)
	for mod in "" "/pipeline" "/analysis" "/toolsdk" "/cmd/go-finding"; do
		if (cd "$scratch" && go mod init postflight-consumer >/dev/null 2>&1; cd "$scratch" && go get "github.com/larsartmann/go-finding${mod}@v${ver}"); then
			echo "OK: clean-dir go get ${mod:-/core}"
		else
			echo "FAIL: clean-dir go get ${mod:-/core}"
			failed=1
		fi
	done
	rm -rf "$scratch"

	verdict postflight "release assets + stamp"
	gh release view "v${ver}" --json assets --jq '.assets | length' |
		xargs -I{} echo "core release assets: {}"
	if gh release view "cmd/go-finding/v${ver}" --json assets --jq '.assets | length' >/dev/null 2>&1; then
		echo "OK: cmd release exists"
	else
		echo "FAIL: cmd release missing"
		failed=1
	fi

	verdict postflight "pkg.go.dev render (one backoff retry)"
	for mod in "" "/pipeline" "/analysis" "/toolsdk" "/cmd/go-finding"; do
		local url="https://pkg.go.dev/github.com/larsartmann/go-finding${mod}@v${ver}"
		local code
		code=$(python3 -c "import urllib.request,sys; print(urllib.request.urlopen(urllib.request.Request(sys.argv[1])).status)" "$url" 2>/dev/null || echo 0)
		if [ "$code" != "200" ]; then
			sleep 30 # pkg.go.dev crawl backoff
			code=$(python3 -c "import urllib.request,sys; print(urllib.request.urlopen(urllib.request.Request(sys.argv[1])).status)" "$url" 2>/dev/null || echo 0)
		fi
		if [ "$code" = "200" ]; then
			echo "OK: $url"
		else
			echo "WARN: $url -> $code (crawl lag; re-check later, GET the page once to re-trigger)"
		fi
	done

	verdict postflight "stamp vs tag"
	local printed=""
	if [ -x "$(go env GOPATH)/bin/go-finding" ]; then
		printed="$("$(go env GOPATH)/bin/go-finding" --version)"
		echo "installed CLI reports: $printed (tag says ${ver})"
		[ "$printed" = "$ver" ] || failed=1
	else
		echo "go install github.com/larsartmann/go-finding/cmd/go-finding@v${ver} first, then re-run"
		failed=1
	fi

	if [ "$failed" -ne 0 ]; then
		die "postflight found failures above"
	fi
	echo "TRAIN PASS: v${ver} fully published and verified."
}

stage_dry() {
	verdict dry "train plan for the current tree"
	echo "core version.go: v$(core_version)"
	local pending
	pending=$(git log '@{u}..HEAD' --oneline 2>/dev/null | wc -l | tr -d ' ')
	echo "unpushed commits: $pending"
	echo "existing v$(core_version) tag family:"
	for prefix in "" "pipeline/" "analysis/" "toolsdk/" "cmd/go-finding/"; do
		local t
		t="${prefix}v$(core_version)"
		git rev-parse -q --verify "refs/tags/$t" >/dev/null &&
			echo "  EXISTS: $t" ||
			echo "  new:    $t"
	done
	echo "stages: prep -> verify -> stress -> tag -> push(master+libs) -> resync -> push(cmd) -> postflight"
	echo "(dry run: nothing executed)"
}

case "${1:-}" in
prep) stage_prep ;;
verify) stage_verify ;;
stress) stage_stress ;;
tag) shift
	stage_tag "$@" ;;
push) shift
	stage_push "$@" ;;
resync) shift
	stage_resync "$@" ;;
postflight) shift
	stage_postflight "$@" ;;
dry) stage_dry ;;
*) usage ;;
esac
