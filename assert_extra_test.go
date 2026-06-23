package finding

import (
	"math/rand"
	"testing"
	"testing/quick"
)

func assertFindingsLen(t *testing.T, name string, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("%s = %d, want %d", name, got, want)
	}
}

// assertReportField asserts a field equals expected value using formatted output.
func assertReportField[T comparable](t *testing.T, fieldName string, got, want T) {
	t.Helper()

	if got != want {
		t.Errorf("%s = %v, want %v", fieldName, got, want)
	}
}

// assertSummaryField asserts a Report summary field equals expected value.
func assertSummaryField(t *testing.T, fieldName string, got, want int) {
	t.Helper()
	assertReportField(t, "Summary."+fieldName, got, want)
}

// assertSummarySeverity asserts the count of findings with a given severity in the report summary.
func assertSummarySeverity(t *testing.T, r *Report, sev Severity, want int) {
	t.Helper()

	if r.Summary.BySeverity[sev] != want {
		t.Errorf("Summary.BySeverity[%v] = %d, want %d", sev, r.Summary.BySeverity[sev], want)
	}
}

// assertSummaryFixStrategy asserts the count of findings with a given fix strategy in the report summary.
func assertSummaryFixStrategy(t *testing.T, r *Report, fs FixStrategy, want int) {
	t.Helper()

	if r.Summary.ByFixStrategy[fs] != want {
		t.Errorf("Summary.ByFixStrategy[%v] = %d, want %d", fs, r.Summary.ByFixStrategy[fs], want)
	}
}

// assertSummarySuppressed asserts the suppressed count in the report summary.
func assertSummarySuppressed(t *testing.T, r *Report, want int) {
	t.Helper()

	if r.Summary.Suppressed != want {
		t.Errorf("Summary.Suppressed = %d, want %d", r.Summary.Suppressed, want)
	}
}

// assertRangeStartOffset asserts a range's start offset equals expected value.
func assertRangeStartOffset(t *testing.T, r Range, want int) {
	t.Helper()

	if r.Start.Offset != want {
		t.Errorf("Start.Offset = %d, want %d", r.Start.Offset, want)
	}
}

// assertRangeEndOffset asserts a range's end offset equals expected value.
func assertRangeEndOffset(t *testing.T, r Range, want int) {
	t.Helper()

	if r.End.Offset != want {
		t.Errorf("End.Offset = %d, want %d", r.End.Offset, want)
	}
}

// assertFindingPosition asserts a finding's position file and line.
func assertFindingPosition(t *testing.T, f Finding, file string, line int) {
	t.Helper()

	if f.Position.File != file {
		t.Errorf("Position.File = %q, want %q", f.Position.File, file)
	}

	if f.Position.Line != line {
		t.Errorf("Position.Line = %d, want %d", f.Position.Line, line)
	}
}

// assertRelatedPosition asserts a related ref's position file and line.
func assertRelatedPosition(t *testing.T, r RelatedRef, file string, line int) {
	t.Helper()

	if r.Position.File != file {
		t.Errorf("Position.File = %q, want %q", r.Position.File, file)
	}

	if r.Position.Line != line {
		t.Errorf("Position.Line = %d, want %d", r.Position.Line, line)
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
		ID:       ID(id),
		Severity: severity,
	}
}

// MakeSimpleReport creates a Report with the given tool name.
func MakeSimpleReport(toolName string) *Report {
	return NewReport(ToolInfo{Name: toolName})
}

// MakeReport creates a Report with the given tool info.
func MakeReport(toolName, version string) *Report {
	return NewReport(ToolInfo{Name: toolName, Version: version})
}

// MakeNonexistentPosition returns a Position pointing to a nonexistent file at line 1, col 1.
func MakeNonexistentPosition() Position {
	return Position{File: "nonexistent.go", Line: 1, Column: 1}
}

// MakeFindingWithFix creates a Finding with fix-related fields.
func MakeFindingWithFix(id, rule, tool, msg, beforeCode, afterCode, file string, line int) Finding {
	return Finding{
		ID:          ID(id),
		Rule:        RuleName(rule),
		ToolName:    ToolName(tool),
		Message:     msg,
		BeforeCode:  beforeCode,
		AfterCode:   afterCode,
		Position:    Position{File: file, Line: line},
		FixStrategy: FixStrategyDirect,
	}
}

// assertFindingErrorFile asserts the File field of a FindingError.
func assertFindingErrorFile(t *testing.T, err *FindingError, want string) {
	t.Helper()

	if err.File != want {
		t.Errorf("File = %q, want %q", err.File, want)
	}
}

// assertRangeContains asserts a Range contains an offset at the expected position.
func assertRangeContains(t *testing.T, r Range, offset int, file string, expect bool) {
	t.Helper()

	pos := Position{File: file, Offset: offset}
	if r.Contains(pos) != expect {
		if expect {
			t.Errorf("expected Range to contain Position at offset %d", offset)
		} else {
			t.Errorf("expected Range NOT to contain Position at offset %d", offset)
		}
	}
}

// checkProperty runs a property-based test using quick.Check with a deterministic seed.
// For functions with multiple parameters, use checkPropertyAny instead.
func checkProperty[T any](t *testing.T, property func(T) bool) {
	t.Helper()
	checkPropertyAny(t, property)
}

// AssertFindingFields asserts common Finding fields match expected values.
func AssertFindingFields(
	t *testing.T,
	f Finding,
	rule, tool, message string,
	sev Severity,
	pos Position,
) {
	t.Helper()

	if string(f.Rule) != rule {
		t.Errorf("Rule = %q, want %q", f.Rule, rule)
	}

	if string(f.ToolName) != tool {
		t.Errorf("ToolName = %q, want %q", f.ToolName, tool)
	}

	if f.Message != message {
		t.Errorf("Message = %q, want %q", f.Message, message)
	}

	if f.Severity != sev {
		t.Errorf("Severity = %v, want %v", f.Severity, sev)
	}

	if f.Position != pos {
		t.Errorf("Position = %v, want %v", f.Position, pos)
	}
}

// AssertFindingsLenAndIDs asserts a slice has the expected length and element IDs in order.
func AssertFindingsLenAndIDs(t *testing.T, findings []Finding, wantIDs []string, context string) {
	t.Helper()

	if len(findings) != len(wantIDs) {
		t.Fatalf("%s: len = %d, want %d", context, len(findings), len(wantIDs))
	}

	for i, f := range findings {
		if string(f.ID) != wantIDs[i] {
			t.Errorf("%s: [%d].ID = %q, want %q", context, i, f.ID, wantIDs[i])
		}
	}
}

// checkPropertyAny runs a property-based test with any function signature.
func checkPropertyAny(t *testing.T, property any) {
	t.Helper()

	err := quick.Check(property, &quick.Config{
		MaxCount: 1000,
		Rand:     rand.New(rand.NewSource(42)),
	})
	if err != nil {
		t.Error(err)
	}
}
