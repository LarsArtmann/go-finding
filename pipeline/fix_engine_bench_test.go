package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline/internal/benchutil"
)

// generateContent creates n lines of Go-like code for benchmarking.
// Each line contains "old()" at a known position for consistent testing.
func generateContent(lines int) []byte {
	var b strings.Builder
	b.Grow(lines * 30)

	for i := 1; i <= lines; i++ {
		fmt.Fprintf(&b, "line%d: old()\n", i)
	}

	return []byte(b.String())
}

// generateOffsetFixes creates n non-overlapping offset-based fixes pointing
// to actual "old()" occurrences in the content. Uses OffsetProvider.
func generateOffsetFixes(n int, content []byte) []finding.Finding {
	positions := benchutil.PickEvenly(benchutil.FindOldOccurrences(content), n)
	fixes := make([]finding.Finding, 0, len(positions))

	for _, pos := range positions {
		fixes = append(fixes, finding.Finding{
			BeforeCode: "old()",
			AfterCode:  "new()",
			Range: &finding.Range{
				Start: finding.Position{File: "bench.go", Offset: pos},
				End:   finding.Position{File: "bench.go", Offset: pos + 5},
			},
			Position: finding.Pos("bench.go", benchutil.OffsetToLineNumber(content, pos), 1),
		})
	}

	return fixes
}

// generateLineFixes creates n non-overlapping line/column-based fixes.
// No byte-offset Range — exercises LineProvider exclusively.
func generateLineFixes(n int, content []byte) []finding.Finding {
	positions := benchutil.PickEvenly(benchutil.FindOldOccurrences(content), n)
	fixes := make([]finding.Finding, 0, len(positions))

	for _, pos := range positions {
		line := benchutil.OffsetToLineNumber(content, pos)
		col := benchutil.ColumnOfOffset(content, pos)

		fixes = append(fixes, finding.Finding{
			BeforeCode: "old()",
			AfterCode:  "new()",
			Position:   finding.Pos("bench.go", line, col),
		})
	}

	return fixes
}

// generateSubstringFixes creates n fixes with BeforeCode but line=1 (mismatched).
// Exercises SubstringProvider with occurrence disambiguation.
func generateSubstringFixes(n int, content []byte) []finding.Finding {
	positions := benchutil.PickEvenly(benchutil.FindOldOccurrences(content), n)
	fixes := make([]finding.Finding, 0, len(positions))

	for _, pos := range positions {
		fixes = append(fixes, finding.Finding{
			BeforeCode: "old()",
			AfterCode:  "new()",
			Position:   finding.Pos("bench.go", benchutil.OffsetToLineNumber(content, pos), 1),
		})
	}

	return fixes
}

// OffsetProvider benchmarks.

func benchmarkFixEngineApply(b *testing.B, fixCount int) {
	b.Helper()

	content := generateContent(10000)
	fixes := generateOffsetFixes(fixCount, content)
	engine := NewFixEngine()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		engine.Apply(content, fixes)
	}
}

func BenchmarkFixEngine_Apply_1(b *testing.B)    { benchmarkFixEngineApply(b, 1) }
func BenchmarkFixEngine_Apply_10(b *testing.B)   { benchmarkFixEngineApply(b, 10) }
func BenchmarkFixEngine_Apply_100(b *testing.B)  { benchmarkFixEngineApply(b, 100) }
func BenchmarkFixEngine_Apply_1000(b *testing.B) { benchmarkFixEngineApply(b, 1000) }

// LineProvider benchmarks.

func benchmarkLineProvider(b *testing.B, fixCount int) {
	b.Helper()

	content := generateContent(10000)
	fixes := generateLineFixes(fixCount, content)
	engine := NewFixEngine()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		engine.Apply(content, fixes)
	}
}

func BenchmarkFixEngine_LineProvider_1(b *testing.B)    { benchmarkLineProvider(b, 1) }
func BenchmarkFixEngine_LineProvider_10(b *testing.B)   { benchmarkLineProvider(b, 10) }
func BenchmarkFixEngine_LineProvider_100(b *testing.B)  { benchmarkLineProvider(b, 100) }
func BenchmarkFixEngine_LineProvider_1000(b *testing.B) { benchmarkLineProvider(b, 1000) }

// SubstringProvider benchmarks.

func benchmarkSubstringProvider(b *testing.B, fixCount int) {
	b.Helper()

	content := generateContent(10000)
	fixes := generateSubstringFixes(fixCount, content)
	engine := NewFixEngine()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		engine.Apply(content, fixes)
	}
}

func BenchmarkFixEngine_Substring_1(b *testing.B)    { benchmarkSubstringProvider(b, 1) }
func BenchmarkFixEngine_Substring_10(b *testing.B)   { benchmarkSubstringProvider(b, 10) }
func BenchmarkFixEngine_Substring_100(b *testing.B)  { benchmarkSubstringProvider(b, 100) }
func BenchmarkFixEngine_Substring_1000(b *testing.B) { benchmarkSubstringProvider(b, 1000) }

// ApplyWithOutcomes benchmarks — the canonical entry point since the outcome
// rework (Apply/ApplyWithConflicts delegate here). Mirrors the offset and line
// provider variants above so benchstat can attribute outcome-bookkeeping cost.

func benchmarkApplyWithOutcomes(b *testing.B, fixes []finding.Finding) {
	b.Helper()

	content := generateContent(10000)
	engine := NewFixEngine()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		engine.ApplyWithOutcomes(content, fixes)
	}
}

func benchmarkOutcomesOffset(b *testing.B, fixCount int) {
	b.Helper()

	content := generateContent(10000)
	benchmarkApplyWithOutcomes(b, generateOffsetFixes(fixCount, content))
}

func benchmarkOutcomesLine(b *testing.B, fixCount int) {
	b.Helper()

	content := generateContent(10000)
	benchmarkApplyWithOutcomes(b, generateLineFixes(fixCount, content))
}

func BenchmarkFixEngine_ApplyWithOutcomes_Offset_1(b *testing.B) {
	benchmarkOutcomesOffset(b, 1)
}

func BenchmarkFixEngine_ApplyWithOutcomes_Offset_10(b *testing.B) {
	benchmarkOutcomesOffset(b, 10)
}

func BenchmarkFixEngine_ApplyWithOutcomes_Offset_100(b *testing.B) {
	benchmarkOutcomesOffset(b, 100)
}

func BenchmarkFixEngine_ApplyWithOutcomes_Offset_1000(b *testing.B) {
	benchmarkOutcomesOffset(b, 1000)
}

func BenchmarkFixEngine_ApplyWithOutcomes_Line_1(b *testing.B) {
	benchmarkOutcomesLine(b, 1)
}

func BenchmarkFixEngine_ApplyWithOutcomes_Line_10(b *testing.B) {
	benchmarkOutcomesLine(b, 10)
}

func BenchmarkFixEngine_ApplyWithOutcomes_Line_100(b *testing.B) {
	benchmarkOutcomesLine(b, 100)
}

func BenchmarkFixEngine_ApplyWithOutcomes_Line_1000(b *testing.B) {
	benchmarkOutcomesLine(b, 1000)
}

// generateMixedOutcomeFixes creates n fixes whose outcome statuses are mixed:
// every 4th finding is a refusal (BeforeCode absent from content), every 4th
// is a no-change (no code change), and the rest apply cleanly. This exercises
// the reconciliation pass across statuses, not just the all-applied path.
func generateMixedOutcomeFixes(n int, content []byte) []finding.Finding {
	positions := benchutil.PickEvenly(benchutil.FindOldOccurrences(content), n)
	fixes := make([]finding.Finding, 0, len(positions))

	for i, pos := range positions {
		switch {
		case i%4 == 1:
			fixes = append(fixes, finding.Finding{
				BeforeCode: "definitely-not-present()",
				AfterCode:  "new()",
				Range: &finding.Range{
					Start: finding.Position{File: "bench.go", Offset: pos},
					End:   finding.Position{File: "bench.go", Offset: pos + 5},
				},
				Position: finding.Pos("bench.go", benchutil.OffsetToLineNumber(content, pos), 1),
			})
		case i%4 == 2:
			fixes = append(fixes, finding.Finding{
				Message:  "report only",
				Position: finding.Pos("bench.go", benchutil.OffsetToLineNumber(content, pos), 1),
			})
		default:
			fixes = append(fixes, finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Range: &finding.Range{
					Start: finding.Position{File: "bench.go", Offset: pos},
					End:   finding.Position{File: "bench.go", Offset: pos + 5},
				},
				Position: finding.Pos("bench.go", benchutil.OffsetToLineNumber(content, pos), 1),
			})
		}
	}

	return fixes
}

// BenchmarkFixEngine_ApplyWithOutcomes_Mixed_1000 measures the outcome
// bookkeeping when statuses are mixed (applied / refused / no-change) at
// n=1000, confirming the reconciliation pass stays allocation-lean relative
// to the all-applied variants.
func BenchmarkFixEngine_ApplyWithOutcomes_Mixed_1000(b *testing.B) {
	b.Helper()

	content := generateContent(10000)
	benchmarkApplyWithOutcomes(b, generateMixedOutcomeFixes(1000, content))
}

// benchmarkFixApplierRun measures ApplyWithReport (writes + backups) vs
// ApplyDryRun (read-only) on the same fix set, quantifying the plan-mode cost.
func benchmarkFixApplierRun(b *testing.B, dryRun bool) {
	b.Helper()

	const lines = 200

	content := generateContent(lines)

	dir := b.TempDir()
	file := filepath.Join(dir, "bench.go")
	if err := os.WriteFile(file, content, 0o600); err != nil {
		b.Fatal(err)
	}

	fixes := generateOffsetFixes(50, content)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		applier, err := NewFixApplier(dir)
		if err != nil {
			b.Fatal(err)
		}

		if dryRun {
			if _, err := applier.ApplyDryRun(context.Background(), fixes); err != nil {
				b.Fatal(err)
			}
		} else if _, err := applier.ApplyWithReport(context.Background(), fixes); err != nil {
			b.Fatal(err)
		}

		_ = applier.Close()

		// Restore original content so iterations are identical.
		if err := os.WriteFile(file, content, 0o600); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFixApplier_ApplyWithReport_50(b *testing.B) { benchmarkFixApplierRun(b, false) }
func BenchmarkFixApplier_ApplyDryRun_50(b *testing.B)     { benchmarkFixApplierRun(b, true) }
