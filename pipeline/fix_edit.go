package pipeline

import (
	"errors"
	"fmt"

	"github.com/larsartmann/go-finding"
)

// FixEdit represents a single byte-level edit operation.
// Edits are applied to file content at specific byte offsets,
// enabling deterministic, order-independent transformations.
//
// Domain-specific FixProviders (e.g., Go AST, Rust syn) should produce
// FixEdits from parsed IR rather than substring matching. The default
// text-based providers are fragile fallbacks — strongly prefer AST-aware
// implementations for production use.
type FixEdit struct {
	// Offset is the 0-based byte offset where the edit begins.
	Offset int
	// Length is the number of bytes to remove starting at Offset.
	// Use 0 for pure insertions.
	Length int
	// Replacement is the new bytes to write at Offset after removal.
	// Use nil or empty for pure deletions.
	Replacement []byte
	// Source is the finding that produced this edit.
	Source finding.Finding
}

// EndOffset returns the byte offset immediately after the edit's removal range.
func (e FixEdit) EndOffset() int {
	return e.Offset + e.Length
}

// IsInsert reports whether this is a pure insertion (no bytes removed).
func (e FixEdit) IsInsert() bool {
	return e.Length == 0 && len(e.Replacement) > 0
}

// IsDelete reports whether this is a pure deletion (no bytes inserted).
func (e FixEdit) IsDelete() bool {
	return e.Length > 0 && len(e.Replacement) == 0
}

// Overlaps reports whether two edits touch overlapping byte ranges.
// Adjacent edits (one ends exactly where the other starts) do NOT overlap.
func (e FixEdit) Overlaps(other FixEdit) bool {
	if e.Offset < other.EndOffset() && other.Offset < e.EndOffset() {
		return true
	}

	// Two zero-length insertions at the same point overlap.
	if e.Length == 0 && other.Length == 0 && e.Offset == other.Offset {
		return true
	}

	return false
}

var (
	errNegativeOffset = errors.New("fix edit: negative offset")
	errNegativeLength = errors.New("fix edit: negative length")
)

// Validate checks the edit for consistency.
func (e FixEdit) Validate() error {
	if e.Offset < 0 {
		return fmt.Errorf("%w: %d", errNegativeOffset, e.Offset)
	}

	if e.Length < 0 {
		return fmt.Errorf("%w: %d", errNegativeLength, e.Length)
	}

	return nil
}
