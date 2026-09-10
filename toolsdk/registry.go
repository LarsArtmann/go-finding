package toolsdk

import (
	"context"
	"strings"
	"sync"
)

// registry holds specs registered by tools. The default registry is process-
// global so that tools registering in init() are discoverable by BuildFlow
// via All().
type registry struct {
	mu    sync.RWMutex
	specs []Spec
}

// defaultRegistry is the process-global registry so that tools registering in
// init() are discoverable by BuildFlow via All().
var defaultRegistry = &registry{} //nolint:gochecknoglobals // process-global registry by design: tools self-register at import time

// Register validates s and adds it to the default registry. Intended for use
// in a tool's package-level var declaration:
//
//	var Provider = toolsdk.Register(toolsdk.Spec{...})
//
// Panics on invalid specs (empty Name, empty Description, or no Detect/Repair
// capability) because a malformed registration is a programming error that
// should surface at startup, not at runtime.
func Register(s Spec) Spec {
	mustValidate(s)

	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()

	defaultRegistry.specs = append(defaultRegistry.specs, s)

	return s
}

// snapshot returns a defensive copy of the specs slice under a read lock.
// Shared by All() and SnapshotForTest() so copy semantics stay consistent.
func (r *registry) snapshot() []Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Spec, len(r.specs))
	copy(out, r.specs)

	return out
}

// All returns every spec registered via Register, in registration order.
// BuildFlow calls this once at startup to discover all tool providers.
func All() []Spec {
	return defaultRegistry.snapshot()
}

// ResetForTest clears the registry. Test-only; production code must not call
// this (it would drop registered providers mid-run).
func ResetForTest() {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()

	defaultRegistry.specs = nil
}

// SnapshotForTest returns a copy of the current registry contents. Test-only;
// pair with RestoreForTest via t.Cleanup to isolate tests that mutate the
// process-global registry (e.g. via ResetForTest + Register) from each other
// and from the production init()-time registrations, so the package's test
// suite becomes shuffle-safe and order-independent.
func SnapshotForTest() []Spec {
	return defaultRegistry.snapshot()
}

// RestoreForTest replaces the registry with a copy of the given snapshot.
// Test-only. Used to undo mutations made during a test so subsequent tests
// observe the pre-test (typically production init()-registered) state.
func RestoreForTest(snapshot []Spec) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()

	defaultRegistry.specs = make([]Spec, len(snapshot))
	copy(defaultRegistry.specs, snapshot)
}

func mustValidate(s Spec) {
	if s.Name == "" {
		panic("toolsdk.Register: Spec.Name is required")
	}

	if strings.TrimSpace(s.Description) == "" {
		panic("toolsdk.Register: Spec " + s.Name + " has no Description")
	}

	if s.Detect == nil && s.Repair == nil {
		panic("toolsdk.Register: Spec " + s.Name + " has no Detect or Repair capability")
	}
}

// EnsureContext returns a non-nil context. Convenience for Detect/Repair
// implementations that receive a nil context in edge cases (should not happen
// in BuildFlow, but defensive for standalone tool use).
func EnsureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}
