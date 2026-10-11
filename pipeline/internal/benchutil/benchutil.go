// Package benchutil provides shared helpers for benchmark tests that generate
// synthetic Go source with known "old()" occurrences and resolve them to
// line/column positions. Used by pipeline and pipeline/goast bench tests.
package benchutil

import "bytes"

// FindOldOccurrences returns byte offsets of all "old()" in content.
func FindOldOccurrences(content []byte) []int {
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

// OffsetToLineNumber returns the 1-based line number for a byte offset in content.
func OffsetToLineNumber(content []byte, offset int) int {
	return bytes.Count(content[:offset], []byte{'\n'}) + 1
}

// ColumnOfOffset returns the 1-based column for a byte offset within its line.
func ColumnOfOffset(content []byte, offset int) int {
	col := 1

	for i := offset - 1; i >= 0; i-- {
		if content[i] == '\n' {
			break
		}

		col++
	}

	return col
}

// PickEvenly returns up to n evenly-spaced indices from positions.
func PickEvenly(positions []int, n int) []int {
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
