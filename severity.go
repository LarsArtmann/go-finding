package finding

// Severity represents the severity level of a finding.
type Severity string

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

// GreaterThan returns true if this severity is greater than the other.
// Order: info < warning < error < critical.
func (s Severity) GreaterThan(other Severity) bool {
	return severityRank(s) > severityRank(other)
}

// LessThan returns true if this severity is less than the other.
func (s Severity) LessThan(other Severity) bool {
	return severityRank(s) < severityRank(other)
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
