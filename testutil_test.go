package finding

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/onsi/gomega"
)

// AssertErrIsIO asserts an error is of type ErrIO using Gomega.
func AssertErrIsIO(g *gomega.GomegaWithT, err error) {
	g.Expect(errors.Is(err, ErrIO)).To(gomega.BeTrue())
}

// AssertErrNotIO asserts an error is NOT of type ErrIO using Gomega.
func AssertErrNotIO(g *gomega.GomegaWithT, err error) {
	g.Expect(errors.Is(err, ErrIO)).To(gomega.BeFalse())
}

// AssertErrIsCanceled asserts an error is context.Canceled using Gomega.
func AssertErrIsCanceled(g *gomega.GomegaWithT, err error) {
	g.Expect(errors.Is(err, context.Canceled)).To(gomega.BeTrue())
}

// AssertErrNotCanceled asserts an error is NOT context.Canceled using Gomega.
func AssertErrNotCanceled(g *gomega.GomegaWithT, err error) {
	g.Expect(errors.Is(err, context.Canceled)).To(gomega.BeFalse())
}

// AssertFindingSeverity asserts a finding has the expected severity using Gomega.
func AssertFindingSeverity(g *gomega.GomegaWithT, f Finding, sev Severity) {
	g.Expect(f.Severity).To(gomega.Equal(sev))
}

// AssertFindingConfidence asserts a finding has the expected confidence using Gomega.
func AssertFindingConfidence(g *gomega.GomegaWithT, f Finding, confidence, tolerance float64) {
	g.Expect(f.Confidence).To(gomega.BeNumerically("~", confidence, tolerance))
}

// AssertFindingPosition asserts a finding has the expected file and line using Gomega.
func AssertFindingPosition(g *gomega.GomegaWithT, f Finding, file string, line int) {
	g.Expect(f.Position.File).To(gomega.Equal(file))
	g.Expect(f.Position.Line).To(gomega.Equal(line))
}

// AssertErrContains asserts an error message contains a substring using Gomega.
func AssertErrContains(g *gomega.GomegaWithT, err error, substr string) {
	g.Expect(err.Error()).To(gomega.ContainSubstring(substr))
}

// AssertErrIs asserts an error matches a specific error using errors.Is via Gomega.
func AssertErrIs[T error](g *gomega.GomegaWithT, err error, target T) {
	g.Expect(errors.Is(err, target)).To(gomega.BeTrue())
}

// MakeFindingWithID creates a Finding with just an ID and Severity.
func MakeFindingWithID(id string, severity Severity) Finding {
	return Finding{ID: ID(id), Severity: severity}
}

// MakeFindingsWithIDs creates a slice of Findings with sequential IDs and given severity.
func MakeFindingsWithIDs(count int, severity Severity) []Finding {
	findings := make([]Finding, count)
	for i := range count {
		findings[i] = Finding{ID: ID(string(rune('1' + i))), Severity: severity}
	}

	return findings
}

// MakeFinding creates a Finding with common fields for testing.
func MakeFinding(id, rule, tool, message string, severity Severity) Finding {
	return Finding{
		ID:       ID(id),
		Rule:     RuleName(rule),
		ToolName: ToolName(tool),
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
		ID:       ID(id),
		Rule:     RuleName(rule),
		ToolName: ToolName(tool),
		Message:  message,
		Severity: severity,
		Position: Position{File: FilePath(file), Line: line, Column: col},
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

	err := json.Unmarshal(data, v)
	if err != nil {
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
		if string(f.ID) != want[i] {
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

// AssertPanics asserts that fn panics when executed, reporting msg on failure.
func AssertPanics(t *testing.T, msg string, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Error(msg)
		}
	}()

	fn()
}
