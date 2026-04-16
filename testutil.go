package finding

import "testing"

// assertFindingsLen asserts the length of a findings slice matches expected.
func assertFindingsLen(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", name, got, want)
	}
}

// assertFindingsLenEq asserts the length of a findings slice matches expected (short form).
func assertFindingsLenEq(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("findings count = %d, want %d", got, want)
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

// newTestReport creates a report with a single finding.
func newTestReport(toolName, id string, severity Severity, file string) *Report {
	r := NewReport(ToolInfo{Name: toolName})
	r.AddFinding(Finding{ID: id, Severity: severity, Position: Position{File: file}})
	return r
}

// newTestReports creates multiple test reports.
func newTestReports(specs []struct {
	toolName string
	id      string
	severity Severity
	file    string
}) []*Report {
	reports := make([]*Report, len(specs))
	for i, spec := range specs {
		reports[i] = newTestReport(spec.toolName, spec.id, spec.severity, spec.file)
	}
	return reports
}
