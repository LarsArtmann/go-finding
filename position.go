package finding

import (
	"cmp"
	"fmt"
)

// Position represents a location in source code.
//
// Sentinel values:
//   - Line, Column: 0 means "not set" (1-based, so 0 is never valid).
//   - Offset: -1 means "not set" (0-based, so 0 means "start of file" which IS valid).
//
// This means Position{} (the zero value) has Offset=0, which IsZero() reports as true
// but HasOffset() also reports as true. Use HasLocation() to check for a meaningful
// position (file + line), or IsZero() to check for the completely-uninitialized state.
type Position struct {
	File   string `json:"file"`             // Required: file path
	Line   int    `json:"line,omitempty"`   // 1-based line number; 0 = not set
	Column int    `json:"column,omitempty"` // 1-based column number; 0 = not set
	Offset int    `json:"offset,omitempty"` // 0-based byte offset; -1 = not set
}

// IsValid returns true if the position has a file set and non-negative line/column.
func (p Position) IsValid() bool {
	return p.File != "" && p.Line >= 0 && p.Column >= 0
}

// IsZero reports whether the position is the zero value (no information set).
func (p Position) IsZero() bool {
	return p.File == "" && p.Line == 0 && p.Column == 0 && p.Offset == 0
}

// HasLocation reports whether the position has a file and line number.
func (p Position) HasLocation() bool {
	return p.File != "" && p.Line > 0
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

