package finding

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

func FuzzFilter(f *testing.F) {
	f.Add("warning", "cat1", "file1.go", "tool1")
	f.Add("error", "cat2", "file2.go", "tool2")
	f.Add("info", "", "", "")

	f.Fuzz(func(t *testing.T, sevStr, category, file, tool string) {
		g := NewWithT(t)
		var severity Severity

		switch sevStr {
		case "info":
			severity = SeverityInfo
		case string(SeverityWarning):
			severity = SeverityWarning
		case string(SeverityError):
			severity = SeverityError
		case "critical":
			severity = SeverityCritical
		default:
			severity = SeverityInfo
		}

		findings := []Finding{
			{
				ID:       "F1",
				Severity: severity,
				Category: Category(category),
				ToolName: tool,
				Position: Position{File: file, Line: 1},
			},
			{
				ID:       "F2",
				Severity: SeverityWarning,
				Category: "default",
				ToolName: "default-tool",
				Position: Position{File: "default.go", Line: 1},
			},
		}

		all := Filter(findings)
		g.Expect(all).To(HaveLen(len(findings)))

		// Property: filtered results all match the predicate
		byCat := Filter(findings, ByCategory(Category(category)))
		for _, f := range byCat {
			g.Expect(f.Category).To(Equal(Category(category)))
		}

		// Property: BySeverity results match exactly
		bySev := Filter(findings, BySeverity(severity))
		for _, f := range bySev {
			g.Expect(f.Severity).To(Equal(severity))
		}

		// Property: combined predicates = intersection
		both := Filter(findings, ByCategory(Category(category)), BySeverity(severity))
		for _, f := range both {
			g.Expect(f.Category).To(Equal(Category(category)))
			g.Expect(f.Severity).To(Equal(severity))
		}

		g.Expect(len(both)).To(BeNumerically("<=", len(byCat)))
		g.Expect(len(both)).To(BeNumerically("<=", len(bySev)))
	})
}

func FuzzGroupBy(f *testing.F) {
	f.Add("key1", "key2", "key1")
	f.Add("", "", "")
	f.Add("a", "a", "a")

	f.Fuzz(func(t *testing.T, key1, key2, key3 string) {
		g := NewWithT(t)
		findings := []Finding{
			{ID: "F1", Category: Category(key1)},
			{ID: "F2", Category: Category(key2)},
			{ID: "F3", Category: Category(key3)},
		}

		groups := GroupBy(findings, func(f Finding) string { return string(f.Category) })

		total := 0
		for _, grp := range groups {
			total += len(grp)
		}

		g.Expect(total).To(Equal(len(findings)))

		for _, grp := range groups {
			key := grp[0].Category
			for _, f := range grp {
				g.Expect(f.Category).To(Equal(key))
			}
		}
	})
}

func FuzzFilterByFile(f *testing.F) {
	f.Add("main.go", "main.go", "test.go")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, target, file1, file2 string) {
		g := NewWithT(t)
		findings := []Finding{
			{ID: "F1", Position: Position{File: file1}},
			{ID: "F2", Position: Position{File: file2}},
		}

		result := Filter(findings, ByFile(target))
		for _, f := range result {
			g.Expect(f.Position.File).To(Equal(target))
		}
	})
}

func FuzzGroupByFile(f *testing.F) {
	f.Add("a.go", "b.go", "a.go")

	f.Fuzz(func(t *testing.T, file1, file2, file3 string) {
		g := NewWithT(t)
		findings := []Finding{
			{ID: "F1", Position: Position{File: file1}},
			{ID: "F2", Position: Position{File: file2}},
			{ID: "F3", Position: Position{File: file3}},
		}

		groups := GroupByFile(findings)

		total := 0
		for _, grp := range groups {
			total += len(grp)
		}

		g.Expect(total).To(Equal(len(findings)))

		if file1 == file2 && file2 == file3 {
			g.Expect(groups).To(HaveLen(1))
		}
	})
}

func FuzzMerge_DedupByID(f *testing.F) {
	f.Add("F1", "F1", "F2")
	f.Add("unique1", "unique2", "unique3")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, id1, id2, id3 string) {
		g := NewWithT(t)
		r1 := NewReport(ToolInfo{Name: "tool1"})
		r1.AddFinding(Finding{ID: id1, Severity: SeverityWarning})
		r1.AddFinding(Finding{ID: id2, Severity: SeverityError})

		r2 := NewReport(ToolInfo{Name: "tool2"})
		r2.AddFinding(Finding{ID: id3, Severity: SeverityInfo})

		merged := Merge(
			[]*Report{r1, r2},
			WithDeduplication(true),
			WithDeduplicateBy(DeduplicateByID),
		)

		seen := make(map[string]int)
		for _, f := range merged.Findings {
			if f.ID == "" {
				continue
			}

			seen[f.ID]++
			g.Expect(seen[f.ID]).To(BeNumerically("<=", 1))
		}

		// Without dedup: total should be sum
		noDedup := Merge([]*Report{r1, r2}, WithDeduplication(false))
		g.Expect(noDedup.Findings).To(HaveLen(3))
	})
}

func FuzzMerge_Idempotent(f *testing.F) {
	f.Add("F1", "F2")

	f.Fuzz(func(t *testing.T, id1, id2 string) {
		g := NewWithT(t)
		if id1 == id2 {
			return // Skip identical IDs for this property
		}

		r := NewReport(ToolInfo{Name: "tool"})
		r.AddFinding(Finding{ID: id1})
		r.AddFinding(Finding{ID: id2})

		merged1 := Merge([]*Report{r})
		merged2 := Merge([]*Report{merged1})

		g.Expect(merged2.Findings).To(HaveLen(len(merged1.Findings)))
	})
}

func FuzzCorrelate(f *testing.F) {
	f.Add("tool1", "tool2", 1, 2, "file.go")
	f.Add("a", "b", 10, 20, "x.go")

	f.Fuzz(func(t *testing.T, tool1, tool2 string, line1, line2 int, file string) {
		g := NewWithT(t)
		findings := []Finding{
			{ID: "F1", ToolName: tool1, Position: Position{File: file, Line: line1}},
			{ID: "F2", ToolName: tool2, Position: Position{File: file, Line: line2}},
		}

		correlations := Correlate(findings)

		for _, c := range correlations {
			g.Expect(c.Confidence).To(BeNumerically(">=", 0.0))
			g.Expect(c.Confidence).To(BeNumerically("<=", 1.0))
			g.Expect(c.FindingIDs).To(HaveLen(2))
		}

		diffFile := []Finding{
			MakeFindingWithPos("F1", "", "a", "", SeverityInfo, "a.go", 1, 0),
			MakeFindingWithPos("F2", "", "b", "", SeverityInfo, "b.go", 1, 0),
		}
		g.Expect(Correlate(diffFile)).To(BeEmpty())
	})
}

func FuzzMergeByPosition(f *testing.F) {
	f.Add("file.go", 1, 1, "file.go", 1, 1)
	f.Add("file.go", 10, 5, "other.go", 10, 5)

	f.Fuzz(func(t *testing.T, file1 string, line1, col1 int, file2 string, line2, col2 int) {
		g := NewWithT(t)
		r1 := NewReport(ToolInfo{Name: "t1"})
		r1.AddFinding(Finding{
			ID:       "A",
			Position: Position{File: file1, Line: line1, Column: col1},
		})

		r2 := NewReport(ToolInfo{Name: "t2"})
		r2.AddFinding(Finding{
			ID:       "B",
			Position: Position{File: file2, Line: line2, Column: col2},
		})

		merged := Merge([]*Report{r1, r2}, WithDeduplicateBy(DeduplicateByPosition))
		key1 := fmt.Sprintf("%s:%d:%d", file1, line1, col1)
		key2 := fmt.Sprintf("%s:%d:%d", file2, line2, col2)

		if key1 == key2 {
			g.Expect(merged.Findings).To(HaveLen(1))
		}
	})
}

func FuzzDedupKey(f *testing.F) {
	f.Add("F1", "file.go", 1, 1, "rule1")
	f.Add("", "", 0, 0, "")

	f.Fuzz(func(t *testing.T, id, file string, line, col int, rule string) {
		g := NewWithT(t)
		f := Finding{
			ID:       id,
			Rule:     rule,
			Position: Position{File: file, Line: line, Column: col},
		}

		k1, _ := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByID})
		g.Expect(k1).To(ContainSubstring(id))

		k2, _ := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByPosition})
		g.Expect(k2).To(ContainSubstring(file))

		k3, _ := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByRule})
		g.Expect(k3).To(ContainSubstring(rule))
	})
}
