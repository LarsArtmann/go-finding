# AGENTS.md - go-finding

## Project Overview

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools.

### Core Purpose

Seven tools detect issues. Zero tools route them to remediation. This library solves that by providing:

1. **Unified Finding type** - Common representation for all tools
2. **Pipeline** - Automated detect → triage → fix → verify loop
3. **SARIF output** - Standard interchange format
4. **LSP integration** - IDE support

### Key Files

| File              | Purpose                                     |
| ----------------- | ------------------------------------------- |
| `finding.go`      | Core Finding type                           |
| `severity.go`     | Severity enum (info/warning/error/critical) |
| `fix_strategy.go` | FixStrategy enum (none/suggest/direct/ai)   |
| `report.go`       | Report container with summary               |
| `filter.go`       | Filtering and grouping utilities            |
| `merge.go`        | Report merging with deduplication           |
| `sarif.go`        | SARIF 2.1.0 output                          |
| `lsp.go`          | LSP Diagnostic conversion                   |
| `diagnostic.go`   | go/analysis integration                     |

### Testing

```bash
just test        # Run tests
just bench       # Run benchmarks
just lint        # Run linter
```

### Dependencies

- `golang.org/x/tools` - go/analysis framework
- Standard library only (minimal deps)

### Design Principles

1. **Zero dependencies** for core types
2. **Immutable** - Findings are data, not state machines
3. **Lossless** - Conversions preserve all data
4. **Compatible** - Works with existing Go analysis tools

### Future Work

See EXECUTION_PLAN.md for roadmap.

---

_Assisted-by: Crush <crush@charm.land>_
