// Command detectorv2 demonstrates that a DetectorV2 plugged into the
// pipeline contributes its coverage evidence automatically: every pipeline
// detect path dispatches through finding.RunDetect, and PipelineResult
// carries FilesScanned from the final detect stage. A v1-only detector in
// the same slot keeps working (coverage zeroed, RunDetect fallback).
//
// Run it from the repository root:
//
//	go run ./pipeline/examples/detectorv2
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

// todoLinter is a DetectorV2: findings plus coverage evidence. A real
// detector would count the files it actually examined; the fake tree keeps
// the demo deterministic.
type todoLinter struct {
	tree map[string]string
}

func (l *todoLinter) Name() string { return "todo-linter" }

func (l *todoLinter) Detect(ctx context.Context) ([]finding.Finding, error) {
	res, err := l.DetectV2(ctx)
	if err != nil {
		return nil, err
	}
	return res.Findings, nil
}

func (l *todoLinter) DetectV2(ctx context.Context) (finding.DetectResult, error) {
	res := finding.DetectResult{}
	for path, src := range l.tree {
		if strings.HasPrefix(path, "vendor/") {
			continue
		}
		res.FilesScanned++
		if ctx.Err() != nil {
			return finding.DetectResult{}, ctx.Err()
		}
		if strings.Contains(src, "TODO") {
			res.Findings = append(res.Findings, finding.NewFinding(
				"no-todo", finding.ToolName(l.Name()), "TODO marker found",
				finding.SeverityWarning, finding.Pos(finding.FilePath(path), 1, 1),
				finding.ConfidenceHigh,
			))
		}
	}
	return res, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	linter := &todoLinter{tree: map[string]string{
		"clean.go":      "package ok\n",
		"dirty.go":      "package dirty // TODO fix\n",
		"vendor/x/a.go": "package x // TODO fix\n",
	}}

	p, err := pipeline.New(pipeline.DefaultConfig(), ".", linter)
	if err != nil {
		return err
	}

	res, err := p.Run(context.Background())
	if err != nil {
		return err
	}

	fmt.Printf("total detected: %d\n", res.TotalDetected)
	fmt.Printf("files scanned:  %d\n", res.FilesScanned)
	return nil
}
