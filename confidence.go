package finding

import "fmt"

// Confidence represents the certainty level of a finding on a 0.0–1.0 scale.
// Use named constants (ConfidenceLow, ConfidenceMedium, ConfidenceHigh) for
// common values, or Confidence(f) for custom levels.
// The zero value is valid and represents no confidence information.
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

// String returns the confidence as a formatted string.
func (c Confidence) String() string {
	return fmt.Sprintf("%.2f", float64(c))
}
