package finding

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMerge_Empty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	merged := Combine(nil)
	g.Expect(merged).NotTo(BeNil())

	assertFindingsLen(t, "Combine(nil) findings", len(merged.findings), 0)
}

func TestMerge_SingleReport(t *testing.T) {
	t.Parallel()

	r := MakeSimpleReport("tool1")
	r.AddFinding(MakeSimpleFinding("1", SeverityError))

	merged := Combine([]*Report{r})
	assertReportField(t, "single report merge tool", merged.Tool.Name, "tool1")
	assertFindingsLen(t, "single report merge findings", len(merged.findings), 1)
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
	assertFindingsLen(t, "merged findings", len(merged.findings), 2)
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

	assertFindingsLen(t, "deduplicated merge", len(merged.findings), 1)
}

func TestMerge_WithoutDeduplication(t *testing.T) {
	t.Parallel()

	r1 := MakeSimpleReport("tool1")
	r1.AddFinding(MakeSimpleFinding("1", SeverityError))

	r2 := MakeSimpleReport("tool2")
	r2.AddFinding(MakeSimpleFinding("1", SeverityWarning))

	merged := Combine([]*Report{r1, r2}, WithDeduplication(false))
	merged.ComputeSummary()

	assertFindingsLen(t, "non-deduplicated merge", len(merged.findings), 2)
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

	assertFindingsLen(t, "dedup by position", len(merged.findings), 1)
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

	assertFindingsLen(t, "dedup by rule", len(merged.findings), 1)
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
