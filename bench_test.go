package finding

import (
	"context"
	"fmt"
	"testing"
)

const (
	benchRule     = "SA1000"
	benchFile     = "file.go"
	benchMainFile = "main.go"
	benchVersion  = "1.0"
)

func benchPosition(fileFmt string, mod, i int) Position {
	return Position{File: fmt.Sprintf(fileFmt, i%mod), Line: i + 1}
}

func benchReport(i int) *Report {
	return NewReport(ToolInfo{Name: fmt.Sprintf("tool%d", i)})
}

func BenchmarkClone(b *testing.B) {
	f := Finding{
		ID:       "tool:rule:file.go:42",
		Rule:     benchRule,
		Message:  "test finding with a longer message",
		Severity: SeverityError,
		Position: Position{File: benchFile, Line: 42, Column: 10},
		Range: &Range{
			Start: Position{Line: 42, Column: 10},
			End:   Position{Line: 42, Column: 20},
		},
		Related:  []RelatedRef{{FindingID: "related1", Relation: "related"}},
		Category: CategorySecurity,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		f.Clone()
	}
}

func BenchmarkClone_Simple(b *testing.B) {
	f := Finding{
		ID:       "tool:rule:file.go:1",
		Severity: SeverityInfo,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		f.Clone()
	}
}

func BenchmarkGenerateID(b *testing.B) {
	pos := Position{File: benchMainFile, Line: 42, Column: 10}

	b.ResetTimer()

	for b.Loop() {
		GenerateID("govet", "printf", pos)
	}
}

func BenchmarkGenerateID_Hash(b *testing.B) {
	pos := Position{File: benchMainFile}

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
			Position: Position{File: benchFile, Line: i + 1},
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
			Position: benchPosition("file%d.go", 10, i),
		}
	}

	b.ResetTimer()

	for b.Loop() {
		Filter(
			findings,
			BySeverityAtLeast(SeverityWarning),
			ByCategory(CategorySecurity),
		)
	}
}

func BenchmarkGroupByFile(b *testing.B) {
	findings := make([]Finding, 500)
	for i := range findings {
		findings[i] = Finding{
			Position: benchPosition("pkg/file%d.go", 20, i),
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
		reports[i] = benchReport(i)
		for j := range 200 {
			reports[i].AddFinding(Finding{
				ID:       fmt.Sprintf("tool%d:rule:file%d.go:%d", i, j%10, j),
				Severity: sevFromInt(j % 4),
				Position: benchPosition("file%d.go", 10, j),
			})
		}
	}

	b.ResetTimer()

	for b.Loop() {
		Combine(reports, WithDeduplication(true))
	}
}

func BenchmarkCorrelate(b *testing.B) {
	findings := make([]Finding, 200)
	for i := range findings {
		findings[i] = Finding{
			ID:       fmt.Sprintf("tool%d:rule:file.go:%d", i%5, i),
			ToolName: fmt.Sprintf("tool%d", i%5),
			Position: benchPosition("file%d.go", 5, i),
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
		reports[i] = benchReport(i)
		for j := range 500 {
			reports[i].AddFinding(Finding{
				ID:       fmt.Sprintf("tool%d:rule:file.go:%d:%d", i, j, i),
				Severity: sevFromInt(j % 4),
			})
		}
	}

	b.ResetTimer()

	for b.Loop() {
		Combine(reports, WithDeduplication(false))
	}
}

func sarifBenchReport() *Report {
	report := NewReport(ToolInfo{Name: "bench", Version: benchVersion})
	for i := range 100 {
		report.AddFinding(Finding{
			ID:       fmt.Sprintf("tool:rule:file.go:%d", i),
			Rule:     benchRule,
			Message:  "test finding",
			Severity: sevFromInt(i % 4),
			Position: Position{File: benchFile, Line: i + 1, Column: 1},
		})
	}

	return report
}

func BenchmarkToSARIF(b *testing.B) {
	report := sarifBenchReport()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, _ = report.ToSARIF()
	}
}

func BenchmarkFromSARIF(b *testing.B) {
	report := sarifBenchReport()

	sarifData, _ := report.ToSARIF()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, _ = FindingsFromSARIF(context.Background(), sarifData)
	}
}

func BenchmarkFindingKey(b *testing.B) {
	f := Finding{
		ID:       "tool:rule:file.go:42:10",
		Rule:     benchRule,
		Message:  "test finding",
		Position: Position{File: benchFile, Line: 42, Column: 10},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = f.Key()
	}
}

func BenchmarkFindingKey_NoID(b *testing.B) {
	f := Finding{
		Rule:     benchRule,
		Message:  "test finding",
		Position: Position{File: benchFile, Line: 42, Column: 10},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = f.Key()
	}
}
