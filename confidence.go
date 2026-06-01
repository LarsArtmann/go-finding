package finding

import (
	"cmp"
	"fmt"
)

// Confidence represents the certainty level of a finding on a 0.0–1.0 scale.
// Use named constants (ConfidenceLow, ConfidenceMedium, ConfidenceHigh) for
// common values, or Confidence(f) for custom levels.
// The zero value is valid and represents no confidence information.
//
// Be aware: Direct construction with Confidence values outside [0.0, 1.0] is
// possible (e.g., Finding{Confidence: 1.5}). The Validate() method catches
// this. For guaranteed-valid values, use the Builder API (WithConfidence)
// or NewFinding (both clamp automatically).
type Confidence float64

// Standard confidence levels.
const (
	ConfidenceNone   Confidence = 0.0
	ConfidenceLow    Confidence = 0.25
	ConfidenceMedium Confidence = 0.5
	ConfidenceHigh   Confidence = 0.75
	ConfidenceFull   Confidence = 1.0
)

// IsValid returns true if the confidence is within [0.0, 1.0].
func (c Confidence) IsValid() bool {
	return c >= 0 && c <= 1
}

// Clamp returns the confidence clamped to [0.0, 1.0].
func (c Confidence) Clamp() Confidence {
	if c < 0 {
		return 0
	}

	if c > 1 {
		return 1
	}

	return c
}

// Compare returns -1, 0, or +1 depending on whether c is less than, equal to,
// or greater than other.
func (c Confidence) Compare(other Confidence) int {
	return cmp.Compare(float64(c), float64(other))
}

// String returns the confidence as a human-readable string.
// Named levels return their label (e.g., "medium"), custom values return a decimal.
func (c Confidence) String() string {
	switch c {
	case ConfidenceNone:
		return "none"
	case ConfidenceLow:
		return "low"
	case ConfidenceMedium:
		return "medium"
	case ConfidenceHigh:
		return "high"
	case ConfidenceFull:
		return "full"
	default:
		return fmt.Sprintf("%.2f", float64(c))
	}
}
