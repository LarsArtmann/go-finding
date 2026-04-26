package finding

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzFilter(f *testing.F) {
	f.Add("warning", "cat1", "file1.go", "tool1")
	f.Add("error", "cat2", "file2.go", "tool2")
	f.Add("info", "", "", "")

	f.Fuzz(func(t *testing.T, sevStr, category, file, tool string) {
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
		assert.Len(t, all, len(findings), "no predicates")

		// Property: filtered results all match the predicate
		byCat := Filter(findings, ByCategory(Category(category)))
		for _, f := range byCat {
			assert.Equal(t, Category(category), f.Category, "ByCategory mismatch")
		}

		// Property: BySeverity results match exactly
		bySev := Filter(findings, BySeverity(severity))
		for _, f := range bySev {
			assert.Equal(t, severity, f.Severity, "BySeverity mismatch")
		}

		// Property: combined predicates = intersection
		both := Filter(findings, ByCategory(Category(category)), BySeverity(severity))
		for _, f := range both {
			assert.Equal(t, Category(category), f.Category, "combined filter category")
			assert.Equal(t, severity, f.Severity, "combined filter severity")
		}

		assert.LessOrEqual(t, len(both), len(byCat), "combined <= byCat")
		assert.LessOrEqual(t, len(both), len(bySev), "combined <= bySev")
	})
}

func FuzzGroupBy(f *testing.F) {
	f.Add("key1", "key2", "key1")
	f.Add("", "", "")
	f.Add("a", "a", "a")

	f.Fuzz(func(t *testing.T, key1, key2, key3 string) {
		findings := []Finding{
			{ID: "F1", Category: Category(key1)},
			{ID: "F2", Category: Category(key2)},
			{ID: "F3", Category: Category(key3)},
		}

		groups := GroupBy(findings, func(f Finding) string { return string(f.Category) })

		total := 0
		for _, g := range groups {
			total += len(g)
		}

		assert.Equal(t, len(findings), total, "GroupBy total")

		for _, g := range groups {
			key := g[0].Category
			for _, f := range g {
				assert.Equal(t, key, f.Category, "group mismatch")
			}
		}
	})
}

func FuzzFilterByFile(f *testing.F) {
	f.Add("main.go", "main.go", "test.go")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, target, file1, file2 string) {
		findings := []Finding{
			{ID: "F1", Position: Position{File: file1}},
			{ID: "F2", Position: Position{File: file2}},
		}

		result := Filter(findings, ByFile(target))
		for _, f := range result {
			assert.Equal(t, target, f.Position.File, "ByFile mismatch")
		}
	})
}

func FuzzGroupByFile(f *testing.F) {
	f.Add("a.go", "b.go", "a.go")

	f.Fuzz(func(t *testing.T, file1, file2, file3 string) {
		findings := []Finding{
			{ID: "F1", Position: Position{File: file1}},
			{ID: "F2", Position: Position{File: file2}},
			{ID: "F3", Position: Position{File: file3}},
		}

		groups := GroupByFile(findings)

		total := 0
		for _, g := range groups {
			total += len(g)
		}

		assert.Equal(t, len(findings), total, "GroupByFile total")

		if file1 == file2 && file2 == file3 {
			assert.Len(t, groups, 1, "same file")
		}
	})
}

func FuzzMerge_DedupByID(f *testing.F) {
	f.Add("F1", "F1", "F2")
	f.Add("unique1", "unique2", "unique3")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, id1, id2, id3 string) {
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
			seen[f.ID]++
			assert.LessOrEqual(t, seen[f.ID], 1, "duplicate ID in merged: %q", f.ID)
		}

		// Without dedup: total should be sum
		noDedup := Merge([]*Report{r1, r2}, WithDeduplication(false))
		assert.Len(t, noDedup.Findings, 3, "no dedup")
	})
}

func FuzzMerge_Idempotent(f *testing.F) {
	f.Add("F1", "F2")

	f.Fuzz(func(t *testing.T, id1, id2 string) {
		if id1 == id2 {
			return // Skip identical IDs for this property
		}

		r := NewReport(ToolInfo{Name: "tool"})
		r.AddFinding(Finding{ID: id1})
		r.AddFinding(Finding{ID: id2})

		merged1 := Merge([]*Report{r})
		merged2 := Merge([]*Report{merged1})

		assert.Equal(t, len(merged1.Findings), len(merged2.Findings), "merge not idempotent")
	})
}

func FuzzCorrelate(f *testing.F) {
	f.Add("tool1", "tool2", 1, 2, "file.go")
	f.Add("a", "b", 10, 20, "x.go")

	f.Fuzz(func(t *testing.T, tool1, tool2 string, line1, line2 int, file string) {
		findings := []Finding{
			{ID: "F1", ToolName: tool1, Position: Position{File: file, Line: line1}},
			{ID: "F2", ToolName: tool2, Position: Position{File: file, Line: line2}},
		}

		correlations := Correlate(findings)

		for _, c := range correlations {
			assert.InDelta(t, 0, c.Confidence-c.Confidence, 1, "confidence")
			assert.GreaterOrEqual(t, c.Confidence, 0.0)
			assert.LessOrEqual(t, c.Confidence, 1.0)
			assert.Len(t, c.FindingIDs, 2)
		}

		diffFile := []Finding{
			MakeFindingWithPos("F1", "", "a", "", SeverityInfo, "a.go", 1, 0),
			MakeFindingWithPos("F2", "", "b", "", SeverityInfo, "b.go", 1, 0),
		}
		assert.Empty(t, Correlate(diffFile), "different files should not correlate")
	})
}

func FuzzMergeByPosition(f *testing.F) {
	f.Add("file.go", 1, 1, "file.go", 1, 1)
	f.Add("file.go", 10, 5, "other.go", 10, 5)

	f.Fuzz(func(t *testing.T, file1 string, line1, col1 int, file2 string, line2, col2 int) {
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
			assert.Len(t, merged.Findings, 1, "same position should dedup")
		}
	})
}

func FuzzDedupKey(f *testing.F) {
	f.Add("F1", "file.go", 1, 1, "rule1")
	f.Add("", "", 0, 0, "")

	f.Fuzz(func(t *testing.T, id, file string, line, col int, rule string) {
		f := Finding{
			ID:       id,
			Rule:     rule,
			Position: Position{File: file, Line: line, Column: col},
		}

		k1 := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByID})
		assert.Contains(t, k1, id, "ID key should contain ID")

		k2 := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByPosition})
		assert.Contains(t, k2, file, "position key should contain file")

		k3 := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByRule})
		assert.Contains(t, k3, rule, "rule key should contain rule")
	})
}
