package finding

import (
	"cmp"
	"fmt"
)

// Position represents a location in source code.
// Line and Column are 1-based; 0 means not set.
type Position struct {
	File   string `json:"file"`             // Required: file path
	Line   int    `json:"line,omitempty"`   // 1-based line number; 0 = not set
	Column int    `json:"column,omitempty"` // 1-based column number; 0 = not set
	Offset int    `json:"offset,omitempty"` // Byte offset from file start; 0 = not set
}

// IsValid returns true if the position has a file set.
func (p Position) IsValid() bool {
	return p.File != ""
}

// Equal reports whether two positions are identical.
func (p Position) Equal(other Position) bool {
	return p.File == other.File &&
		p.Line == other.Line &&
		p.Column == other.Column &&
		p.Offset == other.Offset
}

// Compare returns -1, 0, or 1 depending on whether p is less than, equal to,
// or greater than other. Positions are ordered by file, then line, then column,
// then offset. This is consistent with Equal: Compare returns 0 iff Equal returns true.
func (p Position) Compare(other Position) int {
	if c := cmp.Compare(p.File, other.File); c != 0 {
		return c
	}

	if c := cmp.Compare(p.Line, other.Line); c != 0 {
		return c
	}

	if c := cmp.Compare(p.Column, other.Column); c != 0 {
		return c
	}

	return cmp.Compare(p.Offset, other.Offset)
}

// String returns a human-readable representation.
func (p Position) String() string {
	if p.Line == 0 {
		return p.File
	}

	if p.Column == 0 {
		return fmt.Sprintf("%s:%d", p.File, p.Line)
	}

	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}

// Range represents a span in source code from Start to End.
type Range struct {
	Start Position `json:"start"` // Required: start position
	End   Position `json:"end"`   // Optional: end position
}

// IsValid returns true if the range has a valid start position.
func (r Range) IsValid() bool {
	return r.Start.IsValid()
}

// HasEnd returns true if the range has an end position set.
func (r Range) HasEnd() bool {
	return r.End.Line > 0
}

// Equal reports whether two ranges are identical.
func (r Range) Equal(other Range) bool {
	return r.Start.Equal(other.Start) && r.End.Equal(other.End)
}

// Compare returns -1, 0, or 1 depending on whether r is less than, equal to,
// or greater than other. Ranges are ordered by start position, then end position.
func (r Range) Compare(other Range) int {
	if c := r.Start.Compare(other.Start); c != 0 {
		return c
	}

	return r.End.Compare(other.End)
}

// sameFileAs checks if this range is in the same file as another range.
func (r Range) sameFileAs(other Range) bool {
	return r.Start.File == other.Start.File
}

// hasLineInfo checks if both ranges have line information available.
func (r Range) hasLineInfo(other Range) bool {
	return r.Start.Line > 0 && other.Start.Line > 0
}

// containsSameFile checks if both positions are in the same file.
func (r Range) containsSameFile(p Position) bool {
	return r.Start.File == p.File && (r.End.File == "" || r.End.File == p.File)
}

// hasLineRange checks if the range has defined line numbers and evaluates if p is within line bounds.
func (r Range) hasLineRange(p Position) bool {
	if r.Start.Line == 0 || p.Line == 0 {
		return false
	}

	if p.Line < r.Start.Line {
		return false
	}

	if r.End.Line > 0 && p.Line > r.End.Line {
		return false
	}

	return r.checkColumnRange(p)
}

// checkColumnRange checks if p's column is within the range's column bounds.
func (r Range) checkColumnRange(p Position) bool {
	if r.Start.Column > 0 && p.Column > 0 && p.Column < r.Start.Column {
		return false
	}

	if r.End.Column > 0 && p.Column > 0 && p.Column > r.End.Column {
		return false
	}

	return true
}

// containsByOffset checks if position is within range using byte offsets.
func (r Range) containsByOffset(p Position) bool {
	if r.Start.Offset > 0 && p.Offset > 0 {
		if p.Offset < r.Start.Offset {
			return false
		}

		if r.End.Offset > 0 && p.Offset > r.End.Offset {
			return false
		}

		return true
	}

	return false
}

// Contains reports whether the position is within the range.
// Checks same file, line range, and offset when line ranges aren't available.
func (r Range) Contains(p Position) bool {
	if !r.containsSameFile(p) {
		return false
	}

	// If both range and position have line info, use line-based check exclusively
	if r.Start.Line > 0 && p.Line > 0 {
		return r.hasLineRange(p)
	}

	return r.containsByOffset(p)
}

// Overlaps reports whether this range overlaps with another range.
// Two ranges overlap if they share at least one position.
func (r Range) Overlaps(other Range) bool {
	if !r.sameFileAs(other) {
		return false
	}

	// Check using line numbers if available
	if r.hasLineInfo(other) {
		return r.overlapsByLine(other)
	}

	// Fall back to offset-based check
	return r.overlapsByOffset(other)
}

// overlapsByLine checks overlap using line/column coordinates.
func (r Range) overlapsByLine(other Range) bool {
	// Get effective end positions
	rEndLine := r.End.Line
	if rEndLine == 0 {
		rEndLine = r.Start.Line
	}

	otherEndLine := other.End.Line
	if otherEndLine == 0 {
		otherEndLine = other.Start.Line
	}

	// Ranges overlap if:
	// - This range starts before or at the end of other range
	// - AND this range ends after or at the start of other range
	return r.Start.Line <= otherEndLine && rEndLine >= other.Start.Line
}

// overlapsByOffset checks overlap using byte offsets.
func (r Range) overlapsByOffset(other Range) bool {
	if r.Start.Offset == 0 || other.Start.Offset == 0 {
		// Cannot determine overlap without offsets
		return false
	}

	rEndOffset := r.End.Offset
	if rEndOffset == 0 {
		rEndOffset = r.Start.Offset
	}

	otherEndOffset := other.End.Offset
	if otherEndOffset == 0 {
		otherEndOffset = other.Start.Offset
	}

	return r.Start.Offset <= otherEndOffset && rEndOffset >= other.Start.Offset
}

// Intersection returns the overlapping region of two ranges, or nil if they don't overlap.
func (r Range) Intersection(other Range) *Range {
	if !r.Overlaps(other) {
		return nil
	}

	// Use line-based intersection if available
	if r.hasLineInfo(other) {
		return r.intersectionByLine(other)
	}

	// Fall back to offset-based
	return r.intersectionByOffset(other)
}

// intersectionByLine computes intersection using line/column coordinates.
func (r Range) intersectionByLine(other Range) *Range {
	// Determine max start
	start := r.Start
	if other.Start.Line > start.Line ||
		(other.Start.Line == start.Line && other.Start.Column > start.Column) {
		start = other.Start
	}

	// Determine min end
	end := r.End
	if end.Line == 0 {
		end = r.Start
	}

	otherEnd := other.End
	if otherEnd.Line == 0 {
		otherEnd = other.Start
	}

	if otherEnd.Line < end.Line || (otherEnd.Line == end.Line && otherEnd.Column < end.Column) {
		end = otherEnd
	}

	// If start equals end (same position), return single point range
	if start.Line == end.Line && start.Column == end.Column {
		return &Range{Start: start, End: Position{}}
	}

	return &Range{Start: start, End: end}
}

// intersectionByOffset computes intersection using byte offsets.
func (r Range) intersectionByOffset(other Range) *Range {
	startOffset := max(r.Start.Offset, other.Start.Offset)

	rEndOffset := r.End.Offset
	if rEndOffset == 0 {
		rEndOffset = r.Start.Offset
	}

	otherEndOffset := other.End.Offset
	if otherEndOffset == 0 {
		otherEndOffset = other.Start.Offset
	}

	endOffset := min(rEndOffset, otherEndOffset)

	return &Range{
		Start: Position{File: r.Start.File, Offset: startOffset},
		End:   Position{File: r.Start.File, Offset: endOffset},
	}
}

// Adjacent reports whether this range is immediately adjacent to another range.
// Adjacent means one range ends exactly where the other begins.
func (r Range) Adjacent(other Range) bool {
	if !r.sameFileAs(other) {
		return false
	}

	// Try line-based adjacency first
	if r.End.Line > 0 && other.Start.Line > 0 {
		if r.End.Line == other.Start.Line && r.End.Column == other.Start.Column {
			return true
		}

		if other.End.Line == r.Start.Line && other.End.Column == r.Start.Column {
			return true
		}
	}

	// Fall back to offset-based adjacency
	if r.End.Offset > 0 && other.Start.Offset > 0 && r.End.Offset == other.Start.Offset {
		return true
	}

	if other.End.Offset > 0 && r.Start.Offset > 0 && other.End.Offset == r.Start.Offset {
		return true
	}

	return false
}

// Pos creates a Position with the given file, line, and column.
func Pos(file string, line, column int) Position {
	return Position{File: file, Line: line, Column: column}
}

// NewRange creates a Range with the given file, start/end lines, and columns.
func NewRange(file string, startLine, startCol, endLine, endCol int) Range {
	return Range{
		Start: Position{File: file, Line: startLine, Column: startCol},
		End:   Position{File: file, Line: endLine, Column: endCol},
	}
}

// NewRangePtr creates a pointer to a Range with the given file, start/end lines, and columns.
func NewRangePtr(file string, startLine, startCol, endLine, endCol int) *Range {
	r := NewRange(file, startLine, startCol, endLine, endCol)

	return &r
}
