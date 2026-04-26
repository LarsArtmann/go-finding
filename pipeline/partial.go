package pipeline

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/larsartmann/go-finding"
	"golang.org/x/sync/errgroup"
)

// PartialResult holds findings from detectors that succeeded and errors from those that failed.
type PartialResult struct {
	Findings []finding.Finding
	Errors   map[string]error
}

// HasErrors reports whether any detector failed.
func (r *PartialResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// DetectPartial runs all detectors, collecting results from successful ones
// and errors from failed ones. Unlike Detect, a single detector failure
// does not abort the entire detection.
func (p *Pipeline) DetectPartial(ctx context.Context) (*PartialResult, error) {
	if p.config.ParallelDetectors {
		return p.detectPartialParallel(ctx)
	}

	return p.detectPartialSequential(ctx)
}

// addFindingsToResult adds non-suppressed findings to the partial result.
func (p *Pipeline) addFindingsToResult(result *PartialResult, findings []finding.Finding) {
	for _, f := range findings {
		if !f.IsSuppressed() {
			result.Findings = append(result.Findings, f)
			p.notifyFinding(f)
		}
	}
}

// notifyFinding calls OnFinding callback if configured.
func (p *Pipeline) notifyFinding(f finding.Finding) {
	if p.config.OnFinding != nil {
		p.config.OnFinding(f)
	}
}

// isContextDone returns true if the context is done (cancelled/timed out).
func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func contextError(ctx context.Context, msg string) error {
	return fmt.Errorf("%s: %w", msg, ctx.Err())
}

func (p *Pipeline) detectPartialSequential(ctx context.Context) (*PartialResult, error) {
	//nolint:exhaustruct
	result := &PartialResult{
		Errors: make(map[string]error),
	}

	for _, d := range p.detectors {
		if isContextDone(ctx) {
			return result, contextError(ctx, "context cancelled")
		}

		start := time.Now()
		findings, err := d.Detect(ctx)
		elapsed := time.Since(start)

		if err != nil {
			result.Errors[d.Name()] = err
			p.recordDetectorMetrics(d.Name(), elapsed, nil)

			continue
		}

		p.recordDetectorMetrics(d.Name(), elapsed, findings)
		p.addFindingsToResult(result, findings)
	}

	return result, nil
}

func (p *Pipeline) detectPartialParallel(ctx context.Context) (*PartialResult, error) {
	//nolint:exhaustruct
	result := &PartialResult{
		Errors: make(map[string]error),
	}

	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)

	for _, d := range p.detectors {
		g.Go(func() error {
			start := time.Now()
			findings, err := d.Detect(gctx)
			elapsed := time.Since(start)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.Errors[d.Name()] = err
				p.recordDetectorMetrics(d.Name(), elapsed, nil)

				return nil // Don't propagate — collect partial results
			}

			p.recordDetectorMetrics(d.Name(), elapsed, findings)
			p.addFindingsToResult(result, findings)

			return nil
		})
	}

	_ = g.Wait()

	if isContextDone(ctx) {
		return result, contextError(ctx, "context cancelled")
	}

	return result, nil
}

// ErrPartialDetection indicates one or more detectors failed during partial detection.
var ErrPartialDetection = errors.New("pipeline: partial detection failures")

// FormatPartialErrors formats partial detection errors into a single error message.
func FormatPartialErrors(errors map[string]error) error {
	if len(errors) == 0 {
		return nil
	}

	names := make([]string, 0, len(errors))
	for name := range errors {
		names = append(names, name)
	}

	slices.Sort(names)

	msgs := make([]string, 0, len(errors))
	for _, name := range names {
		msgs = append(msgs, fmt.Sprintf("%s: %v", name, errors[name]))
	}

	return fmt.Errorf("%w: %v", ErrPartialDetection, msgs)
}
