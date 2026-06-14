package finding

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
	if r.Start.Offset < 0 || other.Start.Offset < 0 {
		// Cannot determine overlap without offsets
		return false
	}

	rEndOffset := r.End.Offset
	if rEndOffset < 0 {
		rEndOffset = r.Start.Offset
	}

	otherEndOffset := other.End.Offset
	if otherEndOffset < 0 {
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
		return &Range{Start: start, End: Position{}} //nolint:exhaustruct
	}

	return &Range{Start: start, End: end}
}

// intersectionByOffset computes intersection using byte offsets.
func (r Range) intersectionByOffset(other Range) *Range {
	startOffset := max(r.Start.Offset, other.Start.Offset)

	rEndOffset := r.End.Offset
	if rEndOffset < 0 {
		rEndOffset = r.Start.Offset
	}

	otherEndOffset := other.End.Offset
	if otherEndOffset < 0 {
		otherEndOffset = other.Start.Offset
	}

	endOffset := min(rEndOffset, otherEndOffset)

	return &Range{
		Start: Position{File: r.Start.File, Offset: startOffset}, //nolint:exhaustruct
		End:   Position{File: r.Start.File, Offset: endOffset},   //nolint:exhaustruct
	}
}

// columnAdjacent checks if two columns represent adjacent positions.
// Returns true if either both are 0 (line-based adjacency) or both are
// positive and equal (column-based adjacency).
func columnAdjacent(endCol, startCol int) bool {
	if endCol == 0 && startCol == 0 {
		return true
	}

	return endCol > 0 && startCol > 0 && endCol == startCol
}

// offsetAdjacent checks if two offsets represent adjacent positions.
// Returns true if both are positive and equal.
func offsetAdjacent(endOffset, startOffset int) bool {
	return endOffset > 0 && startOffset > 0 && endOffset == startOffset
}

// Adjacent reports whether this range is immediately adjacent to another range.
// Adjacent means one range ends exactly where the other begins.
func (r Range) Adjacent(other Range) bool {
	if !r.sameFileAs(other) {
		return false
	}

	// Check line-based adjacency if both ends have line info.
	if r.End.Line > 0 && other.Start.Line > 0 {
		if r.End.Line == other.Start.Line && columnAdjacent(r.End.Column, other.Start.Column) {
			return true
		}

		if other.End.Line == r.Start.Line && columnAdjacent(other.End.Column, r.Start.Column) {
			return true
		}
	}

	// Fall back to offset-based adjacency when line info is not available.
	if !r.hasLineInfo(other) {
		startMatch := offsetAdjacent(r.End.Offset, other.Start.Offset)
		endMatch := offsetAdjacent(other.End.Offset, r.Start.Offset)

		return startMatch || endMatch
	}

	return false
}
