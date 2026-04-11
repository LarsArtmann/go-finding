package finding

import "fmt"

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
	Start Position `json:"start"`         // Required: start position
	End   Position `json:"end,omitempty"` // Optional: end position
}

// IsValid returns true if the range has a valid start position.
func (r Range) IsValid() bool {
	return r.Start.IsValid()
}

// Contains reports whether the position is within the range.
// This is a simple check: same file and position >= start and <= end.
func (r Range) Contains(p Position) bool {
	if r.Start.File != p.File {
		return false
	}
	if r.End.File != "" && r.End.File != p.File {
		return false
	}

	// If we have line numbers, use them
	if r.Start.Line > 0 && p.Line > 0 {
		if p.Line < r.Start.Line {
			return false
		}
		if r.End.Line > 0 && p.Line > r.End.Line {
			return false
		}
		if p.Line == r.Start.Line && r.Start.Column > 0 && p.Column > 0 && p.Column < r.Start.Column {
			return false
		}
		if p.Line == r.End.Line && r.End.Column > 0 && p.Column > 0 && p.Column > r.End.Column {
			return false
		}
		return true
	}

	// Fall back to offset comparison
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
