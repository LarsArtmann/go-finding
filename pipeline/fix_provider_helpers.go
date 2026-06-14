package pipeline

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/larsartmann/go-finding"
)

var (
	// ErrPositionUnresolvable indicates a line/column could not be mapped to a byte offset.
	ErrPositionUnresolvable = errors.New("position unresolvable in content")

	errInvalidLine   = errors.New("invalid line number")
	errLineBeyondEOF = errors.New("line beyond end of file")
	errColumnBeyond  = errors.New("column beyond end of line")
)

func newReplacementEdit(offset, length int, f finding.Finding) FixEdit {
	return FixEdit{Offset: offset, Length: length, Replacement: []byte(f.AfterCode), Source: f}
}

// indexLineColToOffset converts a 1-based line and column to a 0-based byte
// offset using a pre-built line offset index for O(1) lookup.
func indexLineColToOffset(index []int, contentLen, line, col int) (int, error) {
	if line < 1 {
		return 0, fmt.Errorf("%w: %d (contentLen=%d)", errInvalidLine, line, contentLen)
	}

	if line > len(index) {
		return 0, fmt.Errorf("%w: %d (contentLen=%d)", errLineBeyondEOF, line, contentLen)
	}

	offset := index[line-1]
	if col > 1 {
		offset += col - 1
	}

	if offset > contentLen {
		return 0, fmt.Errorf(
			"%w: %d at line %d", errColumnBeyond, col, line,
		)
	}

	return offset, nil
}

// buildLineOffsetIndex returns a slice where index[i] is the byte offset of
// the start of line i+1 (1-based line number → 0-based slice index).
func buildLineOffsetIndex(content []byte) []int {
	// Count newlines via bytes.Count — SIMD-accelerated for single-byte needle.
	lineCount := bytes.Count(content, []byte{'\n'}) + 1

	index := make([]int, 0, lineCount)
	index = append(index, 0) // line 1 starts at offset 0

	for i, b := range content {
		if b == '\n' && i+1 < len(content) {
			index = append(index, i+1)
		}
	}

	return index
}

// defaultOccurrenceCapacity is the starting capacity for findAllOccurrences
// results. Chosen to avoid the first 5 reallocation growth phases (0→1→2→4→8→16→32).
const defaultOccurrenceCapacity = 32

// findAllOccurrences returns all starting byte positions of needle in haystack.
func findAllOccurrences(haystack, needle []byte) []int {
	if len(needle) == 0 {
		return nil
	}

	results := make([]int, 0, defaultOccurrenceCapacity)

	idx := 0

	for {
		i := bytes.Index(haystack[idx:], needle)
		if i < 0 {
			break
		}

		results = append(results, idx+i)
		idx += i + 1
	}

	return results
}

// offsetLineDistance returns the absolute difference between the line number
// containing the given byte offset and targetLine. Uses binary search on a
// pre-built line offset index for O(log n) lookup instead of the previous
// O(n) byte-by-byte newline count.
func offsetLineDistance(lineIndex []int, offset, targetLine int) int {
	line := offsetToLine(lineIndex, offset)

	diff := line - targetLine
	if diff < 0 {
		return -diff
	}

	return diff
}

// offsetToLine returns the 1-based line number containing the given byte offset.
// Uses binary search on the line offset index.
func offsetToLine(lineIndex []int, offset int) int {
	if len(lineIndex) == 0 {
		return 1
	}

	// Find the first line start strictly greater than offset.
	// The number of line starts at or before offset equals the 1-based line number.
	idx, _ := slices.BinarySearch(lineIndex, offset+1)
	if idx == 0 {
		return 1
	}

	return idx
}
