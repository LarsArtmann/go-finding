package pipeline

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
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

// findOldOccurrences returns byte offsets of all "old()" in content.
func findOldOccurrences(content []byte) []int {
	needle := []byte("old()")

	var positions []int

	start := 0

	for {
		i := bytes.Index(content[start:], needle)
		if i < 0 {
			break
		}

		positions = append(positions, start+i)
		start += i + len(needle)
	}

	return positions
}

// offsetToLineNumber returns the 1-based line number for a byte offset in content.
func offsetToLineNumber(content []byte, offset int) int {
	return bytes.Count(content[:offset], []byte{'\n'}) + 1
}

// columnOfOffset returns the 1-based column for a byte offset within its line.
func columnOfOffset(content []byte, offset int) int {
	col := 1

	for i := offset - 1; i >= 0; i-- {
		if content[i] == '\n' {
			break
		}

		col++
	}

	return col
}

// pickEvenly returns up to n evenly-spaced indices from positions.
func pickEvenly(positions []int, n int) []int {
	if len(positions) == 0 || n <= 0 {
		return nil
	}

	step := len(positions) / n
	if step == 0 {
		step = 1
	}

	var result []int

	for i := range n {
		idx := i * step
		if idx >= len(positions) {
			break
		}

		result = append(result, positions[idx])
	}

	return result
}

// generateOffsetFixes creates n non-overlapping offset-based fixes pointing
// to actual "old()" occurrences in the content. Uses OffsetProvider.
func generateOffsetFixes(n int, content []byte) []finding.Finding {
	positions := pickEvenly(findOldOccurrences(content), n)
	fixes := make([]finding.Finding, 0, len(positions))

	for _, pos := range positions {
		fixes = append(fixes, finding.Finding{
			BeforeCode: "old()",
			AfterCode:  "new()",
			Range: &finding.Range{
				Start: finding.Position{File: "bench.go", Offset: pos},
				End:   finding.Position{File: "bench.go", Offset: pos + 5},
			},
			Position: finding.Pos("bench.go", offsetToLineNumber(content, pos), 1),
		})
	}

	return fixes
}

// generateLineFixes creates n non-overlapping line/column-based fixes.
// No byte-offset Range — exercises LineProvider exclusively.
func generateLineFixes(n int, content []byte) []finding.Finding {
	positions := pickEvenly(findOldOccurrences(content), n)
	fixes := make([]finding.Finding, 0, len(positions))

	for _, pos := range positions {
		line := offsetToLineNumber(content, pos)
		col := columnOfOffset(content, pos)

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
	positions := pickEvenly(findOldOccurrences(content), n)
	fixes := make([]finding.Finding, 0, len(positions))

	for _, pos := range positions {
		fixes = append(fixes, finding.Finding{
			BeforeCode: "old()",
			AfterCode:  "new()",
			Position:   finding.Pos("bench.go", offsetToLineNumber(content, pos), 1),
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
