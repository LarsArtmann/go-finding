package finding

import (
	"slices"
	"testing"

	. "github.com/onsi/gomega"
)

func TestDeduplicateStrategies_BehaviorDiff(t *testing.T) {
	g := NewParallelGomega(t)

	r1 := NewReport(ToolInfo{Name: "govet"})
	for _, tc := range []struct {
		id, rule string
	}{
		{"1", "nilcheck"},
		{"3", "unused"},
	} {
		r1.AddFinding(Finding{
			ID: ID(tc.id), ToolName: "govet", Rule: RuleName(tc.rule),
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

func TestDeduplicateStrategies_EmptyFileNotDeduplicated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		by   DeduplicateBy
	}{
		{"position", DeduplicateByPosition},
		{"rule", DeduplicateByRule},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)

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
				WithDeduplicateBy(tt.by),
			)
			g.Expect(merged.Len()).To(Equal(2))
		})
	}
}

func TestDeduplicateBy_String(t *testing.T) {
	g := NewParallelGomega(t)

	g.Expect(DeduplicateByID.String()).To(Equal("id"))
	g.Expect(DeduplicateByPosition.String()).To(Equal("position"))
	g.Expect(DeduplicateByRule.String()).To(Equal("rule"))
	g.Expect(DeduplicateBy(99).String()).To(Equal("unknown(99)"))
}

func TestCorrelationScore_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		score CorrelationScore
		want  bool
	}{
		{CorrelationScore(0.0), true},
		{CorrelationScore(0.5), true},
		{CorrelationScore(1.0), true},
		{CorrelationScore(-0.1), false},
		{CorrelationScore(1.1), false},
	}

	for _, tt := range tests {
		if tt.score.IsValid() != tt.want {
			t.Errorf("CorrelationScore(%v).IsValid() = %v, want %v", tt.score, !tt.want, tt.want)
		}
	}
}

func TestCorrelationScore_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		score CorrelationScore
		want  string
	}{
		{CorrelationScore(0.0), "0.00"},
		{CorrelationScore(0.75), "0.75"},
		{CorrelationScore(1.0), "1.00"},
	}

	for _, tt := range tests {
		if tt.score.String() != tt.want {
			t.Errorf("CorrelationScore(%v).String() = %q, want %q", tt.score, tt.score.String(), tt.want)
		}
	}
}

func collectIDs(r *Report) []string {
	var ids []string

	for f := range r.All() {
		ids = append(ids, string(f.ID))
	}

	return ids
}

func makeFinding(id, tool, rule, file string, line int) Finding {
	return Finding{
		ID:       ID(id),
		ToolName: ToolName(tool),
		Rule:     RuleName(rule),
		Position: Position{File: FilePath(file), Line: line},
	}
}

func TestMergeIter(t *testing.T) {
	g := NewParallelGomega(t)

	r1 := NewReport(ToolInfo{Name: "a"})
	r1.AddFinding(Finding{ID: "1", ToolName: "a", Position: Position{File: "f.go", Line: 1}})
	r1.AddFinding(Finding{ID: "2", ToolName: "a", Position: Position{File: "f.go", Line: 2}})

	r2 := NewReport(ToolInfo{Name: "b"})
	r2.AddFinding(Finding{ID: "3", ToolName: "b", Position: Position{File: "f.go", Line: 3}})

	collected := slices.Collect(MergeIter([]*Report{r1, r2}))

	g.Expect(collected).To(HaveLen(3))
	g.Expect(collected[0].ID).To(Equal(ID("1")))
	g.Expect(collected[2].ID).To(Equal(ID("3")))
}

func TestMergeIter_WithDedup(t *testing.T) {
	g := NewParallelGomega(t)

	r1 := NewReport(ToolInfo{Name: "a"})
	r1.AddFinding(Finding{ID: "dup", ToolName: "a"})

	r2 := NewReport(ToolInfo{Name: "b"})
	r2.AddFinding(Finding{ID: "dup", ToolName: "b"})

	collected := slices.Collect(MergeIter([]*Report{r1, r2}, WithDeduplication(true)))

	g.Expect(collected).To(HaveLen(1))
}

func TestMergeIter_Empty(t *testing.T) {
	g := NewParallelGomega(t)

	collected := slices.Collect(MergeIter(nil))
	g.Expect(collected).To(BeEmpty())
}

func TestMergeIter_EarlyStop(t *testing.T) {
	g := NewParallelGomega(t)

	r1 := NewReport(ToolInfo{Name: "a"})
	r1.AddFinding(Finding{ID: "1"})
	r1.AddFinding(Finding{ID: "2"})
	r1.AddFinding(Finding{ID: "3"})

	count := 0

	for range MergeIter([]*Report{r1}) {
		count++
		if count == 1 {
			break
		}
	}

	g.Expect(count).To(Equal(1))
}
