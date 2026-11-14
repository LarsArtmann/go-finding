//go:build !race

package toolsdk

// RaceEnabled reports whether the race detector is active (`-race`).
// The false arm carries the same doc contract as race_on.go — see it
// for the fleet-wide rationale (the canonical copy is go-cqrs-lite's
// testutil/v4).
const RaceEnabled = false
