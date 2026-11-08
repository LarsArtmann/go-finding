# Detector Coverage Guide

How a detector reports **what it actually examined**, not just what it
found. The coverage channel answers the question findings alone cannot:
_did the tool see every file, or did it silently scan nothing?_

## Table of Contents

- [Why coverage evidence?](#why-coverage-evidence)
- [The API surface](#the-api-surface)
- [The v1-fallback contract](#the-v1-fallback-contract)
- [Writing a DetectorV2](#writing-a-detectorv2)
- [Host integration: AddDetectResult](#host-integration-adddetectresult)
- [Where it flows in the pipeline](#where-it-flows-in-the-pipeline)
- [JSON serialization](#json-serialization)

---

## Why coverage evidence?

A report that says "0 findings" is ambiguous. Either the tree is clean, or
the detector never looked at anything, or it skipped a vendored module and
nobody recorded it. Consumers rendering dashboards and CI gates cannot tell
the difference — every provider-built report looked like `filesScanned: 0`.

Since v1.14.0 the report summary carries the evidence fields
(`Summary.FilesScanned`, `Summary.SkippedModules`), but the v1 `Detector`
contract returns `[]Finding` only — a detector had no way to populate them.
`DetectResult` closes that gap: the run result carries findings AND the
measurement that makes "0 findings" checkable.

## The API surface

Four pieces, all in the core module:

| Identifier                          | Kind      | Purpose                                                        | Evidence              |
| ----------------------------------- | --------- | -------------------------------------------------------------- | --------------------- |
| `finding.DetectResult`              | struct    | One detector run: `Findings`, `FilesScanned`, `SkippedModules` | `detect_result.go:15` |
| `finding.DetectorV2`                | interface | Optional richer contract: `Detector` + `DetectV2(ctx)`         | `detect_result.go:39` |
| `finding.RunDetect(ctx, d)`         | func      | Runs the richest contract the detector implements              | `detect_result.go:53` |
| `(*finding.Report).AddDetectResult` | method    | Adds findings + records coverage into `Report.Summary`         | `detect_result.go:72` |

`DetectResult.FilesScanned` counts files actually examined, **including
clean ones**. Zero is legal evidence (an empty target tree) — that is why
the number rides on the run result instead of being inferred from the
findings. `SkippedModules` lists nested module roots the detector
deliberately did not descend into; reporting both sides makes the coverage
arithmetic checkable (absent modules are not scanned files).

## The v1-fallback contract

`RunDetect` is the single dispatch point:

```go
func RunDetect(ctx context.Context, d Detector) (DetectResult, error)
```

- DetectorV2 value → `DetectV2` runs, coverage is whatever the detector
  measured.
- v1-only detector → plain `Detect` runs; the result carries the findings
  with coverage **zeroed**. That is the documented contract, not an error:
  the host cannot know what a v1 detector scanned, and `filesScanned: 0`
  from a v1 detector means "unknown", never "verified empty".

The detector's error passes through unwrapped — callers (pipeline,
convenience wrappers) add their own context. `DetectorV2` embeds
`Detector`, so a v2 value is usable anywhere a `Detector` is, and `DetectV2`
should agree with `Detect` on the findings for the same run.

## Writing a DetectorV2

```go
type myLinter struct{ root string }

func (l *myLinter) Name() string { return "mylinter" }

func (l *myLinter) Detect(ctx context.Context) ([]finding.Finding, error) {
	res, err := l.DetectV2(ctx)
	if err != nil {
		return nil, err
	}
	return res.Findings, nil
}

func (l *myLinter) DetectV2(ctx context.Context) (finding.DetectResult, error) {
	files, modules, findings, err := scan(l.root, ctx)
	if err != nil {
		return finding.DetectResult{}, err
	}
	return finding.DetectResult{
		Findings:       findings,
		FilesScanned:   len(files),      // clean files count too
		SkippedModules: modules,         // vendored/nested trees you refused
	}, nil
}
```

Keep `Detect` as the thin wrapper (delegate to `DetectV2`) so both methods
return identical findings for the same run — hosts that only know v1 and
hosts that dispatch through `RunDetect` must not see different issues.

## Host integration: AddDetectResult

A report fed by several detectors accumulates coverage:

```go
report := finding.NewReport("myhost")
for _, d := range detectors {
	res, err := finding.RunDetect(ctx, d)
	if err != nil {
		return fmt.Errorf("detector %s: %w", d.Name(), err)
	}
	report.AddDetectResult(res) // findings + FilesScanned += ; SkippedModules unioned
}
_ = report.ComputeSummary() // does NOT touch the coverage fields
```

`AddDetectResult` is safe for concurrent use; `FilesScanned` sums across
calls and `SkippedModules` unions in first-seen order. `ComputeSummary`
recomputes counts from findings but never touches either coverage field, so
coverage survives summary recomputation.

## Where it flows in the pipeline

Every pipeline detect path dispatches through `RunDetect`, so a DetectorV2
detector plugged into the pipeline contributes coverage automatically:

- stage runner: `pipeline/pipeline_detect.go:98`
- convenience `pipeline.Detect`: `pipeline/convenience.go:43`
- retry wrapper (V2-aware passthrough): `pipeline/retry.go:139`
- detect-verify stage: `pipeline/verify.go:37`

The serialized report carries the evidence in `summary.filesScanned` (always
present — "scanned 0" is evidence, not an omission) and
`summary.skippedModules` (omitted when empty for module-agnostic tools, so
their JSON is unchanged).

## JSON serialization

```json
{
  "summary": {
    "total": 3,
    "filesScanned": 412,
    "skippedModules": ["vendor/example.com/tool"]
  }
}
```

See `docs/guides/configuration.md` for the report-level output options and
`finding.Summary` for the full field list.
