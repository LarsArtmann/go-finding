package finding

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFinding(t *testing.T) {
	t.Parallel()

	pos := Position{File: "main.go", Line: 42, Column: 5, Offset: 100}
	f := NewFinding("nilcheck", "govet", "possible nil dereference", SeverityError, pos)

	assert.Equal(t, "nilcheck", f.Rule)
	assert.Equal(t, "govet", f.ToolName)
	assert.Equal(t, "possible nil dereference", f.Message)
	assert.Equal(t, SeverityError, f.Severity)
	assert.Equal(t, pos, f.Position)
	assert.Equal(t, FixStrategyNone, f.FixStrategy)
	assert.NotEmpty(t, f.ID)
	assert.True(t, strings.Contains(f.ID, "govet") && strings.Contains(f.ID, "nilcheck"),
		"ID should contain tool name and rule")
}

func TestSuppressionKind_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind SuppressionKind
		want bool
	}{
		{SuppressionInSource, true},
		{SuppressionInConfig, true},
		{SuppressionInReview, true},
		{SuppressionKind(""), false},
		{SuppressionKind("unknown"), false},
		{SuppressionKind("in-source-extra"), false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.kind.IsValid(), "SuppressionKind IsValid")
	}
}

func TestSeverity_GreaterThanOrEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b  Severity
		want  bool
		label string
	}{
		{SeverityCritical, SeverityCritical, true, "critical >= critical"},
		{SeverityCritical, SeverityError, true, "critical >= error"},
		{SeverityError, SeverityWarning, true, "error >= warning"},
		{SeverityWarning, SeverityInfo, true, "warning >= info"},
		{SeverityInfo, SeverityInfo, true, "info >= info"},
		{SeverityInfo, SeverityWarning, false, "info < warning"},
		{SeverityWarning, SeverityError, false, "warning < error"},
		{Severity("unknown"), SeverityInfo, false, "invalid >= valid returns false"},
		{SeverityInfo, Severity("unknown"), false, "valid >= invalid returns false"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.a.GreaterThanOrEqual(tt.b))
	}
}

func TestSeverity_LessThanOrEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b  Severity
		want  bool
		label string
	}{
		{SeverityInfo, SeverityInfo, true, "info <= info"},
		{SeverityInfo, SeverityWarning, true, "info <= warning"},
		{SeverityWarning, SeverityError, true, "warning <= error"},
		{SeverityError, SeverityCritical, true, "error <= critical"},
		{SeverityCritical, SeverityCritical, true, "critical <= critical"},
		{SeverityCritical, SeverityError, false, "critical > error"},
		{SeverityError, SeverityWarning, false, "error > warning"},
		{Severity("unknown"), SeverityInfo, false, "invalid <= valid returns false"},
		{SeverityInfo, Severity("unknown"), false, "valid <= invalid returns false"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.a.LessThanOrEqual(tt.b))
	}
}

func TestFinding_String(t *testing.T) {
	t.Parallel()

	f := Finding{
		Severity: SeverityError,
		ToolName: "govet",
		Rule:     "nilcheck",
		Position: Position{File: "main.go", Line: 42, Column: 5},
		Message:  "possible nil dereference",
	}

	got := f.String()
	want := "error govet [nilcheck] main.go:42:5: possible nil dereference"

	assert.Equal(t, want, got)
}

func TestReport_AddFindings(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})

	r.AddFindings([]Finding{
		{ID: "1", Message: "first"},
		{ID: "2", Message: "second"},
		{ID: "3", Message: "third"},
	})

	require.Len(t, r.Findings, 3, "Findings")

	assert.Equal(t, "1", r.Findings[0].ID)
	assert.Equal(t, "3", r.Findings[2].ID)
}

func TestReport_AddFindings_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFindings(nil)
	AssertEmpty(t, r.Findings, "AddFindings(nil): Findings")

	r.AddFindings([]Finding{})
	AssertEmpty(t, r.Findings, "AddFindings(empty): Findings")
}

func TestMerge_NilReports(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(MakeSimpleFinding("1", SeverityError))

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(MakeSimpleFinding("2", SeverityWarning))

	merged := Merge([]*Report{r1, nil, r2})
	merged.ComputeSummary()

	assertFindingsLen(t, "merge with nil reports", len(merged.Findings), 2)
}
