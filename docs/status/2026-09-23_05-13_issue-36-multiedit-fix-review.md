# Status: Issue #36 Review → Multi-Edit Fix Implementation (2026-09-23 05:13)

Session scope: "Review all open GitHub issues." Exactly one open issue existed:
[#36 — FromDiagnosticWithSource drops TextEdits 2..N (multi-edit fixes are lossy)](https://github.com/LarsArtmann/go-finding/issues/36).
The review verified every claim against source, then implemented Option A (the
plan of record): carry typed edits end-to-end (diagnostic → finding → applier).

## a) FULLY DONE (verified green)

### Review & verification

- All three compounding loss spots from the issue confirmed in source:
  `analysis/analysis.go` kept only `TextEdits[0]`; `Finding` was single-edit
  (`BeforeCode`/`AfterCode`); `ToDiagnostic` reconstructed exactly one TextEdit.
- Key architectural insight: `FixProvider.Edits` already returns `[]FixEdit`
  per finding — only the Finding model and the bridge were single-edit.

### Core module (`github.com/larsartmann/go-finding`)

- New `text_edit.go`: `TextEdit{Start, End Position; NewText string}` mirroring
  go/analysis / LSP TextEdit. Methods: `HasSpan`, `IsInsertion`, `IsDeletion`,
  `EffectiveFile`, `Validate` (sentinel errors: `errEditNoLocation`,
  `errEditInverted`, `errEditCrossFile`, `errEditNegOffset`,
  `errEditOffsetCross`), `Equal`, `Compare`, package `editsEqual`
  (order-independent list equality).
  - Conventions pinned: insertion = End unset (`Offset: -1`) OR End == Start
    (go/analysis style); `Position{}` = byte 0 (zero-value convention); empty
    `Start.File` = finding's file.
- `Finding.Edits []TextEdit` (`json:"edits,omitempty"`), authoritative when set;
  BeforeCode/AfterCode remain first-edit display summary.
- `HasCodeChange`, `HasFix`, `IsAutoFixable`, new `HasEditList` all recognize
  edit lists; `validateFix` accepts Edits for FixStrategyDirect and validates
  each edit (`finding.Edits[N] is invalid: <cause>`).
- `Equal` includes Edits (order-independent, mirroring tags semantics).
- `Builder.WithEdits(...)`.
- SARIF round-trip: `sarifRegion` gained pointer `ByteOffset`/`ByteLength`
  (byte offset 0 distinguishable from unset under omitempty); export emits one
  artifact change per target file with ALL replacements (`sarifEditRegion`);
  import rebuilds the full edit list from every artifact change
  (`editsFromSarifChanges`).
- `ApplySimpleFixes` refuses `len(Edits) > 1` findings with actionable reason
  ("use pipeline.FixApplier") instead of half-applying edit 0.

### Analysis module

- `FromDiagnosticWithSource`/`FromDiagnostic` carry ALL TextEdits of
  `SuggestedFixes[0]` into `Finding.Edits`; BeforeCode/AfterCode = edit 0.
  Doc comments rewritten (remaining loss: SuggestedFixes 1..N are
  alternatives; only the first is carried — now documented).
- `ToDiagnostic` emits every edit via `textEditsFromFinding` (insertions
  resolve End→Pos; unresolvable positions skipped); legacy single-edit path
  retained for Edits-less findings.

### Pipeline module

- New `EditListProvider` (`edit-list`), FIRST in the default chain (now 4
  providers). Verbatim byte-level application; line/col fallback through the
  shared line offset index (`lineIndexAware`); refuses loudly instead of
  half-applying: cross-file lists → `ErrEditCrossFile`, stale/out-of-bounds
  offsets → `ErrEditStale`, unresolvable position → error (outcome Failed).
- Accounting fix: `dedupAppliedFindings` in `FixEngine.apply` — `Applied` /
  `AppliedFixes` now describe FINDINGS (one entry per fix; was one per EDIT,
  which a 2-edit finding inflated to 2); `AppliedEdits` keeps per-edit
  granularity. First dedup attempt (composite key) collapsed distinct ID-less
  findings — caught by existing `TestFixEngine_MixedProviders`, replaced with
  full-equality dedup (O(n²) on findings count, fine at per-file scale).

### Tests (issue's three verification items all covered)

- Core: `text_edit_test.go` (validate table, sentinels, equal/compare,
  editsEqual order-independence, builder, fix predicates, JSON round-trip,
  SARIF round-trip incl. byte-0 insertion), validate cases in
  `finding_valid_test.go`, multi-edit refusal in `simple_fix_test.go`.
- Analysis: `TestFromDiagnosticWithSource_MultiEditFixCarriesAllEdits` (item 1),
  `TestToDiagnostic_MultiEditRoundTrip` (item 2, positions + text),
  `TestToDiagnostic_InsertionEditRoundTrip` (import-insertion variant).
- Pipeline: `TestEditListProvider_AppliesAllEditsInOnePass` (erraudit
  legacyerrors shape: declaration removal + condition rewrite),
  `LineColumnOnlyEdits`, `CrossFileEditListRefused`, `StaleOffsetsRefused`,
  `PureInsertion`; `TestFixApplier_AppliesMultiEditFinding` (item 3 — both
  edits land on disk in one pass through `FixApplier.Apply`).
- Updated for new reality: `TestFixEngine_Providers` (4 providers),
  validate message test.

### Gates run (all green)

- Tests: workspace mode + `GOWORK=off` per module for ALL 5 modules; race for
  core/gotoken/lockutil, analysis, pipeline(+goast); CLI module build+tests.
- `nix run .#lint`: 0 issues (after `t.Parallel` fix + `golangci-lint fmt` +
  `nix fmt`; formatter touched 2 files, re-tested green).
- `json-deterministic-check.sh`, `replace-audit.sh`, `test-naming.sh`,
  `go-work-sync.sh`, `docs-api-check.sh` (318 identifiers): OK.
- `version-check.sh`: OK — after repairing pre-existing damage: the v1.13.0
  tag was cut with `version.go` still at 1.12.0; bumped `VersionMinor` to 13
  and the two README stamps (docs-api-check + version-check now pass; 3 of 4
  version-drift errors gone).

## b) PARTIALLY DONE

- `version-drift.sh`: 1 residual error — `cmd/go-finding/go.mod` requires
  `pipeline v1.12.0` while master expects v1.13.0. Root cause: the v1.13.0
  release is INCOMPLETE — only `v1.13.0` (core) and `toolsdk/v1.13.0` were
  pushed; `pipeline/v1.13.0`, `analysis/v1.13.0`, `cmd/go-finding/v1.13.0` do
  not exist (matches the documented tag-batching/Release-cancellation gotchas
  in AGENTS.md). Completing it requires pushing tags → operator decision, not
  done (pushing is out of bounds without explicit instruction).

## c) NOT STARTED (was queued in this session's plan)

- Docs: ADR #17 entry (Finding carries typed edit list), CHANGELOG
  (Unreleased), `docs/guides/fix-engine.md` provider-chain update (now 4),
  `doc.go` fix/TextEdit mention, AGENTS.md gotcha updates (provider chain,
  Applied-per-finding accounting).
- Issue #36: closing comment (github-voice skill loaded at write time, not yet).
- Gates not yet run: `error-audit.sh` (erraudit binary availability?),
  go-arch-lint, `nix flake check`, stress gate (`ginkgo --repeat=20 --race`
  core+pipeline, `go test -race -count=20` analysis+CLI) — mandatory before
  any tag, not yet exercised this session.
- Benchmarks: not run (dedup is O(n²) on applied-findings count; expected
  noise-level, unmeasured).

## d) TOTALLY FUCKED UP? — nothing unrecoverable

Self-caught and fixed during the session (kept for honesty):

1. First dedup design (composite key id+pos+message) collapsed distinct
   ID-less findings — existing test caught it; replaced with equality dedup.
2. Initially designed "zero End = insertion", contradicting the project's
   Position zero-value convention (Position{} = byte 0). Realigned: insertion
   = End unset (Offset −1) or End == Start; doc + tests corrected.
3. Smaller: dead stub left in SARIF export (removed), duplicated Validate
   branch (removed), eaten comment line in analysis.go (restored), two wrong
   hardcoded test expectations (computed values instead), one unused import
   iteration loop.
4. Residual known gap (documented, not fixed): `Conflict.ConflictsWith` can
   still list the same finding twice when multi-edit overlaps occur — cosmetic,
   not touched.

## e) WHAT WE SHOULD IMPROVE (observations from this session)

1. Release hygiene: version.go/tag stamping slipped at v1.13.0 AND three
   module tags never landed; `release-preflight.sh` should grow a "tag set
   completeness" check (all 5 module tags present for the release commit).
2. The CLI-pinned-to-published-siblings constraint (no replace directives)
   again delays consumer adoption of new sibling APIs (erraudit T13/T14 needs
   `analysis`+`pipeline` tags before it can migrate). Consider whether the
   next release should be cut promptly with this feature.
3. `FixApplier.Apply`/`ApplyWithDetails` docs say "number of successful
   fixes"/"applied findings" — the engine previously returned per-EDIT counts;
   the docs were right, the code was wrong. Worth an assertion-level test of
   count semantics (now implicitly covered).
4. erraudit CI wiring still blocked (private repo) per AGENTS.md — unchanged.

## f) NEXT UP TO 50 THINGS

Docs & issue (immediate):

1. ADR #17: typed edit list on Finding (motivation, alternatives, conventions).
2. CHANGELOG Unreleased entry (core+analysis+pipeline; breaking? no — additive;
   Applied-count semantics change noted).
3. `docs/guides/fix-engine.md`: add EditListProvider to the chain table/flow.
4. `docs/guides/outcomes.md`: mention multi-edit findings & failed outcomes for
   cross-file/stale lists.
5. `doc.go`: Finding.Edits + EditListProvider in the fix section.
6. README: multi-edit fixes section (bridge → applier, one code sample).
7. AGENTS.md: provider chain now 4 (edit-list first); Applied = per finding;
   ApplySimpleFixes refuses multi-edit lists.
8. Comment + close issue #36 with verification evidence (github-voice).
9. Note in issue: consumer adoption requires analysis+pipeline tags.

Quality gates to finish:
10. `error-audit.sh` (all 5 modules) — new code should pass (sentinels + %w).
11. go-arch-lint (`nix` or script).
12. `nix flake check` (vendorHash unaffected — no dep changes; verify).
13. Stress gate: ginkgo --repeat=20 --race (core, pipeline).
14. Stress gate: go test -race -count=20 (analysis, CLI).
15. Bench spot-check (fix engine apply + dedup path) vs baseline.
16. `docs-freshness.sh` after doc edits.
17. Re-run full lint + fmt after doc edits (dprint covers markdown).

Release (operator decisions):
18. Decide: complete v1.13.0 (push 3 missing tags + resync CLI go.mod) vs fold
into v1.14.0 containing this feature.
19. If v1.14.0: bump version.go, CHANGELOG cut, preflight (incl. --bench),
tag push in batches ≤3, rerun cancelled Release runs ONE AT A TIME.
20. After tags: resync CLI go.mod to new sibling versions; resync commit.
21. Verify pkg.go.dev renders new TextEdit/Edits APIs.

Follow-up engineering (candidates, not committed):
22. Conflict.ConflictsWith dedup (same finding twice on multi-edit overlap).
23. `ApplySimpleFixes`: byte-level application of single-edit lists (offsets
known) instead of string matching — or keep refusing (decide).
24. LSP: expose Edits as LSP CodeAction TextEditEdits (LSPDiagnosticData
currently single-fix).
25. Multi-FILE fix support in FixApplier (currently refused loudly; would need
per-file grouping + atomic cross-file rollback policy).
26. SuggestedFixes 1..N (alternatives) representation on Finding (e.g.
AlternativeFixes) — documented as lossy today.
27. erraudit T13/T14 migration once tags exist (upstream repo work).
28. gogenfilter/generated-file awareness for edit-list findings.
29. GoASTProvider: emit Finding.Edits (typed) instead of raw FixEdits, so CLI
dry-runs and SARIF export benefit.
30. Fuzz: TextEdit.Validate + EditListProvider against random coordinates.
31. Benchmark: editsEqual allocation profile on large edit lists.
32. Examples: pipeline/examples/multi-edit showing bridge → applier.
33. Consider `Edits` in `Preview()` (multi-hunk preview).
34. Consider `GenerateID` stability guarantees doc (Edits excluded — deliberate).
35. docs/DOMAIN_LANGUAGE.md: "edit list", "insertion", "span" entries if file
exists for it.
36. Release preflight selftest: add incomplete-tag-set failure class.
37. CLI flag `-fix-provider edit-list` docs (it's default now; registry names).
38. Update `docs/PRO_CONTRA_go-output-integration.md`? No — unrelated; skip.

(a few intentionally spurious-sounding items trimmed; the honest bounded list
above is what I'd actually do.)

## g) QUESTIONS (cannot be answered from the repo)

1. Issue #36 lifecycle: close now (fix verified on master) or keep open until
   the fix is consumable (needs `analysis/v*` + `pipeline/v*` tags) so erraudit
   can pin it?
2. Release completion: should the 3 missing v1.13.0 module tags be pushed to
   finish that release, or is everything folded into a prompt v1.14.0 that
   ships this feature? (Tag pushes need your go-ahead; I won't push.)
3. `ApplySimpleFixes` philosophy: keep refusing multi-edit lists (current,
   safe) or teach it byte-level application so the no-pipeline path is also
   lossless?

## h) ADDENDUM — residual gates + fallout fixes (same day, follow-up pass)

All remaining gates ran (next-up items 10-17 done). Fallout found and fixed:

- **erraudit / go-arch-lint**: 0 violations, no warnings — green as-is.
- **Stress gates**: `ginkgo -r --race --repeat=20 --skip-package=examples`
  green in core (10 suites, 5m42s) and pipeline (3 suites, 1m26s);
  `GOWORK=off go test -race -count=20 ./...` green in analysis (4.2s) and
  CLI (26.6s + detectors 5.7s).
- **`nix flake check` was BROKEN since the Sep 20-22 go 1.27 go.mod bump**
  (nobody ran it locally; CI has no nix jobs — a silent dead gate). Three
  causes, all fixed in `flake.nix`:
  1. `packages.default`/overlay still built with nixpkgs default go (1.26.7,
     GOTOOLCHAIN=local) → "go.mod requires go >= 1.27". Fixed: both call
     sites now `buildGoModule.override { go = pkgs.go_1_27; }`.
  2. `vendorHash` changed under go 1.27 (proxyVendor output differs across
     toolchains) → regenerated.
  3. treefmt-check: `goimports` (nixpkgs gotools, built with go 1.26.7)
     embeds its build-time GOROOT and falls back to that `go` when none is
     on PATH; go.mod's 1.27 floor made it attempt a GOTOOLCHAIN download —
     fatal in the offline sandbox. Outside the sandbox the same fallback
     silently DOWNLOADED a full go1.27.0 toolchain into the repo
     (`./go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.0...` — still sitting
     there, untracked, read-only; candidate for `trash`, operator call).
     Fixed: `programs.goimports.package` is a wrapper exporting
     `PATH=go_1_27/bin` + `GOTOOLCHAIN=local` before exec'ing real goimports.
     After all three: `nix flake check` fully green (plain form, per the
     documented `--all-systems` decision).
- **Benchmark gate FAILED on first run — real regression, fixed properly.**
  The `dedupAppliedFindings` full-equality O(n²) dedup cost +288%..+706%
  time on applier/GoAST benchmarks (GoAST_1000: 974µ → 7853µ). Replaced with
  O(1) owner-index tracking: `apply()` records each edit's owning input
  finding (`editOwner`, parallel to `allEdits`) and
  `applyEditsWithConflicts` appends a finding to `Applied` once, on its
  first applied edit. Same per-finding semantics, no post-hoc dedup.
  Residual intentional cost: `Finding` grew by the `Edits` slice header
  (+24 B/copy) → +11..25% B/op on Finding-copy benchmarks (GroupByFile,
  Correlate, MergeIter/Combine) with flat allocs/op — the struct-growth
  class of the v1.7.0 GroupID precedent; baseline regenerated with
  rationale in `benchmarks/README.md`.
- **test-naming gate was red** (pre-existing, not from this fix):
  `summary_coverage_test.go` (another session's file) violates the
  "never name files after metrics" rule. Merged its two Summary-contract
  tests into `report_test.go` (Summary lives in report.go) and `git rm`'d
  the file. Gate green.
- **Docs written** (next-up items 1-7): CHANGELOG entries (root + pipeline +
  analysis Unreleased), ADR (typed edit lists — note: the plan-of-record
  said "#17" but the log already had #17 and #18, so it landed as **#19**),
  fix-engine guide (multi-edit section + 4-provider chain table), doc.go
  (TextEdit/EditListProvider), README (multi-edit section + updated default
  chain), AGENTS.md (core multi-edit rules, pipeline Applied-per-finding
  gotcha, ADR pointer, text_edit.go in key files). outcomes.md left as-is
  (provider-agnostic already; multi-edit failures surface as ordinary
  `failed` outcomes).
- Doc gates after edits: docs-api-check (318 identifiers OK),
  docs-freshness OK, dprint clean, version-check OK, replace-audit OK,
  go-work-sync OK, json-deterministic OK.
- **Known residual (operator-blocked)**: `version-drift.sh` still reports
  `cmd/go-finding requires pipeline v1.12.0, expected v1.13.0` — resolves
  only when the missing v1.13.0 sibling tags are pushed (question 2).

---

_Assisted-by: Crush <crush@charm.land>_
