package pipeline

import (
	"context"
	"fmt"
	"sync"

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
			if p.config.OnFinding != nil {
				p.config.OnFinding(f)
			}
		}
	}
}

func (p *Pipeline) detectPartialSequential(ctx context.Context) (*PartialResult, error) {
	result := &PartialResult{
		Errors: make(map[string]error),
	}

	for _, d := range p.detectors {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		findings, err := d.Detect(ctx)
		if err != nil {
			result.Errors[d.Name()] = err
			continue
		}

		p.addFindingsToResult(result, findings)
	}

	return result, nil
}

func (p *Pipeline) detectPartialParallel(ctx context.Context) (*PartialResult, error) {
	result := &PartialResult{
		Errors: make(map[string]error),
	}
	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)

	for _, d := range p.detectors {
		d := d
		g.Go(func() error {
			findings, err := d.Detect(ctx)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.Errors[d.Name()] = err
				return nil // Don't propagate — collect partial results
			}
			p.addFindingsToResult(result, findings)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return result, err
	}
	return result, nil
}

// FormatPartialErrors formats partial detection errors into a single error message.
func FormatPartialErrors(errors map[string]error) error {
	if len(errors) == 0 {
		return nil
	}
	msgs := make([]string, 0, len(errors))
	for name, err := range errors {
		msgs = append(msgs, fmt.Sprintf("%s: %v", name, err))
	}
	return fmt.Errorf("partial detection failures: %v", msgs)
}
