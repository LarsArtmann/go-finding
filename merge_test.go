package finding

import (
	"slices"
	"testing"

	. "github.com/onsi/gomega"
)

func TestMerge_Empty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	merged := Combine(nil)
	g.Expect(merged).NotTo(BeNil())

	assertFindingsLen(t, "Combine(nil) findings", len(merged.Findings), 0)
}

func TestMerge_SingleReport(t *testing.T) {
	t.Parallel()

	r := MakeSimpleReport("tool1")
	r.AddFinding(MakeSimpleFinding("1", SeverityError))

	merged := Combine([]*Report{r})
	assertReportField(t, "single report merge tool", merged.Tool.Name, "tool1")
	assertFindingsLen(t, "single report merge findings", len(merged.Findings), 1)
}

func TestMerge_MultipleReports(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(Finding{ID: "1", Severity: SeverityError, Position: Position{File: "a.go"}})

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(Finding{ID: "2", Severity: SeverityWarning, Position: Position{File: "b.go"}})

	merged := Combine([]*Report{r1, r2})
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

	merged := Combine([]*Report{r1, r2}, WithDeduplication(true))
	merged.ComputeSummary()

	assertFindingsLen(t, "deduplicated merge", len(merged.Findings), 1)
}

func TestMerge_WithoutDeduplication(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(MakeSimpleFinding("1", SeverityError))

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(MakeSimpleFinding("1", SeverityWarning))

	merged := Combine([]*Report{r1, r2}, WithDeduplication(false))
	merged.ComputeSummary()

	assertFindingsLen(t, "non-deduplicated merge", len(merged.Findings), 2)
}

func TestMerge_DeduplicateByPosition(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(
		Finding{ID: "a", Severity: SeverityError, Position: Position{File: "a.go", Line: 10}},
	)

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(
		Finding{ID: "b", Severity: SeverityWarning, Position: Position{File: "a.go", Line: 10}},
	)

	merged := Combine([]*Report{r1, r2}, WithDeduplicateBy(DeduplicateByPosition))
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

	merged := Combine([]*Report{r1, r2}, WithDeduplicateBy(DeduplicateByRule))
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
			g := NewWithT(t)

			got, _ := dedupKey(f, MergeOptions{DeduplicateBy: tt.by})
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestDeduplicateStrategiesDistinct(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	f := Finding{
		Rule:     "rule1",
		Position: Position{File: "a.go", Line: 10, Column: 5},
	}

	posKey, _ := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByPosition})
	ruleKey, _ := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByRule})

	g.Expect(ruleKey).NotTo(Equal(posKey))
}

func TestDedupKey_EmptyID(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{Rule: "r", Position: Position{File: "a.go", Line: 1}}

	key, ok := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByID})
	g.Expect(ok).To(BeFalse())
	g.Expect(key).To(BeEmpty())
}

func TestDeduplicateByID_EmptyIDNotDeduplicated(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r1 := NewReport(ToolInfo{Name: "t1"})
	for _, msg := range []string{"m1", "m2"} {
		r1.AddFinding(Finding{
			ID: "", Rule: "r", ToolName: "t1", Message: msg,
			Severity: SeverityInfo, Position: Position{File: "a.go", Line: 1},
		})
	}

	merged := Combine([]*Report{r1}, WithDeduplication(true), WithDeduplicateBy(DeduplicateByID))
	g.Expect(merged.Len()).To(Equal(2))
}

func TestDeduplicateStrategies_BehaviorDiff(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r1 := NewReport(ToolInfo{Name: "govet"})
	for _, tc := range []struct {
		id, rule string
	}{
		{"1", "nilcheck"},
		{"3", "unused"},
	} {
		r1.AddFinding(Finding{
			ID: tc.id, ToolName: "govet", Rule: tc.rule,
			Position: Position{File: "a.go", Line: 10},
		})
	}

	r2 := NewReport(ToolInfo{Name: "staticcheck"})
	r2.AddFinding(Finding{
		ID: "2", ToolName: "staticcheck", Rule: "nilcheck",
		Position: Position{File: "a.go", Line: 10},
	})

	reports := []*Report{r1, r2}

	byPos := Combine(reports, WithDeduplicateBy(DeduplicateByPosition))
	byRule := Combine(reports, WithDeduplicateBy(DeduplicateByRule))

	g.Expect(collectIDs(byPos)).To(HaveLen(2))

	g.Expect(collectIDs(byRule)).To(HaveLen(2))
}

func TestDeduplicateByPosition_EmptyFileNotDeduplicated(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r1 := NewReport(ToolInfo{Name: "t1"})
	r1.AddFinding(
		Finding{ID: "1", ToolName: "t1", Rule: "r", Message: "m1", Position: Position{Line: 1}},
	)
	r1.AddFinding(
		Finding{ID: "2", ToolName: "t1", Rule: "r", Message: "m2", Position: Position{Line: 1}},
	)

	merged := Combine(
		[]*Report{r1},
		WithDeduplication(true),
		WithDeduplicateBy(DeduplicateByPosition),
	)
	g.Expect(merged.Len()).To(Equal(2))
}

func TestDeduplicateByRule_EmptyFileNotDeduplicated(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r1 := NewReport(ToolInfo{Name: "t1"})
	r1.AddFinding(
		Finding{ID: "1", ToolName: "t1", Rule: "r", Message: "m1", Position: Position{Line: 1}},
	)
	r1.AddFinding(
		Finding{ID: "2", ToolName: "t1", Rule: "r", Message: "m2", Position: Position{Line: 1}},
	)

	merged := Combine([]*Report{r1}, WithDeduplication(true), WithDeduplicateBy(DeduplicateByRule))
	g.Expect(merged.Len()).To(Equal(2))
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
	g := NewWithT(t)
	t.Parallel()

	findings := []Finding{
		makeFinding("1", "govet", "nilcheck", "a.go", 10),
		makeFinding("2", "staticcheck", "nilcheck", "a.go", 12),
		makeFinding("3", "govet", "unused", "a.go", 50),
	}

	correlations := Correlate(findings)
	g.Expect(correlations).To(HaveLen(1))

	c := correlations[0]
	g.Expect(c.FindingIDs).To(HaveLen(2))
	g.Expect(c.Reason).To(Equal("same file, nearby lines"))

	wantConf := 1.0 - (2.0 / 5.0)
	g.Expect(c.Score).To(BeNumerically("~", wantConf, 0.0001))

	hasID := func(id string) bool {
		return slices.Contains(c.FindingIDs, id)
	}

	g.Expect(hasID("1")).To(BeTrue())
	g.Expect(hasID("2")).To(BeTrue())
}

func TestCorrelate_TooFewFindings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	findings := []Finding{
		{ID: "1", ToolName: "govet", Position: Position{File: "a.go", Line: 10}},
	}

	correlations := Correlate(findings)
	g.Expect(correlations).To(BeEmpty())
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
	g := NewWithT(t)
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
	g.Expect(cloned).To(HaveLen(2))

	if !cloned[0].Equal(original[0]) {
		t.Error("cloned[0] should be Equal to original[0]")
	}

	// Mutate clone — original should be unaffected.
	cloned[0].Metadata = map[string]string{"key": "mutated"}
	if original[0].Metadata != nil {
		t.Error("mutating clone Metadata should not affect original")
	}
}

func TestMerge_DeduplicateByID_EmptyIDs(t *testing.T) {
	t.Parallel()

	f1 := Finding{
		Rule: "R1", ToolName: "tool-a", Message: "msg1",
		Severity: SeverityError, Position: Pos("a.go", 1, 1),
	}
	f2 := Finding{
		Rule: "R2", ToolName: "tool-b", Message: "msg2",
		Severity: SeverityWarning, Position: Pos("b.go", 2, 1),
	}

	r1 := NewReport(ToolInfo{Name: "a"})
	r1.AddFinding(f1)
	r2 := NewReport(ToolInfo{Name: "b"})
	r2.AddFinding(f2)

	merged := Combine(
		[]*Report{r1, r2},
		WithDeduplication(true),
		WithDeduplicateBy(DeduplicateByID),
	)

	if len(merged.Findings) != 2 {
		t.Errorf(
			"got %d findings, want 2 (empty-ID findings should NOT collide)",
			len(merged.Findings),
		)
	}
}
