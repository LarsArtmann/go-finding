package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/larsartmann/go-finding"
)

// FixEdit represents a single byte-level edit operation. //nolint:recvcheck // value receivers for read-only, pointer for UnmarshalJSON
// Edits are applied to file content at specific byte offsets,
// enabling deterministic, order-independent transformations.
//
// Domain-specific FixProviders (e.g., Go AST, Rust syn) should produce
// FixEdits from parsed IR rather than substring matching. The default
// text-based providers are fragile fallbacks — strongly prefer AST-aware
// implementations for production use.
type FixEdit struct { //nolint:recvcheck // value receivers for read-only, pointer for UnmarshalJSON
	// Offset is the 0-based byte offset where the edit begins.
	Offset int `json:"offset"`
	// Length is the number of bytes to remove starting at Offset.
	// Use 0 for pure insertions.
	Length int `json:"length"`
	// Replacement is the new bytes to write at Offset after removal.
	// Use nil or empty for pure deletions.
	Replacement []byte `json:"replacement,omitempty"`
	// Source is the finding that produced this edit.
	Source finding.Finding `json:"-"`
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

// MarshalJSON implements json.Marshaler for FixEdit.
// Replacement is base64-encoded per JSON spec for []byte fields.
// Source is omitted from JSON output (use property bag for SARIF round-tripping).
func (e FixEdit) MarshalJSON() ([]byte, error) {
	type jsonEdit struct {
		Offset      int    `json:"offset"`
		Length      int    `json:"length"`
		Replacement []byte `json:"replacement,omitempty"`
	}

	return json.Marshal(jsonEdit{ //nolint:wrapcheck // standard JSON marshaling
		Offset:      e.Offset,
		Length:      e.Length,
		Replacement: e.Replacement,
	})
}

// UnmarshalJSON implements json.Unmarshaler for FixEdit.
func (e *FixEdit) UnmarshalJSON(data []byte) error {
	type jsonEdit struct {
		Offset      int    `json:"offset"`
		Length      int    `json:"length"`
		Replacement []byte `json:"replacement,omitempty"`
	}

	var j jsonEdit
	if err := json.Unmarshal(data, &j); err != nil {
		return fmt.Errorf("unmarshal fix edit: %w", err)
	}

	e.Offset = j.Offset
	e.Length = j.Length
	e.Replacement = j.Replacement

	return nil
}

// ToSARIFProperties converts the edit to a SARIF property bag for round-tripping.
// Store in the finding's Metadata under keys prefixed with "go-finding/edit/".
func (e FixEdit) ToSARIFProperties() map[string]string {
	props := map[string]string{
		"go-finding/edit/offset": strconv.Itoa(e.Offset),
		"go-finding/edit/length": strconv.Itoa(e.Length),
	}

	if len(e.Replacement) > 0 {
		props["go-finding/edit/replacement"] = string(e.Replacement)
	}

	return props
}

// FixEditFromSARIFProperties reconstructs a FixEdit from SARIF property bag values.
// Returns nil if the required offset/length keys are missing.
func FixEditFromSARIFProperties(props map[string]string) *FixEdit {
	offsetStr, ok := props["go-finding/edit/offset"]
	if !ok {
		return nil
	}

	lengthStr, ok := props["go-finding/edit/length"]
	if !ok {
		return nil
	}

	var offset, length int
	if v, err := strconv.Atoi(offsetStr); err != nil {
		return nil
	} else {
		offset = v
	}

	if v, err := strconv.Atoi(lengthStr); err != nil {
		return nil
	} else {
		length = v
	}

	edit := FixEdit{ //nolint:exhaustruct // partial construction from SARIF props
		Offset: offset,
		Length: length,
	}

	if replacement, ok := props["go-finding/edit/replacement"]; ok {
		edit.Replacement = []byte(replacement)
	}

	return &edit
}
