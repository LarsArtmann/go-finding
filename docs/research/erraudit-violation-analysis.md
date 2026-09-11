# erraudit Violation Analysis (2026-08-09)

**Trigger:** `GOEXPERIMENT=jsonv2 erraudit ./... --type-aware --enforce-go-error-family --no-suppress --enforce-samber-oops`

**Result:** 58 ERROR-severity violations reported.

**Verdict:** **No action required.** All 58 violations originate from the `--enforce-samber-oops` flag, which is inapplicable to this project. The 3 baseline violations (without enforcement flags) are false positives.

---

## 1. Root Cause

The `--enforce-samber-oops` flag description states:

> Flag stdlib error constructors as violations + detect oops anti-patterns (redundant nil guards) — **use when the project enforces samber/oops**

go-finding does **not** enforce samber/oops. The project conducted a detailed evaluation
([samber-oops-v1220-assessment.html](./samber-oops-v1220-assessment.html), 2026-06-17) and
explicitly rejected adoption:

| Metric                    | Score                               |
| ------------------------- | ----------------------------------- |
| oops quality as a library | 8.5 / 10                            |
| oops fit for go-finding   | **3 / 10**                          |
| Recommendation            | **Do not adopt oops in go-finding** |
| Recommended action        | **Do nothing**                      |

The assessment concluded: "Adopting it would violate the project's minimal-dependency
philosophy and bloat the binary." oops is a service observability toolkit (user/tenant
context, HTTP capture, distributed traces, panic recovery); go-finding is a static-analysis
**library**.

---

## 2. Violation Breakdown

### With `--enforce-samber-oops` (the command that was run): 58 violations

| Category                         | Count | Examples                                                                                     |
| -------------------------------- | ----- | -------------------------------------------------------------------------------------------- |
| `fmt.Errorf` with `%w`           | 40    | `json.go:50`, `sarif_export.go:80`, `format.go:23`, etc.                                     |
| `errors.New` (sentinels)         | 13    | `errors.go:12-16`, `category.go:75`, `severity.go:166`, `json.go:27-28`, `registry.go:16-17` |
| `errors.Join`                    | 4     | `finding_validate.go:21`, `report.go:44`, `sarif_types.go:152`                               |
| `fmt.Errorf` (no `%w`)           | 1     | `confidence.go:75` (`ErrInvalidConfidence` sentinel)                                         |
| Ignored errors (false positives) | 3     | See section 3 below                                                                          |

### Without enforcement flags (baseline): 3 violations

All three are false positives (see Section 3).

---

## 3. The 3 Baseline Violations (False Positives)

### 3.1 `id.go:202-203` — `hash.Hash.Write` never returns an error

```go
func writeLenField(h hash.Hash, field string) {
    binary.BigEndian.PutUint32(buf[:], uint32(len(field)))
    _, _ = h.Write(buf[:])
    _, _ = h.Write([]byte(field))
}
```

**Why this is correct:** `hash.Hash.Write` is documented in the Go standard library as:

> Write (via the embedded io.Writer interface) always returns (len(p), nil)

The `crypto/sha256` hasher used by `GenerateID` can never produce an error from `Write`.
The `_, _ =` pattern is the idiomatic way to satisfy `io.Writer` return signatures when
the error is structurally impossible. This is not a swallowed error.

### 3.2 `errors.go:162` — Type assertion, not an ignored error

```go
func IsFindingError(err error) bool {
    _, ok := errors.AsType[*FindingError](err)
    return ok
}
```

**Why this is correct:** This is a type-assertion pattern using `errors.AsType` (from
go-error-family). The first return value is intentionally discarded — only the boolean
matters. This is analogous to the stdlib `errors.As(err, &target)` pattern where you
discard the target and check success. It is not an ignored function-call error.

---

## 4. The Project's Error Model Is Intentionally Layered

go-finding uses a deliberate, three-layer error design. Replacing all stdlib constructors
with a single library would collapse these layers and misclassify errors.

| Layer                 | Pattern                                                                       | Purpose                                                         | Replacing with oops?                                       |
| --------------------- | ----------------------------------------------------------------------------- | --------------------------------------------------------------- | ---------------------------------------------------------- |
| **Domain errors**     | `FindingError` struct with `Category`, `Finding`, `File`, `Position`, `Cause` | Public API; implements go-error-family `Coded` + `Classified`   | No — oops has no domain model                              |
| **Sentinel errors**   | `var ErrValidation = errors.New(...)`                                         | Package-level values for `errors.Is` matching                   | No — oops sentinels don't support `Is` the same way        |
| **Internal wrapping** | `fmt.Errorf("marshal report: %w", err)`                                       | Add implementation context to errors from stdlib/external calls | No — oops adds observability overhead for internal details |
| **Error aggregation** | `errors.Join(errs...)`                                                        | Collect multiple validation errors into one                     | No — `errors.Join` is stdlib since Go 1.20                 |

Using `errorfamily.NewRejection()` (as the `--enforce-go-error-family` flag suggests) for all
of these would **misclassify** internal errors. For example, a "marshal report" error is
not a Rejection (caller input error) — it is a Transient/IO error. The existing `FindingError`
type with its `ErrorFamily()` method already handles classification correctly at the right
boundary.

---

## 5. Technical Incompatibilities with oops

Even if adoption were desired, five call sites use multi-`%w` wrapping, which
`oops.Wrapf` **cannot express** (it accepts exactly one error to wrap):

| File                               | Line    | Code                                                                                  |
| ---------------------------------- | ------- | ------------------------------------------------------------------------------------- |
| `json.go`                          | 149     | `fmt.Errorf("%w: %w", ErrInvalidFinding, err)`                                        |
| `pipeline/config_file.go`          | 144     | `fmt.Errorf("%w: %q: %w", errResolveDetector, name, err)`                             |
| `pipeline/fix_provider_helpers.go` | 57      | `fmt.Errorf("%w: %w", ErrPositionUnresolvable, err)`                                  |
| `pipeline/fix_applier.go`          | 129-133 | `fmt.Errorf("%w (rollback also failed: %w)", backupErr, rollbackErr)`                 |
| `pipeline/fix_applier.go`          | 160-164 | `fmt.Errorf("%w (rollback also failed: %w)", applyErr, errors.Join(rollbackErrs...))` |

Additionally, the project uses `errors.Join` in 10 locations across 9 files for error
aggregation — a stdlib pattern that oops does not replace.

---

## 6. Cross-Validation

| Source                                   | Finding                                                                                                                                  |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **how-to-golang skill** (project policy) | Mandates `cockroachdb/errors + uniflow` for error wrapping. samber/oops is not listed anywhere — not required, not banned, not mentioned |
| **go.mod** (all 4 modules)               | samber/oops is completely absent — zero imports, zero dependencies                                                                       |
| **CI** (`.github/workflows/ci.yml`)      | erraudit is **not** wired into CI; the command was run ad-hoc                                                                            |
| **Prior assessment**                     | Explicitly rejected at 3/10 fit                                                                                                          |

---

## 7. Command Reference

### Correct way to run erraudit on this project

```bash
# Default audit (with false-positive heuristics) — catches real issues
GOEXPERIMENT=jsonv2 erraudit ./... --type-aware

# Full audit (no suppression) — shows everything for manual review
GOEXPERIMENT=jsonv2 erraudit ./... --type-aware --no-suppress

# With go-error-family enforcement (for reviewing domain error boundaries)
# NOTE: Will flag all stdlib constructors. Review manually.
GOEXPERIMENT=jsonv2 erraudit ./... --type-aware --enforce-go-error-family --no-suppress
```

### Incorrect for this project

```bash
# This flag is for projects that enforce samber/oops — go-finding does NOT
GOEXPERIMENT=jsonv2 erraudit ./... --enforce-samber-oops
```

---

## 8. Conclusion

**No changes to the codebase.** The 58 violations are an artifact of running an
enforcement flag that contradicts the project's documented architecture. The 3 baseline
violations are tool false positives on idiomatic Go patterns (`hash.Hash.Write` and
type assertions).

The project's error model — `FindingError` + go-error-family interfaces + stdlib
constructors for internal wrapping — was deliberately designed and is fit for purpose.

---

## 9. Addendum (2026-09-11): policy made executable

This document's "no changes" verdict did not age well: a month later the same
`--enforce-samber-oops` run was interpreted as a defect list again. The policy is now
executable instead of prose.

### What was done

| Action | Detail |
| ------ | ------ |
| **3 false positives suppressed** | `id.go` hash writes ×2 and `errors.go` `AsType` ok-pattern now carry `//nolint:erraudit` with a reason (verified: still false positives on current code) |
| **3 CRITICAL context_loss FIXED** | `pipeline/file_backup.go` (backup dir now carried via `ioErrorAt`), `pipeline/pipeline.go` ×2 (`rootDir=%q` added to validate-config and fix-applier errors) |
| **2 legacy `errors.As` migrated** | `cmd/go-finding/internal/detectors/{govet,staticcheck}.go` → `errors.AsType[*exec.ExitError]` (the genuine Go 1.26 modernization) |
| **Intentional patterns documented** | 24 `//nolint:erraudit // <reason>` directives on verified-intentional sites (deferred cleanup, error-path secondary closes, best-effort CLI output, partial-results `g.Wait()`, invalid-edit skip, external-tool JSON tolerance, `AsType` ok-patterns) |
| **Gate script added** | `scripts/error-audit.sh` — runs `erraudit ./... --type-aware` in all 5 modules, prints per-module violation counts and an explicit OK/FAIL verdict; FAIL path verified by injecting a violation (exit 1) and re-verifying clean (exit 0) |
| **generic_return warnings: intentionally not applied** | All 21 come from `--enforce-generic-return`, which erraudit itself ships off by default ("returning error is standard Go practice"). Bespoke error types for every formatter/helper would be non-idiomatic; the domain boundary already returns `*FindingError`. |

### Gate status and CI blocker

- Local gate: `./scripts/error-audit.sh` → `OK` across all 5 modules.
- **CI wiring is blocked**: the erraudit repository is private, so `go install` cannot
  fetch it from a public-repo runner without credentials. Reintroducing
  `GOPRIVATE`/token plumbing was deliberately avoided (it was removed when this repo
  went public). Revisit when erraudit is public or a fine-grained token is acceptable.
- Suppression hygiene: `erraudit nolint-audit .` from the repo root reports stale
  directives (verified 2026-09-11: 27 needed, 0 stale). NOTE: the `./...` form
  silently scans nothing — always pass `.`.

### Canonical invocation (repeat of section 7)

```bash
GOEXPERIMENT=jsonv2 erraudit ./... --type-aware   # per module; or ./scripts/error-audit.sh
```

---

_Assisted-by: Crush_
