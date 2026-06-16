package pipeline

import (
	"bytes"
	"cmp"
	"slices"

	"github.com/larsartmann/go-finding"
)

// LineShiftEntry records how lines and columns shift starting at a byte offset.
type LineShiftEntry struct {
	// ByteOffset is where the shift takes effect in the original content.
	// For pure insertions, this equals the edit's Offset. For replacements,
	// this equals Offset + Length (shift starts after the replaced region).
	ByteOffset int
	// Delta is the net line change: positive means lines were added, negative means removed.
	Delta int
	// ByteDelta is the net byte change on the edit's first line: len(replacement) - length.
	// Non-zero only for single-line edits that don't change the line count
	// (Delta == 0). Used to shift column positions of findings on the same line.
	ByteDelta int
}

// LineShiftMap tracks how line numbers shift after byte-level edits.
// Create one with [NewLineShiftMap] and query with [LineShiftMap.ShiftedLine].
//
// Lines within a deleted range return their original line number (unshifted),
// which is NOT a valid new line number — the caller must handle this case.
type LineShiftMap struct {
	lineOffsets []int
	entries     []LineShiftEntry
}

// NewLineShiftMap computes a line shift map from the original content and
// the applied edits. Edits must be the edits that were actually applied
// (not conflicting ones). The order does not matter — entries are sorted internally.
func NewLineShiftMap(original []byte, edits []FixEdit) *LineShiftMap {
	if len(edits) == 0 {
		return &LineShiftMap{lineOffsets: nil, entries: []LineShiftEntry{}}
	}

	lineOffsets := buildLineOffsetIndex(original)
	entries := make([]LineShiftEntry, 0, len(edits))

	for _, edit := range edits {
		removedNewlines := countNewlines(original, edit.Offset, edit.EndOffset())
		addedNewlines := bytes.Count(edit.Replacement, []byte{'\n'})
		delta := addedNewlines - removedNewlines
		byteDelta := len(edit.Replacement) - edit.Length

		// Keep entries that change line count OR change bytes within a single line
		// (the latter enables column shifting).
		if delta == 0 && byteDelta == 0 {
			continue
		}

		effectOffset := edit.Offset
		if edit.Length > 0 {
			effectOffset = edit.EndOffset()
		}

		entries = append(entries, LineShiftEntry{
			ByteOffset: effectOffset,
			Delta:      delta,
			ByteDelta:  byteDelta,
		})
	}

	slices.SortFunc(entries, func(a, b LineShiftEntry) int {
		return cmp.Compare(a.ByteOffset, b.ByteOffset)
	})

	return &LineShiftMap{lineOffsets: lineOffsets, entries: entries}
}

// ShiftedLine returns the new 1-based line number for a given original line.
// It sums the delta of all entries whose effect offset is at or before the
// start of the requested line. Only entries that change line count (Delta != 0)
// are considered.
//
// For pure insertions, lines at the insertion point ARE shifted.
// For replacements, lines in the replaced region are NOT shifted
// (they retain their position from the first line of the replacement).
// Lines within a deleted range return their original number, which is NOT
// a valid new line number.
func (m *LineShiftMap) ShiftedLine(originalLine int) int {
	if len(m.entries) == 0 {
		return originalLine
	}

	if originalLine < 1 || originalLine > len(m.lineOffsets) {
		return originalLine
	}

	offset := m.lineOffsets[originalLine-1]
	cumulative := 0

	for _, entry := range m.entries {
		if entry.Delta == 0 {
			continue
		}

		if entry.ByteOffset > offset {
			break
		}

		cumulative += entry.Delta
	}

	return originalLine + cumulative
}

// ShiftedPosition returns a copy of pos with the line and column shifted to
// reflect applied edits.
//
// Line shifting follows the same rules as [LineShiftMap.ShiftedLine].
//
// Column shifting applies to single-line edits (no line-count change) that
// occur on the same line and before the position. Multi-line edits do not
// affect columns on subsequent lines. Positions inside an edited multi-line
// region have unreliable columns (only the line is shifted).
func (m *LineShiftMap) ShiftedPosition(pos finding.Position) finding.Position {
	if len(m.entries) == 0 {
		return pos
	}

	if pos.Line < 1 || pos.Line > len(m.lineOffsets) {
		return pos
	}

	lineStart := m.lineOffsets[pos.Line-1]
	posByteOffset := lineStart

	if pos.Column > 1 {
		posByteOffset += pos.Column - 1
	}

	lineShift := 0
	colShift := 0

	for _, entry := range m.entries {
		if entry.ByteOffset > posByteOffset {
			break
		}

		if entry.Delta != 0 {
			// Multi-line edit: affects line count but not column (for lines after it).
			lineShift += entry.Delta

			continue
		}

		// Single-line byte change: shift column only if on the same line.
		if entry.ByteDelta != 0 && entry.ByteOffset >= lineStart {
			colShift += entry.ByteDelta
		}
	}

	result := pos
	result.Line = pos.Line + lineShift
	result.Column = max(1, pos.Column+colShift)

	return result
}

// ShiftedRange returns a copy of r with both Start and End positions shifted.
// Returns nil if r is nil.
func (m *LineShiftMap) ShiftedRange(r *finding.Range) *finding.Range {
	if r == nil {
		return nil
	}

	result := *r
	result.Start = m.ShiftedPosition(r.Start)
	result.End = m.ShiftedPosition(r.End)

	return &result
}

// Entries returns a copy of the shift entries sorted by byte offset.
func (m *LineShiftMap) Entries() []LineShiftEntry {
	return slices.Clone(m.entries)
}

// countNewlines counts newline bytes in content[start:end].
func countNewlines(content []byte, start, end int) int {
	if end > len(content) {
		end = len(content)
	}

	return bytes.Count(content[start:end], []byte{'\n'})
}

// Compile-time check.
var _ = (*LineShiftMap)(nil)
