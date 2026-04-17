package finding

import (
	"testing"
	"testing/quick"
)

// assertFindingsLen asserts the length of a findings slice matches expected.
func assertFindingsLen(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", name, got, want)
	}
}

// assertReportFieldStr asserts a string field on a report.
func assertReportFieldStr(t *testing.T, got, want, fieldName string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", fieldName, got, want)
	}
}

// assertReportFieldInt asserts an int field on a report summary.
func assertReportFieldInt(t *testing.T, got, want int, fieldName string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", fieldName, got, want)
	}
}

// sevFromInt maps an integer to a Severity (0=info, 1=warning, 2=error, 3=critical).
func sevFromInt(i int) Severity {
	sevs := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}
	if i < 0 || i >= len(sevs) {
		return SeverityInfo
	}
	return sevs[i]
}

// MakeSimpleFinding creates a Finding with minimal required fields.
func MakeSimpleFinding(id string, severity Severity) Finding {
	return Finding{
		ID:       id,
		Severity: severity,
	}
}

// MakeSimpleReport creates a Report with the given tool name.
func MakeSimpleReport(toolName string) *Report {
	return NewReport(ToolInfo{Name: toolName})
}

// assertFindingErrorFile asserts the File field of a FindingError.
func assertFindingErrorFile(t *testing.T, err *FindingError, want string) {
	if err.File != want {
		t.Errorf("File = %q, want %q", err.File, want)
	}
}

// checkProperty runs a property-based test using quick.Check.
func checkProperty[T any](t *testing.T, property func(T) bool) {
	if err := quick.Check(property, nil); err != nil {
		t.Error(err)
	}
}
