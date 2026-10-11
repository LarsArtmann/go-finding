> **OUTCOME (2026-10-11): filed as [#43](https://github.com/LarsArtmann/go-finding/issues/43), fixed on master the same day (commit 10429a5: `StringFromContext`/`IntFromContext`/`BoolFromContext` + `ErrOptionKindMismatch`), issue closed with evidence. This draft is historical; the shipped contract is in `options.go` + `options_test.go`.**

<!-- proposed title:
Typed option accessors: every tool re-implements kind-assert + default boilerplate (23 lines for one option in go-auto-upgrade)
labels: enhancement
-->

## Why

Options are declared with a `Kind` and an informational `Default` (`toolsdk/options.go:44-57`), but the SDK stops at kind validation: `OptionsFromContext` returns raw `map[string]any` (`options.go:120-134`) and the tool does the rest by hand. Every tool with one option re-implements the same plumbing: lookup, declared-check, type-assert with a `%T` error, fallback to the default, then the tool's own value validation.

First real consumer, measured: go-auto-upgrade's `blameFromContext` is 23 lines to read ONE declared string option (`larsartmann/go-auto-upgrade` `pkg/provider/provider.go:234-256`, @ d6c1158). Multiply by options per tool and tools per fleet.

## Source-level cause

```go
// toolsdk/options.go:51-53 (v1.15.0, go-finding @ 91a216f)
// Default documents the tool's built-in fallback. Informational: the
// SDK does not inject it — when no value is set the tool applies its
// own default. Must match Kind when non-nil.
Default any
```

## Proposal

- [ ] Typed accessors alongside `OptionsFromContext`: `StringFromContext(ctx, name, fallback) (string, error)`, `IntFromContext(...)`, `BoolFromContext(...)` — or one generic `ValueFromContext[T]`. Lookup miss → fallback, no error. Kind mismatch → error matching `ErrOptionKindMismatch` (belt and braces: `ValidateOptions` should have caught it before the run).
- [ ] Semantic/enum validation stays tool-side — that contract (`spec.go:63-65`) is correct and unchanged. go-auto-upgrade keeps its `all|bisect` switch; it just stops hand-rolling the lookup/assert/fallback plumbing.
- [ ] `options_test.go`: fallback cases (unset, empty set, nil context) + kind-mismatch error cases per accessor.

## To verify

`blameFromContext` collapses to one accessor call plus the enum switch; existing `ValidateOptions` behavior unchanged; all existing specs compile.

💘 Generated with Crush
