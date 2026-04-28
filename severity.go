package finding

import "cmp"

// Severity represents the severity level of a finding.
type Severity string

// Severity levels for findings, ordered by urgency.
const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// IsValid returns true if the severity is a valid value.
func (s Severity) IsValid() bool {
	switch s {
	case SeverityInfo, SeverityWarning, SeverityError, SeverityCritical:
		return true
	}

	return false
}

// comparisonOp represents a severity comparison operation.
type comparisonOp int

const (
	cmpGreaterThan comparisonOp = iota
	cmpLessThan
	cmpGreaterThanOrEqual
	cmpLessThanOrEqual
)

// GreaterThan returns true if this severity is greater than the other.
// Order: info < warning < error < critical.
func (s Severity) GreaterThan(other Severity) bool {
	return s.compareOp(other, cmpGreaterThan)
}

// LessThan returns true if this severity is less than the other.
func (s Severity) LessThan(other Severity) bool {
	return s.compareOp(other, cmpLessThan)
}

// GreaterThanOrEqual returns true if this severity is greater than or equal to the other.
func (s Severity) GreaterThanOrEqual(other Severity) bool {
	return s.compareOp(other, cmpGreaterThanOrEqual)
}

// LessThanOrEqual returns true if this severity is less than or equal to the other.
func (s Severity) LessThanOrEqual(other Severity) bool {
	return s.compareOp(other, cmpLessThanOrEqual)
}

// compareOp is an internal helper that performs comparison based on the given operation.
func (s Severity) compareOp(other Severity, op comparisonOp) bool {
	if !s.isValidWith(other) {
		return false
	}
	c := s.Compare(other)
	switch op {
	case cmpGreaterThan:
		return c > 0
	case cmpLessThan:
		return c < 0
	case cmpGreaterThanOrEqual:
		return c >= 0
	case cmpLessThanOrEqual:
		return c <= 0
	default:
		return false
	}
}

// String returns the string representation of the severity.
func (s Severity) String() string {
	return string(s)
}

// Compare returns -1, 0, or 1 depending on whether s is less than, equal to,
// or greater than other. Invalid severities rank below all valid ones.
// Two different invalid severities are ordered lexicographically to ensure
// a total ordering.
func (s Severity) Compare(other Severity) int {
	rankS, rankOther := severityRank(s), severityRank(other)
	if rankS != rankOther {
		return cmp.Compare(rankS, rankOther)
	}

	if rankS < 0 {
		// Both invalid — use string comparison as tiebreaker.
		return cmp.Compare(string(s), string(other))
	}

	return 0
}

func severityRank(s Severity) int {
	switch s {
	case SeverityInfo:
		return 0
	case SeverityWarning:
		return 1
	case SeverityError:
		return 2
	case SeverityCritical:
		return 3
	}

	return -1
}

// isValidWith returns true if both severities are valid.
func (s Severity) isValidWith(other Severity) bool {
	return s.IsValid() && other.IsValid()
}
