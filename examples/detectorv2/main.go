// Command detectorv2 demonstrates the coverage-evidence channel: a
// DetectorV2 reports not only findings but also what it actually examined
// (files scanned, modules skipped), and a report accumulates that evidence
// across detectors via AddDetectResult. "0 findings" becomes checkable.
//
// Run it from the repository root:
//
//	go run ./examples/detectorv2
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	finding "github.com/larsartmann/go-finding"
)

// fakeTree is a deterministic stand-in for a filesystem walk.
var fakeTree = map[string]string{
	"clean.go":        "package ok\n",
	"dirty.go":        "package dirty\nfunc TODO() {}\n",
	"vendor/x/a.go":   "package x\nfunc TODO() {}\n",
	"vendor/x/nest.go": "package x\n",
}

// todoLinter is a DetectorV2: it reports findings plus coverage evidence.
type todoLinter struct {
	root string
}

func (l *todoLinter) Name() string { return "todo-linter" }

// Detect keeps the v1 contract as a thin wrapper over DetectV2 so both
// methods agree on the findings for the same run.
func (l *todoLinter) Detect(ctx context.Context) ([]finding.Finding, error) {
	res, err := l.DetectV2(ctx)
	if err != nil {
		return nil, err
	}
	return res.Findings, nil
}

func (l *todoLinter) DetectV2(ctx context.Context) (finding.DetectResult, error) {
	res := finding.DetectResult{}
	for path, src := range fakeTree {
		if strings.HasPrefix(path, "vendor/") {
			if path == "vendor/x/a.go" {
				res.SkippedModules = append(res.SkippedModules, "vendor/x")
			}
			continue // vendored module: deliberately not scanned
		}
		res.FilesScanned++ // clean files count as evidence too
		if ctx.Err() != nil {
			return finding.DetectResult{}, ctx.Err()
		}
		if strings.Contains(src, "TODO") {
			res.Findings = append(res.Findings, finding.NewFinding(
				"no-todo", finding.ToolName(l.Name()), "TODO marker found",
				finding.SeverityWarning, finding.Pos(finding.FilePath(path), 2, 6),
				finding.ConfidenceHigh,
			))
		}
	}
	return res, nil
}

func main() {
	report := finding.NewReport(finding.ToolInfo{Name: "todo-linter", Version: "1.0.0"})

	res, err := finding.RunDetect(context.Background(), &todoLinter{root: "."})
	if err != nil {
		log.Fatal(err)
	}
	report.AddDetectResult(res)
	report.ComputeSummary()

	fmt.Printf("findings:      %d\n", report.Summary.Total)
	fmt.Printf("files scanned: %d\n", report.Summary.FilesScanned)
	fmt.Printf("skipped:       %v\n", report.Summary.SkippedModules)
}
