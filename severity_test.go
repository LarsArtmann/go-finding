package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeverity_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sev  Severity
		want bool
	}{
		{SeverityInfo, true},
		{SeverityWarning, true},
		{SeverityError, true},
		{SeverityCritical, true},
		{Severity(""), false},
		{Severity("unknown"), false},
		{Severity("Info"), false},
		{Severity("WARNING"), false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.sev.IsValid(), "IsValid")
	}
}

func TestSeverity_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sev  Severity
		want string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
		{SeverityCritical, "critical"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, string(tt.sev), "Severity constant")
	}
}

func TestSeverity_GreaterThan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b  Severity
		want  bool
		label string
	}{
		{SeverityCritical, SeverityError, true, "critical > error"},
		{SeverityError, SeverityWarning, true, "error > warning"},
		{SeverityWarning, SeverityInfo, true, "warning > info"},
		{SeverityInfo, SeverityInfo, false, "info == info"},
		{SeverityInfo, SeverityWarning, false, "info < warning"},
		{SeverityWarning, SeverityError, false, "warning < error"},
		{Severity("unknown"), SeverityInfo, false, "invalid < valid returns false"},
		{SeverityInfo, Severity("unknown"), false, "valid > invalid returns false"},
		{Severity("unknown"), Severity("unknown"), false, "invalid > invalid returns false"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.a.GreaterThan(tt.b), tt.label)
	}
}

func TestSeverity_LessThan(t *testing.T) {
	t.Parallel()

	assert.True(t, SeverityInfo.LessThan(SeverityWarning), "info < warning")
	assert.False(t, SeverityError.LessThan(SeverityWarning), "error < warning")
	assert.False(t, SeverityWarning.LessThan(SeverityWarning), "warning < warning")
	assert.False(t, Severity("unknown").LessThan(SeverityInfo), "invalid < valid")
	assert.False(t, SeverityInfo.LessThan(Severity("unknown")), "valid < invalid")
}

func TestSeverity_Ordering(t *testing.T) {
	t.Parallel()

	ordered := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}

	for i := range ordered {
		for j := range ordered {
			assert.Equal(t, i > j, ordered[i].GreaterThan(ordered[j]))
		}
	}
}

func TestSeverity_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b  Severity
		want  int
		label string
	}{
		{SeverityInfo, SeverityInfo, 0, "info == info"},
		{SeverityWarning, SeverityInfo, 1, "warning > info"},
		{SeverityInfo, SeverityWarning, -1, "info < warning"},
		{SeverityCritical, SeverityError, 1, "critical > error"},
		{SeverityError, SeverityCritical, -1, "error < critical"},
		{Severity("unknown"), SeverityInfo, -1, "invalid < valid"},
		{SeverityInfo, Severity("unknown"), 1, "valid > invalid"},
		{Severity("foo"), Severity("bar"), 1, "different invalids ordered lexicographically"},
		{Severity("bar"), Severity("foo"), -1, "different invalids ordered lexicographically reverse"},
		{Severity("foo"), Severity("foo"), 0, "same invalid == same invalid"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.a.Compare(tt.b), tt.label)
	}
}
