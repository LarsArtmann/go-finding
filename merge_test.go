package finding

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerge_Empty(t *testing.T) {
	t.Parallel()

	merged := Merge(nil)
	assert.NotNil(t, merged)

	assertFindingsLen(t, "Merge(nil) findings", len(merged.Findings), 0)
}

func TestMerge_SingleReport(t *testing.T) {
	t.Parallel()

	r := MakeSimpleReport("tool1")
	r.AddFinding(MakeSimpleFinding("1", SeverityError))

	merged := Merge([]*Report{r})
	assertReportField(t, "single report merge tool", merged.Tool.Name, "tool1")
	assertFindingsLen(t, "single report merge findings", len(merged.Findings), 1)
}

func TestMerge_MultipleReports(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(Finding{ID: "1", Severity: SeverityError, Position: Position{File: "a.go"}})

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(Finding{ID: "2", Severity: SeverityWarning, Position: Position{File: "b.go"}})

	merged := Merge([]*Report{r1, r2})
	merged.ComputeSummary()

	assertReportField(t, "merged tool name", merged.Tool.Name, "merged")
	assertFindingsLen(t, "merged findings", len(merged.Findings), 2)
	assertReportField(t, "merged files", merged.Summary.FilesAffected, 2)
}

func TestMerge_WithDeduplication(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(MakeSimpleFinding("1", SeverityError))

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(MakeSimpleFinding("1", SeverityWarning))

	merged := Merge([]*Report{r1, r2}, WithDeduplication(true))
	merged.ComputeSummary()

	assertFindingsLen(t, "deduplicated merge", len(merged.Findings), 1)
}

func TestMerge_WithoutDeduplication(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(MakeSimpleFinding("1", SeverityError))

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(MakeSimpleFinding("1", SeverityWarning))

	merged := Merge([]*Report{r1, r2}, WithDeduplication(false))
	merged.ComputeSummary()

	assertFindingsLen(t, "non-deduplicated merge", len(merged.Findings), 2)
}

func TestMerge_DeduplicateByPosition(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(MakeSimpleFinding("a", SeverityError))

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(MakeSimpleFinding("b", SeverityWarning))

	merged := Merge([]*Report{r1, r2}, WithDeduplicateBy(DeduplicateByPosition))
	merged.ComputeSummary()

	assertFindingsLen(t, "dedup by position", len(merged.Findings), 1)
}

func TestMerge_DeduplicateByRule(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(
		Finding{
			ID:       "a",
			Rule:     "nilcheck",
			Severity: SeverityError,
			Position: Position{File: "a.go", Line: 10},
		},
	)

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(
		Finding{
			ID:       "b",
			Rule:     "nilcheck",
			Severity: SeverityWarning,
			Position: Position{File: "a.go", Line: 10},
		},
	)

	merged := Merge([]*Report{r1, r2}, WithDeduplicateBy(DeduplicateByRule))
	merged.ComputeSummary()

	assertFindingsLen(t, "dedup by rule", len(merged.Findings), 1)
}

func TestDedupKey(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:       "test-id",
		Rule:     "rule1",
		ToolName: "tool1",
		Position: Position{File: "a.go", Line: 10, Column: 5},
	}

	tests := []struct {
		name string
		by   DeduplicateBy
		want string
	}{
		{"by ID", DeduplicateByID, "test-id"},
		{"by position", DeduplicateByPosition, "tool1:a.go:10:5"},
		{"by rule", DeduplicateByRule, "rule1:a.go:10:5"},
		{"default", DeduplicateBy(99), "test-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := dedupKey(f, MergeOptions{DeduplicateBy: tt.by})
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeduplicateStrategiesDistinct(t *testing.T) {
	t.Parallel()

	f := Finding{
		Rule:     "rule1",
		Position: Position{File: "a.go", Line: 10, Column: 5},
	}

	posKey := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByPosition})
	ruleKey := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByRule})

	assert.NotEqual(t, posKey, ruleKey)
}

func TestDeduplicateStrategies_BehaviorDiff(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "govet"})
	r1.AddFinding(Finding{
		ID: "1", ToolName: "govet", Rule: "nilcheck",
		Position: Position{File: "a.go", Line: 10},
	})
	r1.AddFinding(Finding{
		ID: "3", ToolName: "govet", Rule: "unused",
		Position: Position{File: "a.go", Line: 10},
	})

	r2 := NewReport(ToolInfo{Name: "staticcheck"})
	r2.AddFinding(Finding{
		ID: "2", ToolName: "staticcheck", Rule: "nilcheck",
		Position: Position{File: "a.go", Line: 10},
	})

	reports := []*Report{r1, r2}

	byPos := Merge(reports, WithDeduplicateBy(DeduplicateByPosition))
	byRule := Merge(reports, WithDeduplicateBy(DeduplicateByRule))

	assert.Len(
		t, collectIDs(byPos), 2,
		"by position: govet:a.go:10 and staticcheck:a.go:10 are distinct",
	)

	assert.Len(
		t, collectIDs(byRule), 2,
		"by rule: nilcheck:a.go:10 and unused:a.go:10 are distinct",
	)
}

func collectIDs(r *Report) []string {
	var ids []string

	for f := range r.All() {
		ids = append(ids, f.ID)
	}

	return ids
}

func makeFinding(id, tool, rule, file string, line int) Finding {
	return Finding{
		ID:       id,
		ToolName: tool,
		Rule:     rule,
		Position: Position{File: file, Line: line},
	}
}

func TestCorrelate(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		makeFinding("1", "govet", "nilcheck", "a.go", 10),
		makeFinding("2", "staticcheck", "nilcheck", "a.go", 12),
		makeFinding("3", "govet", "unused", "a.go", 50),
	}

	correlations := Correlate(findings)
	require.Len(t, correlations, 1, "correlations")

	c := correlations[0]
	assert.Len(t, c.FindingIDs, 2)
	assert.Equal(t, "same file, nearby lines", c.Reason)

	wantConf := 1.0 - (2.0 / 5.0)
	assert.InDelta(t, wantConf, c.Confidence, 0.0001)

	hasID := func(id string) bool {
		return slices.Contains(c.FindingIDs, id)
	}

	assert.True(t, hasID("1"))
	assert.True(t, hasID("2"))
}

func TestCorrelate_TooFewFindings(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", ToolName: "govet", Position: Position{File: "a.go", Line: 10}},
	}

	correlations := Correlate(findings)
	assert.Empty(t, correlations)
}

func TestCloneFindings_EmptySlice(t *testing.T) {
	t.Parallel()

	cloned := cloneFindings(nil)
	if cloned != nil {
		t.Errorf("cloneFindings(nil) = %v, want nil", cloned)
	}

	cloned = cloneFindings([]Finding{})
	if cloned != nil {
		t.Errorf("cloneFindings(empty) = %v, want nil", cloned)
	}
}

func TestCloneFindings_DeepCopy(t *testing.T) {
	t.Parallel()

	original := []Finding{
		{
			ID: "1", Rule: "r1", ToolName: "t1", Message: "m1",
			Severity: SeverityError, Position: Position{File: "a.go", Line: 10},
		},
		{
			ID: "2", Rule: "r2", ToolName: "t2", Message: "m2",
			Severity: SeverityWarning, Position: Position{File: "b.go", Line: 20},
		},
	}

	cloned := cloneFindings(original)
	require.Len(t, cloned, 2)

	if !cloned[0].Equal(original[0]) {
		t.Error("cloned[0] should be Equal to original[0]")
	}

	// Mutate clone — original should be unaffected.
	cloned[0].Metadata = map[string]string{"key": "mutated"}
	if original[0].Metadata != nil {
		t.Error("mutating clone Metadata should not affect original")
	}
}
