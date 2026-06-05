package pipeline

import (
	"fmt"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
)

// generateContent creates n lines of Go-like code for benchmarking.
func generateContent(lines int) []byte {
	var b strings.Builder
	b.Grow(lines * 30)

	for i := 1; i <= lines; i++ {
		fmt.Fprintf(&b, "line%d: old()\n", i)
	}

	return []byte(b.String())
}

// generateOffsetFixes creates n non-overlapping offset-based fixes on content.
func generateOffsetFixes(n int, content []byte) []finding.Finding {
	fixes := make([]finding.Finding, 0, n)
	step := len(content) / (n + 1)

	for i := range n {
		offset := (i + 1) * step
		if offset+4 > len(content) {
			break
		}

		fixes = append(fixes, finding.Finding{
			BeforeCode: "old()",
			AfterCode:  "new()",
			Range: &finding.Range{
				Start: finding.Position{File: "bench.go", Offset: offset},
				End:   finding.Position{File: "bench.go", Offset: offset + 4},
			},
			Position: finding.Pos("bench.go", 1, 1),
		})
	}

	return fixes
}

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
