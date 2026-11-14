//go:build race

package toolsdk

// RaceEnabled reports whether the race detector is active (`-race`).
// Timing, heap, and latency thresholds in tests relax under -race because
// instrumentation inflates allocations and CPU 5-10x.
//
// This is the fleet-wide copy of the two-file build-tag idiom (the
// canonical, documented version lives in go-cqrs-lite's testutil/v4).
// BuildFlow providers (the consumers of this package) use it to keep
// their e2e timing assertions green under the CI race arm instead of
// carrying a private race_on/race_off pair per repo.
const RaceEnabled = true
