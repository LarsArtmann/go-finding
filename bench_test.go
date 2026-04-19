package finding

import (
	"fmt"
	"testing"
)

func BenchmarkGenerateID(b *testing.B) {
	pos := Position{File: "main.go", Line: 42, Column: 10}

	b.ResetTimer()

	for b.Loop() {
		GenerateID("govet", "printf", pos)
	}
}

func BenchmarkGenerateID_Hash(b *testing.B) {
	pos := Position{File: "main.go"}

	b.ResetTimer()

	for b.Loop() {
		GenerateID("govet", "printf", pos)
	}
}

func BenchmarkParseID(b *testing.B) {
	ids := []string{
		"govet:printf:main.go:42:10",
		"staticcheck:SA1000:pkg/handler.go:100:5",
		"govet:copylocks:a:b:c:deep/path.go:999:1",
	}

	b.ResetTimer()

	for b.Loop() {
		for _, id := range ids {
			ParseID(id)
		}
	}
}

func BenchmarkFilter(b *testing.B) {
	findings := make([]Finding, 1000)
	for i := range findings {
		findings[i] = Finding{
			ID:       fmt.Sprintf("tool:rule:file.go:%d", i),
			Severity: sevFromInt(i % 4),
			Category: CategoryStyle,
			Position: Position{File: "file.go", Line: i + 1},
		}
	}

	b.ResetTimer()

	for b.Loop() {
		Filter(findings, BySeverityAtLeast(SeverityWarning))
	}
}

func BenchmarkFilterMultiple(b *testing.B) {
	findings := make([]Finding, 1000)
	for i := range findings {
		findings[i] = Finding{
			ID:       fmt.Sprintf("tool:rule:file.go:%d", i),
			Severity: sevFromInt(i % 4),
			Category: []Category{CategoryStyle, CategorySecurity, CategoryPerformance, CategoryCorrectness}[i%4],
			Position: Position{File: fmt.Sprintf("file%d.go", i%10), Line: i + 1},
		}
	}

	b.ResetTimer()

	for b.Loop() {
		Filter(findings,
			BySeverityAtLeast(SeverityWarning),
			ByCategory(CategorySecurity),
		)
	}
}

func BenchmarkGroupByFile(b *testing.B) {
	findings := make([]Finding, 500)
	for i := range findings {
		findings[i] = Finding{
			Position: Position{File: fmt.Sprintf("pkg/file%d.go", i%20), Line: i + 1},
		}
	}

	b.ResetTimer()

	for b.Loop() {
		GroupByFile(findings)
	}
}

func BenchmarkMerge(b *testing.B) {
	reports := make([]*Report, 5)
	for i := range reports {
		reports[i] = NewReport(ToolInfo{Name: fmt.Sprintf("tool%d", i)})
		for j := range 200 {
			reports[i].AddFinding(Finding{
				ID:       fmt.Sprintf("tool%d:rule:file%d.go:%d", i, j%10, j),
				Severity: sevFromInt(j % 4),
				Position: Position{File: fmt.Sprintf("file%d.go", j%10), Line: j + 1},
			})
		}
	}

	b.ResetTimer()

	for b.Loop() {
		Merge(reports, WithDeduplication(true))
	}
}

func BenchmarkCorrelate(b *testing.B) {
	findings := make([]Finding, 200)
	for i := range findings {
		findings[i] = Finding{
			ID:       fmt.Sprintf("tool%d:rule:file.go:%d", i%5, i),
			ToolName: fmt.Sprintf("tool%d", i%5),
			Position: Position{File: fmt.Sprintf("file%d.go", i%5), Line: i + 1},
		}
	}

	b.ResetTimer()

	for b.Loop() {
		Correlate(findings)
	}
}

func BenchmarkMergeNoDedup(b *testing.B) {
	reports := make([]*Report, 3)
	for i := range reports {
		reports[i] = NewReport(ToolInfo{Name: fmt.Sprintf("tool%d", i)})
		for j := range 500 {
			reports[i].AddFinding(Finding{
				ID:       fmt.Sprintf("tool%d:rule:file.go:%d:%d", i, j, i),
				Severity: sevFromInt(j % 4),
			})
		}
	}

	b.ResetTimer()

	for b.Loop() {
		Merge(reports, WithDeduplication(false))
	}
}

func BenchmarkToSARIF(b *testing.B) {
	report := NewReport(ToolInfo{Name: "bench", Version: "1.0"})
	for i := range 100 {
		report.AddFinding(Finding{
			ID:       fmt.Sprintf("tool:rule:file.go:%d", i),
			Rule:     "SA1000",
			Message:  "test finding",
			Severity: sevFromInt(i % 4),
			Position: Position{File: "file.go", Line: i + 1, Column: 1},
		})
	}

	b.ResetTimer()

	for b.Loop() {
		_, _ = report.ToSARIF()
	}
}

func BenchmarkFromSARIF(b *testing.B) {
	report := NewReport(ToolInfo{Name: "bench", Version: "1.0"})
	for i := range 100 {
		report.AddFinding(Finding{
			ID:       fmt.Sprintf("tool:rule:file.go:%d", i),
			Rule:     "SA1000",
			Message:  "test finding",
			Severity: sevFromInt(i % 4),
			Position: Position{File: "file.go", Line: i + 1, Column: 1},
		})
	}

	sarifData, _ := report.ToSARIF()

	b.ResetTimer()

	for b.Loop() {
		_, _ = FindingsFromSARIF(sarifData)
	}
}
