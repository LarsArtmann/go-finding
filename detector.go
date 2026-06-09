package finding

import "context"

// Detector is the interface implemented by tools that can find issues.
type Detector interface {
	// Name returns the detector's name.
	Name() string
	// Detect runs the detector and returns findings.
	Detect(ctx context.Context) ([]Finding, error)
}

// DetectorFunc is an adapter to use ordinary functions as Detectors.
type DetectorFunc func(ctx context.Context) ([]Finding, error)

// Detect implements Detector.
func (f DetectorFunc) Detect(ctx context.Context) ([]Finding, error) {
	return f(ctx)
}

// Name implements Detector. Returns "anonymous" — use NamedDetectorFunc for a custom name.
//
//nolint:revive // receiver unused by design — method exists only to satisfy Detector interface
func (f DetectorFunc) Name() string {
	return "anonymous"
}

// NamedDetectorFunc returns a Detector with the given name wrapping the provided function.
//
//nolint:ireturn // intentional: factory function returns interface for polymorphism
func NamedDetectorFunc(name string, fn DetectorFunc) Detector {
	return &namedDetector{name: name, fn: fn}
}

type namedDetector struct {
	name string
	fn   DetectorFunc
}

func (n *namedDetector) Detect(ctx context.Context) ([]Finding, error) {
	return n.fn(ctx)
}

func (n *namedDetector) Name() string {
	return n.name
}
