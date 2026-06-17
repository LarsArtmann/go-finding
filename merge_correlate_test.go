package finding

import (
	"slices"
	"testing"

	. "github.com/onsi/gomega"
)

func makeRangeFinding(id, tool, rule, file string, startLine, endLine int) Finding {
	return Finding{
		ID: id, ToolName: tool, Rule: rule,
		Severity: SeverityError, Message: "test",
		Position: Position{File: file, Line: startLine},
		Range:    &Range{Start: Position{File: file, Line: startLine}, End: Position{File: file, Line: endLine}},
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

func TestCorrelate_OverlappingRanges(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	findings := []Finding{
		makeRangeFinding("1", "govet", "r1", "a.go", 10, 20),
		makeRangeFinding("2", "staticcheck", "r2", "a.go", 15, 25),
		makeRangeFinding("3", "govet", "r3", "a.go", 30, 40),
	}

	correlations := Correlate(findings)
	g.Expect(correlations).To(HaveLen(1))
	g.Expect(correlations[0].Reason).To(Equal("overlapping ranges"))
	g.Expect(correlations[0].Score).To(BeNumerically(">", 0.5))
}

func TestCorrelate_PointWithinRange(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	findings := []Finding{
		makeRangeFinding("1", "govet", "r1", "a.go", 10, 20),
		makeFinding("2", "staticcheck", "r2", "a.go", 15),
		makeFinding("3", "staticcheck", "r3", "a.go", 50),
	}

	correlations := Correlate(findings)
	g.Expect(correlations).To(HaveLen(1))
	g.Expect(correlations[0].Reason).To(Equal("point within range"))
}

func TestCorrelate_NoCorrelations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		findings []Finding
	}{
		{
			name: "NonOverlappingRanges",
			findings: []Finding{
				makeRangeFinding("1", "govet", "r1", "a.go", 10, 15),
				makeRangeFinding("2", "staticcheck", "r2", "a.go", 20, 25),
			},
		},
		{
			name: "SameToolFiltered",
			findings: []Finding{
				makeRangeFinding("1", "govet", "r1", "a.go", 10, 20),
				makeRangeFinding("2", "govet", "r2", "a.go", 15, 25),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := NewWithT(t)

			correlations := Correlate(tt.findings)
			g.Expect(correlations).To(BeEmpty())
		})
	}
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
