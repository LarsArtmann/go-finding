# PRO/CONTRA: Adopting Azure/go-workflow for `pipeline/`

**Date:** 2026-06-18 | **Status:** Analysis complete — **recommendation: DO NOT adopt**

---

## Executive Summary

| Dimension            | go-finding `pipeline/`                                                                                                          | Azure/go-workflow                                                           |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| **What it is**       | Fixed **detect → triage → apply → verify loop** that repeats until the codebase is stable or capped                             | General-purpose **DAG orchestrator**: wire steps + edges, runs concurrently |
| **Topology**         | **Cyclic** — re-runs detection after every fix iteration (`for p.iterations < MaxIterations`)                                   | **Acyclic** — cycles rejected at preflight (`ErrCycleDependency`)           |
| **Core abstraction** | `Pipeline.Run(ctx)` + domain `PipelineResult`/`Iteration`/`CompletionReason`                                                    | `Steper.Do(ctx) error` + `Workflow.Do(ctx)` returning `ErrWorkflow` map     |
| **Domain logic**     | Owns `FixEngine`, `FixApplier`, byte-conflict, line-shift, triage, generated-filter, `goast/` (~2.5k LOC, the engineering risk) | None — pure orchestration, no domain knowledge                              |
| **Concurrency**      | `errgroup` (`ParallelDetectors`)                                                                                                | Native per-step goroutines + `MaxConcurrency` cap                           |
| **Retry / timeout**  | `RetryConfig` + `NewRetryDetector` wrapper; per-detector `DetectorTimeouts`                                                     | Native per-step `Retry(RetryOption)` + `Timeout(d)` + `cenkalti/backoff`    |
| **Partial failure**  | `GracefulDegradation` / `DetectPartial` collects per-detector errors                                                            | Native: a failing step skips only downstream; errors in `ErrWorkflow` map   |
| **Observability**    | `Metrics` + `StageHook` (before/after per stage)                                                                                | `StepInterceptor` / `AttemptInterceptor` + `contrib/otel`                   |
| **Maturity**         | v0.9.1, ~7.4k LOC of tests, stabilizing for v1.0                                                                                | **v0.1.13, pre-1.0**, ~108 stars, single primary maintainer                 |
| **Dependencies**     | `golang.org/x/sync` only (in `pipeline/`)                                                                                       | `cenkalti/backoff/v4` + `benbjohnson/clock` + transitive                    |

**The one-paragraph verdict:** go-workflow is a well-built DAG orchestrator, but go-finding's pipeline is **not a DAG** — it is a _fixed cyclic loop with stable-state termination_, plus ~2.5k LOC of domain logic (fix engine, conflict detection, the `goast/` AST provider — the actual engineering risk) that no orchestrator can touch. The parts go-workflow excels at (concurrency, retry, timeout, partial failure, observability) are **already solved** in go-finding (~1.8k LOC of orchestration) and tightly integrated with the domain result model. Adopting it would add a **pre-1.0 dependency**, violate the "minimal dependencies" principle, force a rewrite of ~7.4k LOC of passing tests, and **still leave the iteration loop hand-written** — because a fix→redetect loop is conceptually cyclic and DAGs are acyclic by definition.

---

## Capability Mapping

| go-finding feature                         | go-workflow equivalent                                              | Fit      |
| ------------------------------------------ | ------------------------------------------------------------------- | -------- |
| `Pipeline.Run` (single-use)                | `Workflow.Do` (single-runner guard)                                 | Trivial  |
| `for iterations < Max` cyclic loop         | **None** — DAGs reject cycles; must wrap a Go `for` around `Do`     | **None** |
| `ParallelDetectors` (errgroup)             | Native concurrency + `MaxConcurrency`                               | Good     |
| `DetectorTimeouts` (per-detector)          | `Step(s).Timeout(d)`                                                | Good     |
| `RetryConfig` / `NewRetryDetector`         | `Step(s).Retry(RetryOption)` + backoff                              | Good     |
| `GracefulDegradation` / `DetectPartial`    | Native partial failure + `ErrWorkflow` map                          | Good     |
| `StageHook` (before/after per _stage_)     | `BeforeStep`/`AfterStep` + interceptors (per _step_)                | Partial  |
| `Metrics` / `StageTiming`                  | Interceptors + `contrib/otel`                                       | Good     |
| `TriageFunc` / `FindingProcessor` chain    | Manual: steps with `Input`/`Output` callbacks                       | Partial  |
| `FixEngine` / byte-conflict / line-shift   | **Domain logic — out of scope**                                     | None     |
| `PipelineResult{Reason, Iterations[], …}`  | Manual: aggregate from `StateOf`/`ErrWorkflow`                      | **None** |
| `CompletionReason` taxonomy                | Partial: per-step `Succeeded/Failed/Canceled/Skipped`; no aggregate | Weak     |
| `DryRun` / `CorrelateFindings` / `Verify…` | Manual wiring                                                       | None     |

---

## Three Integration Models

### A. Full replacement — rewrite `pipeline/` on go-workflow

The orchestration core (`pipeline.go`, `pipeline_detect.go`, `pipeline_iteration.go`, `retry.go`, `partial.go`, `stage_hook.go`, `metrics.go`) is rewritten on top of `flow.Workflow`. Domain logic (`fix_engine*.go`, `conflict*.go`, `line_shift.go`, `fix_applier*.go`, `generated_filter.go`) stays.

### B. Partial adoption — go-workflow drives only concurrent detection + retry

Keep the outer iteration loop and result model; replace only the `detectParallel`/`runOneDetector`/`NewRetryDetector` cluster with a `flow.Workflow` that fans out detectors as steps.

### C. Optional engine — go-workflow as an opt-in `Executor` for advanced DAGs

Introduce a `pipeline.Executor` interface. The default is the current built-in loop; a new `pipeline/dag` sub-package wraps go-workflow for users who want custom multi-stage graphs (a feature go-finding does **not** currently offer).

---

## PRO (all models)

| #   | Argument                                                                                                                                                                  | Weight | Model |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ----- |
| 1   | **Best-in-class retry/backoff** — go-workflow's `RetryOption` (attempts, per-try timeout, custom backoff, notify) via `cenkalti/backoff` is richer than our `RetryConfig` | Medium | A, B  |
| 2   | **Native bounded concurrency** — `MaxConcurrency` caps parallel steps; we currently fan out unbounded via `errgroup`                                                      | Low    | A, B  |
| 3   | **First-class observability** — `StepInterceptor`/`AttemptInterceptor` + `contrib/otel` give tracing/metrics for free, replacing ad-hoc `Metrics` + `StageHook`           | Medium | A, B  |
| 4   | **Partial failure is the default** — a failing detector skips only its downstream, never siblings; maps cleanly onto `GracefulDegradation`                                | Medium | A, B  |
| 5   | **Tiny conformance interface** — `Detector.Detect(ctx)([]Finding,error)` adapts to `Steper.Do(ctx)error` in ~5 lines                                                      | Low    | A, B  |
| 6   | **Composable nesting** — `Workflow` is itself a `Step`; could enable future "sub-pipeline per language/file" without new abstractions                                     | Low    | A, C  |
| 7   | **Conditional branching** — `If`/`Switch`/`When` could express "apply only if findings > 0" declaratively                                                                 | Low    | A, C  |
| 8   | **Preflight cycle detection** — `ErrCycleDependency` catches graph errors before execution                                                                                | Low    | A, C  |
| 9   | **Reduces our orchestration test burden** — concurrency/retry/timeout edge cases become the library's responsibility                                                      | Low    | A, B  |
| 10  | **Industry-standard backoff lib** — `cenkalti/backoff` is the de-facto Go retry library, audited and widely deployed                                                      | Low    | A, B  |

---

## CONTRA (all models)

| #   | Argument                                                                                                                                                                                                                                                                                                                                                                                                                           | Weight       | Model   |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------- |
| 1   | **CRITICAL — Pre-1.0, unstable API.** Latest is `v0.1.13`. Breaking changes observed between minors (e.g. `v0.1.8` changed `AfterStep` semantics; `v0.1.13` removed the react UI as a breaking change). go-finding is **stabilizing for v1.0**; building its core loop on a v0.1 lib directly contradicts that goal                                                                                                                | **Critical** | A, B, C |
| 2   | **CRITICAL — Topology mismatch: cyclic loop vs acyclic DAG.** The pipeline's defining behavior is `for iterations < Max { detect→triage→apply }` — re-detect after each fix. go-workflow explicitly rejects cycles (`ErrCycleDependency`). One iteration can be a DAG, but the **outer stable-state loop must remain a hand-written Go `for`**, so the headline feature (DAG execution) does not apply to our central control flow | **Critical** | A       |
| 3   | **CRITICAL — Violates "minimal dependencies" principle.** AGENTS.md: _"Minimal dependencies — core types depend only on stdlib."_ `pipeline/` today depends only on `golang.org/x/sync`. go-workflow drags `cenkalti/backoff/v4` + `benbjohnson/clock` (+ transitive) into the root pipeline package, which every consumer (7+ tools) inherits                                                                                     | **Critical** | A, B    |
| 4   | **The hard parts are untouched.** ~2.5k LOC of `pipeline/` + `pipeline/goast/` is **domain logic** (`FixEngine`, byte-conflict, `line_shift`, `FixApplier`, `generated_filter`, triage, AST provider) — roughly half the package and _all_ of its engineering risk. No orchestrator helps here; adoption buys orchestration polish for the easy half while the actual complexity stays                                             | High         | A, B, C |
| 5   | **Loss of the domain result model.** `PipelineResult{CompletionReason, Iterations[], TotalDetected, Verification, Correlations, Metrics}` is a rich, domain-specific contract. go-workflow yields `ErrWorkflow map[Steper]StepResult` + per-step `State`. We would have to **rebuild the entire aggregation layer** — net negative                                                                                                 | High         | A       |
| 6   | **StageHook ↔ interceptor impedance mismatch.** `StageHook` fires on _stage boundaries inside an iteration loop_ (detect/process/triage/apply/verify × N iterations). Interceptors fire on _step boundaries_. The iteration axis has no equivalent; every existing `StageHook` consumer breaks                                                                                                                                     | High         | A       |
| 7   | **Massive test rewrite for marginal gain.** ~7.4k LOC of `pipeline/*_test.go` (BDD, fuzz, bench, integration, end-to-end) encode behavior in the current model. Migration is high-effort, high-regression-risk, and re-validates behavior that already passes                                                                                                                                                                      | High         | A       |
| 8   | **Narrow adoption + single maintainer.** ~108 stars, primary maintainer `@xuxife`. Lives under the `Azure` GitHub org but is **not an official Azure SDK product** — it is a community library. Not battle-tested across diverse domains                                                                                                                                                                                           | Medium       | A, B, C |
| 9   | **Iteration loop stays manual anyway.** Because of #2, `MaxIterations`, `CompletionReason` (`ReasonStable`/`ReasonMaxIterations`), and the `for` loop all remain. We keep the brain of the pipeline and offload only the limbs — and the limbs already work (`errgroup` + `RetryConfig` are ~150 LOC total)                                                                                                                        | High         | A, B    |
| 10  | **Partial adoption (#B) gives the least value for the most friction.** Wrapping a `flow.Workflow` inside our existing loop to replace ~150 LOC of clean `errgroup`+retry code introduces a dependency, a second concurrency model, and an adapter layer — for no user-visible feature                                                                                                                                              | Medium       | B       |
| 11  | **Deprecated `Reset()` blocks re-run ergonomics.** go-workflow's `Workflow.Reset()` (the natural way to re-run the same DAG per iteration) is **deprecated**, slated for removal. The alternative is reconstructing the workflow each iteration — friction that our single-use `Pipeline` does not have                                                                                                                            | Low          | A, B    |
| 12  | **`Skipped`/`Canceled` semantics differ from `CompletionReason`.** go-workflow settles Skipped/Canceled inline with no concurrency lease; our `ReasonCancelled`/`ReasonTimeout`/`ReasonStable` carry different domain meaning. Mapping is lossy and surprising                                                                                                                                                                     | Low          | A       |

---

## Analysis by Model

### Model A — Full replacement

**Verdict: NO**

The promise (declarative DAG, free retry/observability) is attractive, but the central control flow — a **cyclic fix→redetect loop** — is exactly what a DAG cannot express. One iteration could be a DAG, but then the outer `for`, `MaxIterations`, `CompletionReason`, and `PipelineResult` aggregation all stay hand-written. We would rewrite ~7.4k LOC of passing tests, lose the domain result model, break every `StageHook` consumer, and add a pre-1.0 dependency to the root pipeline package — to offload ~1.8k LOC of orchestration (of which only ~150 is the `errgroup`+retry code go-workflow would actually replace). The math does not work.

### Model B — Partial adoption (detection + retry only)

**Verdict: NO**

Replaces the smallest, cleanest, best-tested part of the pipeline (`detectParallel`, `runOneDetector`, `NewRetryDetector`) with a heavier abstraction, while introducing a _second_ concurrency/execution model alongside the existing one. Maximum friction, minimum value. If we ever want bounded concurrency, a 10-line semaphore around the existing `errgroup` is strictly better than a new dependency.

### Model C — Optional `Executor` for custom DAGs

**Verdict: NOT NOW (revisit if a real need appears)**

This is the only model that does not break existing behavior — it would live in a `pipeline/dag` sub-package behind an `Executor` interface, used only by advanced consumers who want user-defined multi-stage graphs. It is architecturally clean **but** go-finding currently has no such requirement (the loop is fixed by design — see AGENTS.md), and building the abstraction now is premature. Revisit **only if** a concrete use case (e.g. pluggable per-language sub-pipelines) emerges **and** go-workflow reaches v1.0.

---

## Recommendation

**Do not adopt Azure/go-workflow.** Keep the current pipeline. It is purpose-built, well-tested, and already solves the orchestration problems go-workflow targets — in a form tightly integrated with go-finding's domain result model and cyclic iteration semantics.

### What is actually worth doing instead

| Step | Action                                                                                                                                         | Effort   |
| ---- | ---------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| 1    | **Add bounded concurrency** to `detectParallel` via a `golang.org/x/sync/semaphore` (10-line change) — closes the only real gap vs go-workflow | 30 min   |
| 2    | **Borrow the interceptor pattern** — consider refactoring `StageHook` toward a chain-of-handlers shape if hook composition becomes painful     | optional |
| 3    | **Keep `RetryConfig`** — it is ~150 LOC, fully tested, and dependency-free. Do not trade it for a v0.1 library                                 | 0 min    |
| 4    | **Record this decision** in `docs/architecture-decisions.md` as an ADR so the question does not recur                                          | 10 min   |
| 5    | **Revisit if** both triggers fire: (a) a concrete need for user-defined DAGs appears, **and** (b) go-workflow ships a stable v1.0              | —        |

---

## What NOT to do

- **Do NOT** add `github.com/Azure/go-workflow` to `go.mod` — it is pre-1.0 and would enter the root pipeline package's dependency graph.
- **Do NOT** rewrite `pipeline.go` / `pipeline_iteration.go` as a `flow.Workflow` — the cyclic iteration loop cannot be expressed as an acyclic DAG.
- **Do NOT** replace `PipelineResult`/`CompletionReason`/`Iteration` with `ErrWorkflow`/`State` — the domain result contract is richer and consumed by the CLI, metrics, and tests.
- **Do NOT** adopt Model B (partial) — it maximizes friction for minimum gain and introduces a second concurrency model.
- **Do NOT** assume "Azure org = official Azure support" — go-workflow is a community library under the Azure org, not an Azure SDK product, with a single primary maintainer and ~108 stars.

---

### Triggers that would reopen this decision

1. go-workflow reaches a **stable v1.0** with a compatibility guarantee.
2. go-finding gains a **concrete requirement** for user-defined, multi-stage DAGs (not the current fixed loop).
3. A consumer explicitly asks to plug in a custom execution graph.

Until at least #1 and #2 are both true, the answer is no.

---

_Assisted-by: Crush <crush@charm.land>_
