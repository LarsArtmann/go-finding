package finding

import (
	"encoding/json"
	"testing"
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

// RunCompareTests runs a table-driven comparison test for types with Compare methods.
func RunCompareTests[T any](t *testing.T, tests []struct {
	name string
	a, b T
	want int
}, cmpFunc func(a, b T) int, formatName string,
) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := cmpFunc(tt.a, tt.b); got != tt.want {
				t.Errorf("%s.Compare() = %d, want %d", formatName, got, tt.want)
			}
		})
	}
}
