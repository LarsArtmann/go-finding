package finding

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
)

// Sentinel errors for the detector registry.
var (
	errDetectorRegistered = errors.New("detector already registered")
	errUnknownDetector    = errors.New("unknown detector")
)

// DetectorRegistry manages named detector constructors.
// Create one with [NewDetectorRegistry] and register detectors
// with [DetectorRegistry.Register]. Use [DetectorRegistry.Build] to
// instantiate detectors by name.
//
// The registry is safe for concurrent use.
type DetectorRegistry struct {
	mu       sync.RWMutex
	builders map[string]func() Detector
}

// NewDetectorRegistry creates an empty registry.
func NewDetectorRegistry() *DetectorRegistry {
	return &DetectorRegistry{builders: make(map[string]func() Detector)} //nolint:exhaustruct
}

// Register adds a detector constructor under the given name.
// Returns an error if a detector with the same name is already registered.
func (r *DetectorRegistry) Register(name string, builder func() Detector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.builders[name]; exists {
		return fmt.Errorf("%w: %s", errDetectorRegistered, name)
	}

	r.builders[name] = builder

	return nil
}

// MustRegister panics if registration fails.
func (r *DetectorRegistry) MustRegister(name string, builder func() Detector) {
	err := r.Register(name, builder)
	if err != nil {
		panic(err)
	}
}

// Build instantiates a detector by name. Returns an error if not found.
//
//nolint:ireturn
func (r *DetectorRegistry) Build(name string) (Detector, error) {
	r.mu.RLock()
	builder, ok := r.builders[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: %s", errUnknownDetector, name)
	}

	return builder(), nil
}

// BuildAll instantiates all registered detectors in sorted name order.
func (r *DetectorRegistry) BuildAll() ([]Detector, error) {
	r.mu.RLock()
	names := slices.Sorted(maps.Keys(r.builders))
	r.mu.RUnlock()

	detectors := make([]Detector, 0, len(names))

	for _, name := range names {
		d, err := r.Build(name)
		if err != nil {
			return nil, err
		}

		detectors = append(detectors, d)
	}

	return detectors, nil
}

// Names returns registered detector names in sorted order.
func (r *DetectorRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Sorted(maps.Keys(r.builders))
}

// Has reports whether a detector with the given name is registered.
func (r *DetectorRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.builders[name]

	return ok
}
