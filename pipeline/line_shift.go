package pipeline

import (
	"bytes"
	"cmp"
	"slices"
)

// LineShiftEntry records how many lines were added or removed starting at a byte offset.
type LineShiftEntry struct {
	// ByteOffset is where the line shift takes effect in the original content.
	// For pure insertions, this equals the edit's Offset. For replacements,
	// this equals Offset + Length (shift starts after the replaced region).
	ByteOffset int
	// Delta is the net line change: positive means lines were added, negative means removed.
	Delta int
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

		if delta == 0 {
			continue
		}

		effectOffset := edit.Offset
		if edit.Length > 0 {
			effectOffset = edit.EndOffset()
		}

		entries = append(entries, LineShiftEntry{
			ByteOffset: effectOffset,
			Delta:      delta,
		})
	}

	slices.SortFunc(entries, func(a, b LineShiftEntry) int {
		return cmp.Compare(a.ByteOffset, b.ByteOffset)
	})

	return &LineShiftMap{lineOffsets: lineOffsets, entries: entries}
}

// ShiftedLine returns the new 1-based line number for a given original line.
// It sums the delta of all entries whose effect offset is at or before the
// start of the requested line.
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
		if entry.ByteOffset > offset {
			break
		}

		cumulative += entry.Delta
	}

	return originalLine + cumulative
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
