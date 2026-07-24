---
name: Bug report
about: Report a defect in go-finding
title: "[bug] "
labels: bug
---

## Summary

A concise description of what's wrong.

## Expected behavior

What you expected to happen.

## Actual behavior

What actually happened, including any error messages or unexpected output.

## Reproduction

Minimal, self-contained steps to reproduce. A failing test case is ideal:

```go
// paste code here
```

```bash
# commands to run, e.g.
export GOEXPERIMENT=jsonv2
go test -run ExampleFoo ./...
```

## Environment

- **go-finding version:** (`go list -m github.com/larsartmann/go-finding`)
- **Go version:** (`go version`)
- **OS / arch:**
- **GOEXPERIMENT:** (run `go env GOEXPERIMENT`)

## Additional context

Anything else relevant: SARIF/JSON snippets, related findings, logs, or links.
