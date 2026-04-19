package finding

import (
	"fmt"
	"strings"
	"testing"
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

		// Property: no predicates returns all
		all := Filter(findings)
		if len(all) != len(findings) {
			t.Errorf("no predicates: expected %d, got %d", len(findings), len(all))
		}

		// Property: filtered results all match the predicate
		byCat := Filter(findings, ByCategory(Category(category)))
		for _, f := range byCat {
			if f.Category != Category(category) {
				t.Errorf("ByCategory mismatch: got %q, want %q", f.Category, category)
			}
		}

		// Property: BySeverity results match exactly
		bySev := Filter(findings, BySeverity(severity))
		for _, f := range bySev {
			if f.Severity != severity {
				t.Errorf("BySeverity mismatch: got %v, want %v", f.Severity, severity)
			}
		}

		// Property: combined predicates = intersection
		both := Filter(findings, ByCategory(Category(category)), BySeverity(severity))
		for _, f := range both {
			if f.Category != Category(category) || f.Severity != severity {
				t.Errorf("combined filter mismatch")
			}
		}

		if len(both) > len(byCat) || len(both) > len(bySev) {
			t.Error("combined filter should be subset of each individual filter")
		}
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

		if total != len(findings) {
			t.Errorf("GroupBy total: got %d, want %d", total, len(findings))
		}

		for _, g := range groups {
			key := g[0].Category
			for _, f := range g {
				if f.Category != key {
					t.Errorf("group mismatch: %q != %q", f.Category, key)
				}
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
			if f.Position.File != target {
				t.Errorf("ByFile mismatch: got %q, want %q", f.Position.File, target)
			}
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

		if total != len(findings) {
			t.Errorf("GroupByFile total: got %d, want %d", total, len(findings))
		}

		if file1 == file2 && file2 == file3 {
			if len(groups) != 1 {
				t.Errorf("same file: expected 1 group, got %d", len(groups))
			}
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
			if seen[f.ID] > 1 {
				t.Errorf("duplicate ID in merged: %q", f.ID)
			}
		}

		// Without dedup: total should be sum
		noDedup := Merge([]*Report{r1, r2}, WithDeduplication(false))
		if len(noDedup.Findings) != 3 {
			t.Errorf("no dedup: expected 3 findings, got %d", len(noDedup.Findings))
		}
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

		if len(merged1.Findings) != len(merged2.Findings) {
			t.Errorf("merge not idempotent: %d vs %d", len(merged1.Findings), len(merged2.Findings))
		}
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
			conf := c.Confidence
			if conf < 0 || conf > 1 {
				t.Errorf("confidence out of range: %f", conf)
			}

			if len(c.FindingIDs) != 2 {
				t.Errorf("expected 2 finding IDs, got %d", len(c.FindingIDs))
			}
		}

		// Different tools in different files should not correlate
		diffFile := []Finding{
			{ID: "F1", ToolName: "a", Position: Position{File: "a.go", Line: 1}},
			{ID: "F2", ToolName: "b", Position: Position{File: "b.go", Line: 1}},
		}
		if len(Correlate(diffFile)) != 0 {
			t.Error("different files should not correlate")
		}
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

		if key1 == key2 && len(merged.Findings) != 1 {
			t.Errorf("same position should dedup: got %d findings", len(merged.Findings))
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
		if !strings.Contains(k1, id) {
			t.Error("ID key should contain ID")
		}

		k2 := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByPosition})
		if !strings.Contains(k2, file) {
			t.Error("position key should contain file")
		}

		k3 := dedupKey(f, MergeOptions{DeduplicateBy: DeduplicateByRule})
		if !strings.Contains(k3, rule) {
			t.Error("rule key should contain rule")
		}
	})
}
