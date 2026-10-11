# Launch Announcement Draft (LAUNCH2)

> **Status:** READY (finalized 2026-09-23). Repo is PUBLIC, Latest release is
> v1.13.0 with full GoReleaser assets, master CI is green (23 jobs incl. nix +
> consumer-compat), pkg.go.dev renders all 5 modules, and the consumer-compat
> gate resolves/builds/runs every module from the proxy. Toolchain note: Go
> 1.27+ floor, and since json/v2 went GA there is NO GOEXPERIMENT setup.
> Pending: v1.14.0 train (multi-edit feature, issue #36) and the actual
> Awesome Go submission (needs an awesome-go fork + PR).

## Short version (X/Mastodon/Slack line)

go-finding: one Finding data model + fix pipeline for Go static analysis. Plug govet, staticcheck, art-dupl (or your own detector) into a detect → triage → fix → verify loop with per-finding outcomes, typed multi-edit fixes, per-file rollback, and SARIF/LSP interop. Now at v1.13.0. https://github.com/LarsArtmann/go-finding

## Announcement post (blog / r/golang)

**One data model for Go linting results — and a pipeline that actually fixes them**

Every Go linter reinvents the same plumbing: how to represent a finding, how to serialize it (SARIF? JSON? LSP?), how to dedupe across tools, and — the part nobody finishes — how to turn diagnostics into applied fixes without wrecking the file.

`go-finding` is a library that does that plumbing once:

- **A unified `Finding` type** — position, severity, confidence, categories, tags, related findings, groups, suppression, and fix payloads (`BeforeCode`/`AfterCode` or AST edits). Immutable data, strongly typed (`ID`, `RuleName`, `FilePath` are distinct types, not strings).
- **Lossless interchange** — hand-rolled SARIF export/import, LSP diagnostics with full round-trip fidelity, deterministic JSON (byte-identical output, enforced in CI).
- **A fix engine that reports honestly** — since v1.7.0 every finding gets an outcome: `applied`, `refused`, `conflict`, `invalid`, `failed`. A provider that declines is no longer indistinguishable from success.
- **Per-file rollback by default** — one bad file no longer discards the clean fixes from every other file (the behavior that motivated v1.7.0). All-or-nothing is one flag away.
- **Grouped findings** — `GroupID` marks clone groups (art-dupl's N duplicated blocks become one logical set through SARIF, LSP, and JSON).

It is multi-module on purpose: `go-finding` (core, tiny dep surface), `go-finding/pipeline` (detect → fix → verify loop, metrics, flight recorder), `go-finding/analysis` (go/analysis bridge), `go-finding/cmd/go-finding` (CLI with govet/staticcheck built in).

```go
engine := pipeline.NewFixEngine()
result := engine.ApplyWithOutcomes(content, findings)
for _, o := range result.Outcomes {
    fmt.Println(o.Finding.ID, o.Status) // applied, refused, ...
}
```

Docs: `docs/guides/` (fix engine, outcomes, groups, flight recorder, migration guides). ADRs in `docs/architecture-decisions.md`. MIT.

Feedback welcome — especially from tool authors who want their linter on the fix pipeline.

## Awesome Go submission entry (LAUNCH3, when public)

- **Name:** go-finding
- **Description:** Unified data model and detect→fix pipeline for Go static analysis findings. SARIF/LSP interop, per-finding fix outcomes, grouped findings.
- **Category:** Code Analysis (or Linters/Tools)
- **Repo:** https://github.com/LarsArtmann/go-finding

## Launch checklist state (2026-09-23)

1. ~~Flip repo visibility to public; remove `GOPRIVATE`~~ DONE 2026-09-08.
2. ~~pkg.go.dev rendering for all modules~~ DONE — all 5 modules render; v1.13.0 verified via the consumer-compat gate (proxy resolve + build + run + CLI install for all 5).
3. GoReleaser assets: v1.13.0 published with 17 assets (5 platform archives, 8 packages, 5 SBOMs, checksums); sigstore signing was skipped for this recovery release (cosign keyless is CI-only) — restore full signing on the next workflow-run train.
4. Homebrew tap: still pending `LarsArtmann/homebrew-tap` + `HOMEBREW_TAP_GITHUB_TOKEN` (ROADMAP open question).
5. Awesome Go: submit the entry above (fork + PR) when ready.
6. Submit to Awesome Go (entry above) (LAUNCH3).
7. Publish the announcement (drafts above) (LAUNCH2).
