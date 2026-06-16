package pipeline

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
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

// notifyFinding calls OnFinding callback if configured.
// Safe for concurrent use when ParallelDetectors is enabled.
func (p *Pipeline) notifyFinding(f finding.Finding) {
	if p.config.OnFinding == nil {
		return
	}

	if p.config.ParallelDetectors {
		p.callbackMu.Lock()
		defer p.callbackMu.Unlock()
	}

	p.config.OnFinding(f)
}

func (p *Pipeline) detectPartialSequential(ctx context.Context) (*PartialResult, error) {
	//nolint:exhaustruct
	result := &PartialResult{
		Errors: make(map[string]error),
	}

	for _, d := range p.detectors {
		err := CheckCanceledWithMsg(ctx, "context cancelled")
		if err != nil {
			return result, err
		}

		findings, err := p.runOneDetector(ctx, d)
		if err != nil {
			if IsContextError(err) {
				return result, err
			}

			result.Errors[d.Name()] = err

			continue
		}

		result.Findings = append(result.Findings, findings...)
	}

	return result, nil
}

func (p *Pipeline) detectPartialParallel(ctx context.Context) (*PartialResult, error) {
	//nolint:exhaustruct
	result := &PartialResult{
		Errors: make(map[string]error),
	}

	var (
		mu        sync.Mutex
		ctxErr    error
		hasCtxErr bool
	)

	g, gctx := errgroup.WithContext(ctx)

	for _, d := range p.detectors {
		g.Go(func() error {
			findings, err := p.runOneDetector(gctx, d)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if IsContextError(err) {
					if !hasCtxErr {
						ctxErr = err
						hasCtxErr = true
					}
				} else {
					result.Errors[d.Name()] = err
				}

				return nil // Don't propagate to errgroup — collect partial results
			}

			result.Findings = append(result.Findings, findings...)

			return nil
		})
	}

	_ = g.Wait()

	if hasCtxErr {
		return result, ctxErr
	}

	err := CheckCanceledWithMsg(ctx, "context cancelled")
	if err != nil {
		return result, err
	}

	return result, nil
}

// ErrPartialDetection indicates one or more detectors failed during partial detection.
var ErrPartialDetection = errors.New("pipeline: partial detection failures")

// FormatPartialErrors formats partial detection errors into a single error message.
// The returned error wraps ErrPartialDetection and supports errors.Is/As for
// both the sentinel and each individual detector error.
func FormatPartialErrors(errs map[string]error) error {
	if len(errs) == 0 {
		return nil
	}

	names := slices.Sorted(maps.Keys(errs))

	inner := make([]error, 0, len(errs)+1)
	inner = append(inner, ErrPartialDetection)

	for _, name := range names {
		inner = append(inner, fmt.Errorf("%s: %w", name, errs[name]))
	}

	return errors.Join(inner...)
}
