package finding

import "testing"

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
