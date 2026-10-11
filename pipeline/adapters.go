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

// FindingTransformer transforms findings between detection and triage.
// Processors are chained in order, allowing filtering, enrichment, or transformation.
type FindingTransformer interface {
	// Name returns the processor's name for logging and debugging.
	Name() string
	// Transform applies a transformation to the findings and returns the result.
	// The context is used for cancellation. Return an error to abort the pipeline.
	Transform(ctx context.Context, findings []finding.Finding) ([]finding.Finding, error)
}

// TransformerFunc is an adapter to use ordinary functions as FindingTransformers.
type TransformerFunc func(findings []finding.Finding) []finding.Finding

// Transform implements FindingTransformer.
func (f TransformerFunc) Transform(
	_ context.Context,
	findings []finding.Finding,
) ([]finding.Finding, error) {
	return f(findings), nil
}

// Name implements FindingTransformer. Returns "".
func (TransformerFunc) Name() string {
	return ""
}

// NamedTransformerFunc returns a FindingTransformer with the given name wrapping the provided function.
func NamedTransformerFunc(name string, fn TransformerFunc) FindingTransformer {
	return &namedTransformer{name: name, fn: fn}
}

type namedTransformer struct {
	name string
	fn   TransformerFunc
}

func (n *namedTransformer) Transform(
	_ context.Context,
	findings []finding.Finding,
) ([]finding.Finding, error) {
	return n.fn(findings), nil
}

func (n *namedTransformer) Name() string {
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
