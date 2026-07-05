package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/larsartmann/go-finding"
	"golang.org/x/sync/errgroup"
)

// Detect runs all detectors concurrently, collects their findings, and returns
// the combined result. Suppressed findings are excluded.
//
// This is the simplest entry point for running detectors without the full
// pipeline (detect → triage → fix → verify loop). For fix application, retry
// logic, metrics, or stage hooks, use [Pipeline] instead.
//
// Example:
//
//	findings, err := pipeline.Detect(ctx, myDetector1, myDetector2)
//	report := finding.NewReport(toolInfo)
//	report.AddFindings(findings)
func Detect(ctx context.Context, detectors ...Detector) ([]finding.Finding, error) {
	if len(detectors) == 0 {
		return nil, nil
	}

	var (
		mu          sync.Mutex
		allFindings []finding.Finding
	)

	g, ctx := errgroup.WithContext(ctx)

	for _, d := range detectors {
		g.Go(func() error {
			err := CheckCanceled(ctx)
			if err != nil {
				return err
			}

			findings, err := d.Detect(ctx)
			if err != nil {
				return fmt.Errorf("detector %s: %w", d.Name(), err)
			}

			var active []finding.Finding

			for _, f := range findings {
				if !f.IsSuppressed() {
					active = append(active, f)
				}
			}

			mu.Lock()

			allFindings = append(allFindings, active...)
			mu.Unlock()

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	return allFindings, nil
}

// ApplyToContent resolves and applies fixes to the given content using the
// default FixProvider chain (Offset → Line → Substring). It returns the modified
// content and the number of fixes successfully applied.
//
// This is the content-level API for applying fixes without filesystem operations.
// Use this when you already have file content in memory (e.g. from a git blob,
// editor buffer, or API response). For disk-based fix application with backup
// and rollback, use [NewFixApplier] or [Pipeline] instead.
//
// Example:
//
//	content, err := os.ReadFile(path)
//	if err != nil { return err }
//	result, applied := pipeline.ApplyToContent(content, findings)
//	if applied > 0 {
//	    _ = os.WriteFile(path, result, 0o644)
//	}
func ApplyToContent(content []byte, fixes []finding.Finding) ([]byte, int) {
	engine := NewFixEngine()
	result, _, applied := engine.Apply(content, fixes)

	return result, applied
}
