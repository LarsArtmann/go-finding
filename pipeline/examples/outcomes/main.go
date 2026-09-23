// Command outcomes demonstrates the per-finding fix outcome API and the
// configurable rollback policy:
//
//  1. FixEngine.ApplyWithOutcomes reports what happened to every finding
//     (applied, refused, failed, ...) instead of collapsing everything into
//     an applied count.
//  2. FixApplier.ApplyWithReport applies fixes on disk and reports rolled-back
//     files; RollbackPolicy controls whether a hard failure restores only the
//     failing file or every file modified earlier in the run.
//
// Run it from the repository root:
//
//	go run ./pipeline/examples/outcomes
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

const demoContent = "package demo\n\nfunc old() {}\n"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := demoEngineOutcomes(); err != nil {
		return err
	}

	return demoApplierRollback()
}

// demoEngineOutcomes applies fixes to an in-memory file and prints one
// outcome per finding.
func demoEngineOutcomes() error {
	tmpl := finding.NewTemplate("outcomes-demo").WithFixStrategy(finding.FixStrategyDirect)

	applied := tmpl.Builder("rename-func", "rename old to new", finding.SeverityWarning, finding.Pos("demo.go", 3, 1)).
		WithBeforeCode("func old()").
		WithAfterCode("func new()").
		MustBuild()

	refused := tmpl.Builder("rename-func", "BeforeCode does not match the file", finding.SeverityWarning, finding.Pos("demo.go", 3, 1)).
		WithBeforeCode("func ancient()").
		WithAfterCode("func new()").
		MustBuild()

	failed := tmpl.Builder("rename-func", "position past end of file cannot be resolved", finding.SeverityWarning, finding.Pos("demo.go", 99, 1)).
		WithBeforeCode("func ancient()").
		WithAfterCode("func new()").
		MustBuild()

	result := pipeline.NewFixEngine().ApplyWithOutcomes(
		[]byte(demoContent),
		[]finding.Finding{applied, refused, failed},
	)

	fmt.Printf("content after fixes:\n%s\n", result.Content)

	for _, o := range result.Outcomes {
		switch o.Status {
		case pipeline.FixOutcomeApplied:
			fmt.Printf("  applied: %s\n", o.Finding.Rule)
		case pipeline.FixOutcomeRefused:
			fmt.Printf("  refused (providers matched but produced no edits): %s\n", o.Finding.Message)
		case pipeline.FixOutcomeFailed:
			fmt.Printf("  failed: %s (%v)\n", o.Finding.Message, o.Err)
		default:
			fmt.Printf("  %s: %s\n", o.Status, o.Finding.Rule)
		}
	}

	return nil
}

// demoApplierRollback applies fixes on disk with an explicit rollback policy
// and prints the resulting report.
func demoApplierRollback() error {
	dir, err := os.MkdirTemp("", "outcomes-demo-*")
	if err != nil {
		return fmt.Errorf("create demo dir: %w", err)
	}

	defer func() { _ = os.RemoveAll(dir) }() //nolint:erraudit // best-effort demo cleanup

	file := filepath.Join(dir, "demo.go")
	if err := os.WriteFile(file, []byte(demoContent), 0o600); err != nil {
		return fmt.Errorf("write demo file: %w", err)
	}

	applier, err := pipeline.NewFixApplier(dir)
	if err != nil {
		return err
	}

	defer func() { _ = applier.Close() }() //nolint:erraudit // best-effort demo cleanup

	// RollbackPolicyFailingFile (default): on a hard file failure only the
	// failing file is restored; use RollbackPolicyAllFiles for
	// all-or-nothing semantics.
	applier.SetRollbackPolicy(pipeline.RollbackPolicyFailingFile)

	tmpl := finding.NewTemplate("outcomes-demo").WithFixStrategy(finding.FixStrategyDirect)
	fix := tmpl.Builder("rename-func", "rename old to new", finding.SeverityWarning, finding.Pos("demo.go", 3, 1)).
		WithBeforeCode("func old()").
		WithAfterCode("func new()").
		MustBuild()

	report, err := applier.ApplyWithReport(context.Background(), []finding.Finding{fix})
	if err != nil {
		return err
	}

	fmt.Printf("disk run: applied=%d outcomes=%d rolledBack=%v\n",
		report.Applied,
		len(report.Outcomes),
		report.RolledBack,
	)

	return nil
}
