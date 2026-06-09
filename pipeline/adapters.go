package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/larsartmann/go-finding"
)

// Detector is re-exported from the root finding package for backward compatibility.
// New code should use finding.Detector directly.
type Detector = finding.Detector

// DetectorFunc is re-exported from the root finding package for backward compatibility.
// New code should use finding.DetectorFunc directly.
type DetectorFunc = finding.DetectorFunc

// NamedDetectorFunc is re-exported from the root finding package for backward compatibility.
// New code should use finding.NamedDetectorFunc directly.
//
//nolint:gochecknoglobals // intentional: re-exported function variable for backward compatibility
var NamedDetectorFunc = finding.NamedDetectorFunc

// FindingProcessor transforms findings between detection and triage.
// Processors are chained in order, allowing filtering, enrichment, or transformation.
type FindingProcessor interface {
	// Name returns the processor's name for logging and debugging.
	Name() string
	// Process applies a transformation to the findings and returns the result.
	// The context is used for cancellation. Return an error to abort the pipeline.
	Process(ctx context.Context, findings []finding.Finding) ([]finding.Finding, error)
}

// ProcessorFunc is an adapter to use ordinary functions as FindingProcessors.
type ProcessorFunc func(findings []finding.Finding) []finding.Finding

// Process implements FindingProcessor.
func (f ProcessorFunc) Process(
	_ context.Context,
	findings []finding.Finding,
) ([]finding.Finding, error) {
	return f(findings), nil
}

// Name implements FindingProcessor. Returns "anonymous".
func (ProcessorFunc) Name() string {
	return "anonymous"
}

// NamedProcessorFunc returns a FindingProcessor with the given name wrapping the provided function.
func NamedProcessorFunc(name string, fn ProcessorFunc) FindingProcessor {
	return &namedProcessor{name: name, fn: fn}
}

type namedProcessor struct {
	name string
	fn   ProcessorFunc
}

func (n *namedProcessor) Process(
	_ context.Context,
	findings []finding.Finding,
) ([]finding.Finding, error) {
	return n.fn(findings), nil
}

func (n *namedProcessor) Name() string {
	return n.name
}

// IsContextError reports whether the error is caused by context cancellation
// or deadline exceeded. This is the single canonical check for context errors
// across the pipeline — use it instead of inline errors.Is comparisons.
func IsContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// CheckCanceled checks if the context is done and returns an appropriate error.
// Use this helper instead of inline context cancellation checks to avoid duplication.
func CheckCanceled(ctx context.Context) error {
	return CheckCanceledWithMsg(ctx, "operation cancelled")
}

// CheckCanceledWithMsg checks if the context is done and returns an error with the given message.
func CheckCanceledWithMsg(ctx context.Context, msg string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", msg, ctx.Err())
	default:
		return nil
	}
}

// WaitWithContext waits for the done channel while checking for context cancellation.
// Returns an error if context is cancelled before the done channel completes.
func WaitWithContext(ctx context.Context, done <-chan time.Time) (bool, error) {
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-done:
		return true, nil
	}
}
