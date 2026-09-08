# Launch Announcement Draft (LAUNCH2)

> **Status:** DRAFT — parked until the repo goes public. Content derived from
> FEATURES.md and CHANGELOG v1.7.0 (verified 2026-09-08). Adapt tone/length
> per channel. Do not publish while the repo is PRIVATE.

## Short version (X/Mastodon/Slack line)

go-finding: one Finding data model + fix pipeline for Go static analysis. Plug govet, staticcheck, art-dupl (or your own detector) into a detect → triage → fix → verify loop with per-finding outcomes, per-file rollback, and SARIF/LSP interop. Now at v1.7.0. https://github.com/LarsArtmann/go-finding

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

## Launch checklist (when the flip happens)

1. Flip repo visibility to public; remove `GOPRIVATE` requirement (release-procedure.md "Private-Repo Consumer Setup").
2. First `go get github.com/larsartmann/go-finding@latest` to trigger pkg.go.dev rendering; verify module pages for all 4 modules (LAUNCH1).
3. Verify the GoReleaser release assets and Homebrew tap on the first public tag — requires creating `HOMEBREW_TAP_GITHUB_TOKEN` first (missing as of 2026-09-08) (LAUNCH4).
4. Submit to Awesome Go (entry above) (LAUNCH3).
5. Publish the announcement (drafts above) (LAUNCH2).
