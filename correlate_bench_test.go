package finding

import (
	"fmt"
	"testing"
)

// BenchmarkCorrelate_RangeBased benchmarks Correlate with range-based findings
// that use IntervalIndex for overlap queries.
func BenchmarkCorrelate_RangeBased(b *testing.B) {
	benchmarks := []struct {
		name     string
		findings int
	}{
		{"10", 10},
		{"100", 100},
		{"1000", 1000},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			findings := generateRangeFindings(bench.findings)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				_ = Correlate(findings)
			}
		})
	}
}

// BenchmarkCorrelate_PointBased benchmarks Correlate with point-based findings
// (no Range) that use line-proximity heuristics.
func BenchmarkCorrelate_PointBased(b *testing.B) {
	benchmarks := []struct {
		name     string
		findings int
	}{
		{"10", 10},
		{"100", 100},
		{"1000", 1000},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			findings := generatePointFindings(bench.findings)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				_ = Correlate(findings)
			}
		})
	}
}

// BenchmarkMergeIter_vs_Combine compares streaming MergeIter vs materializing Combine.
func BenchmarkMergeIter_vs_Combine(b *testing.B) {
	benchmarks := []struct {
		name        string
		reports     int
		perReport   int
		deduplicate bool
	}{
		{"5_reports_100_each_dedup", 5, 100, true},
		{"10_reports_100_each_dedup", 10, 100, true},
		{"10_reports_1000_each_dedup", 10, 1000, true},
		{"10_reports_100_each_nodedup", 10, 100, false},
	}

	for _, bench := range benchmarks {
		reports := makeReports(bench.reports, bench.perReport)

		b.Run(bench.name+"/Combine", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				_ = Combine(reports, WithDeduplication(bench.deduplicate))
			}
		})

		b.Run(bench.name+"/MergeIter", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				for range MergeIter(reports, WithDeduplication(bench.deduplicate)) {
				}
			}
		})
	}
}

// BenchmarkIntervalIndex_Query benchmarks overlap queries on IntervalIndex.
func BenchmarkIntervalIndex_Query(b *testing.B) {
	benchmarks := []struct {
		name      string
		intervals int
		queries   int
	}{
		{"100_intervals_10_queries", 100, 10},
		{"1000_intervals_100_queries", 1000, 100},
		{"10000_intervals_100_queries", 10000, 100},
	}

	for _, bench := range benchmarks {
		intervals := generateIntervals(bench.intervals)
		idx := NewIntervalIndex(intervals)
		queries := generateQueries(bench.queries, bench.intervals)

		b.Run(bench.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				for _, q := range queries {
					_ = idx.Query(q.start, q.end)
				}
			}
		})
	}
}

// BenchmarkIntervalIndex_Build benchmarks the construction of IntervalIndex.
func BenchmarkIntervalIndex_Build(b *testing.B) {
	benchmarks := []struct {
		name      string
		intervals int
	}{
		{"100", 100},
		{"1000", 1000},
		{"10000", 10000},
	}

	for _, bench := range benchmarks {
		intervals := generateIntervals(bench.intervals)

		b.Run(bench.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				_ = NewIntervalIndex(intervals)
			}
		})
	}
}

func generateRangeFindings(n int) []Finding {
	findings := make([]Finding, n)
	tools := []string{"tool-a", "tool-b", "tool-c"}

	for i := range n {
		startLine := (i % 100) + 1
		endLine := startLine + (i % 10) + 1

		findings[i] = Finding{
			ID:       ID(fmt.Sprintf("f%d", i)),
			Rule:     "test-rule",
			ToolName: ToolName(tools[i%len(tools)]),
			Position: Position{File: "test.go", Line: startLine},
			Range: &Range{
				Start: Position{File: "test.go", Line: startLine, Column: 1},
				End:   Position{File: "test.go", Line: endLine, Column: 1},
			},
		}
	}

	return findings
}

func generatePointFindings(n int) []Finding {
	findings := make([]Finding, n)
	tools := []string{"tool-a", "tool-b", "tool-c"}

	for i := range n {
		findings[i] = Finding{
			ID:       ID(fmt.Sprintf("f%d", i)),
			Rule:     "test-rule",
			ToolName: ToolName(tools[i%len(tools)]),
			Position: Position{File: "test.go", Line: (i % 100) + 1, Column: 1},
		}
	}

	return findings
}

func makeReports(count, perReport int) []*Report {
	reports := make([]*Report, count)

	for i := range count {
		r := NewReport(ToolInfo{Name: fmt.Sprintf("tool-%d", i)})
		findings := make([]Finding, perReport)

		for j := range perReport {
			findings[j] = Finding{
				ID:       ID(fmt.Sprintf("tool-%d:f%d", i, j)),
				Rule:     "rule",
				ToolName: ToolName(fmt.Sprintf("tool-%d", i)),
				Position: Position{File: "test.go", Line: j + 1, Column: 1},
			}
		}

		r.AddFindings(findings)
		reports[i] = r
	}

	return reports
}

func generateIntervals(n int) []Interval[int] {
	intervals := make([]Interval[int], n)
	for i := range n {
		start := (i * 10) % (n * 5)
		intervals[i] = Interval[int]{
			Start: start,
			End:   start + 5,
			Value: i,
		}
	}

	return intervals
}

type queryRange struct {
	start int
	end   int
}

func generateQueries(n, maxLine int) []queryRange {
	queries := make([]queryRange, n)
	for i := range n {
		start := (i * 7) % (maxLine * 5)
		queries[i] = queryRange{start: start, end: start + 3}
	}

	return queries
}
