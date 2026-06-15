package goast

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

// generateGoContent produces valid Go source with n lines, each containing old().
func generateGoContent(lines int) []byte {
	var b bytes.Buffer

	fmt.Fprintf(&b, "package main\n\nfunc main() {\n")

	for range lines {
		fmt.Fprintf(&b, "\told()\n")
	}

	fmt.Fprintf(&b, "}\n")

	return b.Bytes()
}

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

func offsetToLineNumber(content []byte, offset int) int {
	return bytes.Count(content[:offset], []byte{'\n'}) + 1
}

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

func generateGoASTFixes(n int, content []byte) []finding.Finding {
	occurrences := findOldOccurrences(content)
	selected := pickEvenly(occurrences, n)
	fixes := make([]finding.Finding, 0, len(selected))

	for _, pos := range selected {
		line := offsetToLineNumber(content, pos)
		col := columnOfOffset(content, pos)

		fixes = append(fixes, finding.Finding{
			Position:   finding.Position{File: "bench.go", Line: line, Column: col},
			BeforeCode: "old()",
			AfterCode:  "new()",
		})
	}

	return fixes
}

func benchmarkGoASTProvider(b *testing.B, fixCount int) {
	b.Helper()

	content := generateGoContent(10000)
	fixes := generateGoASTFixes(fixCount, content)
	engine := pipeline.NewFixEngineWithProviders(
		&Provider{},
		&pipeline.SubstringProvider{},
	)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		engine.Apply(content, fixes)
	}
}

func BenchmarkGoAST_1(b *testing.B)    { benchmarkGoASTProvider(b, 1) }
func BenchmarkGoAST_10(b *testing.B)   { benchmarkGoASTProvider(b, 10) }
func BenchmarkGoAST_100(b *testing.B)  { benchmarkGoASTProvider(b, 100) }
func BenchmarkGoAST_1000(b *testing.B) { benchmarkGoASTProvider(b, 1000) }

// BenchmarkGoAST_vs_Substring compares GoAST vs SubstringProvider on the same
// valid Go source. GoAST uses AST structure for disambiguation; Substring scans
// all occurrences with line-proximity binary search.
func BenchmarkGoAST_vs_Substring(b *testing.B) {
	content := generateGoContent(10000)
	fixes := generateGoASTFixes(100, content)

	b.Run("GoAST", func(b *testing.B) {
		engine := pipeline.NewFixEngineWithProviders(&Provider{})

		b.ResetTimer()
		b.ReportAllocs()

		for range b.N {
			engine.Apply(content, fixes)
		}
	})

	b.Run("Substring", func(b *testing.B) {
		engine := pipeline.NewFixEngine()

		b.ResetTimer()
		b.ReportAllocs()

		for range b.N {
			engine.Apply(content, fixes)
		}
	})
}
