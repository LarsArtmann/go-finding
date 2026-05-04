package finding

import (
	"encoding/json"
	"math/rand"
	"testing"
	"testing/quick"
)

// MakeFinding creates a Finding with common fields for testing.
func MakeFinding(id, rule, tool, message string, severity Severity) Finding {
	return Finding{
		ID:       id,
		Rule:     rule,
		ToolName: tool,
		Message:  message,
		Severity: severity,
	}
}

// MakeFindingWithPos creates a Finding with position info.
func MakeFindingWithPos(
	id, rule, tool, message string,
	severity Severity,
	file string,
	line, col int,
) Finding {
	return Finding{
		ID:       id,
		Rule:     rule,
		ToolName: tool,
		Message:  message,
		Severity: severity,
		Position: Position{File: file, Line: line, Column: col},
	}
}

// MakeFindings creates a slice of findings from field tuples.
func MakeFindings(fields []struct {
	ID, Rule, Tool, Message string
	Severity                Severity
},
) []Finding {
	result := make([]Finding, len(fields))
	for i, f := range fields {
		result[i] = MakeFinding(f.ID, f.Rule, f.Tool, f.Message, f.Severity)
	}

	return result
}

// RangesEq checks if two ranges are equal.
func RangesEq(a, b Range) bool {
	return a.Start.Line == b.Start.Line &&
		a.Start.Column == b.Start.Column &&
		a.Start.File == b.Start.File &&
		a.End.Line == b.End.Line &&
		a.End.Column == b.End.Column &&
		a.End.File == b.End.File
}

// AssertRangesEq asserts two ranges are equal in a test.
func AssertRangesEq(t *testing.T, got, want Range) {
	t.Helper()

	if !RangesEq(got, want) {
		t.Errorf("Range = %+v, want %+v", got, want)
	}
}

// AssertRangeLinesEq asserts two ranges have equal lines in a test.
func AssertRangeLinesEq(t *testing.T, got, want Range) {
	t.Helper()

	if !RangeLinesEq(got, want) {
		t.Errorf("Range lines = %+v, want %+v", got, want)
	}
}

// RunEqualTests runs a table-driven equality test for types with Equal methods.
func RunEqualTests[T any](t *testing.T, tests []struct {
	name string
	a, b T
	want bool
}, eqFunc func(a, b T) bool, formatName string,
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := eqFunc(tt.a, tt.b); got != tt.want {
				t.Errorf("%s.Equal() = %v, want %v", formatName, got, tt.want)
			}
		})
	}
}

func unmarshalJSON(t *testing.T, data []byte, v any) {
	t.Helper()

	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

// AssertLen asserts a slice has the expected length.
func AssertLen[T any](t *testing.T, got []T, want int, msg string) {
	t.Helper()

	if len(got) != want {
		t.Errorf("%s = %d, want %d", msg, len(got), want)
	}
}

// AssertEmpty asserts a slice is empty.
func AssertEmpty[T any](t *testing.T, got []T, msg string) {
	t.Helper()

	if len(got) != 0 {
		t.Errorf("%s = %d, want 0", msg, len(got))
	}
}

// AssertFindingsIDs asserts findings have expected IDs in order.
func AssertFindingsIDs(t *testing.T, findings []Finding, want []string) {
	t.Helper()

	if len(findings) != len(want) {
		t.Errorf("findings len = %d, want %d", len(findings), len(want))
		return
	}
	for i, f := range findings {
		if f.ID != want[i] {
			t.Errorf("findings[%d].ID = %q, want %q", i, f.ID, want[i])
		}
	}
}

// RunCompareTests runs a table-driven comparison test for types with Compare methods.
func RunCompareTests[T any](t *testing.T, tests []struct {
	name string
	a, b T
	want int
}, cmpFunc func(a, b T) int, formatName string,
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := cmpFunc(tt.a, tt.b); got != tt.want {
				t.Errorf("%s.Compare() = %d, want %d", formatName, got, tt.want)
			}
		})
	}
}

// assertFindingsLen asserts the length of a findings slice matches expected.
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

	if f.Rule != rule {
		t.Errorf("Rule = %q, want %q", f.Rule, rule)
	}
	if f.ToolName != tool {
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
		if f.ID != wantIDs[i] {
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
