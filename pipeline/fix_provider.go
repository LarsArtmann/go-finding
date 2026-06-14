package pipeline

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/larsartmann/go-finding"
)

// FixProvider converts findings into byte-level edits for a specific domain.
//
// For production use, implement domain-specific providers that use AST, IR, or
// typed analysis (e.g., Go's go/ast, Rust's syn, TypeScript's compiler API).
// The default text-based providers are fallbacks for when no domain provider is
// available — they are inherently fragile because they rely on substring matching
// and line/column heuristics rather than structural understanding.
//
// Example domain-specific provider:
//
//	type GoASTProvider struct{}
//	func (p *GoASTProvider) Name() string { return "go-ast" }
//	func (p *GoASTProvider) CanHandle(f finding.Finding) bool {
//	    return strings.HasSuffix(f.Position.File, ".go")
//	}
//	func (p *GoASTProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
//	    fset := token.NewFileSet()
//	    file, _ := parser.ParseFile(fset, f.Position.File, content, parser.ParseComments)
//	    // ... produce precise byte-level edits from the AST ...
//	}
type FixProvider interface {
	// Name returns the provider's name (e.g., "go-ast", "rust-syntax", "byte-offset").
	Name() string
	// CanHandle reports whether this provider can produce edits for the given finding.
	CanHandle(f finding.Finding) bool
	// Edits converts a finding into one or more byte-level edits.
	// content is the file's raw bytes, available for context and validation.
	// Return an empty slice if the finding cannot be resolved to edits.
	Edits(content []byte, f finding.Finding) ([]FixEdit, error)
}

// lineIndexAware is an optional interface for FixProviders that can accept a
// pre-built line offset index. When a provider implements this interface,
// FixEngine builds the index once per file and passes it to every finding,
// avoiding O(n) rebuilds per finding (where n = file size).
type lineIndexAware interface {
	EditsWithLineIndex(content []byte, lineIndex []int, f finding.Finding) ([]FixEdit, error)
}

// OffsetProvider handles findings with byte-offset Range information.
// This is the most accurate text-based provider — findings with byte offsets
// bypass line/column conversion entirely.
type OffsetProvider struct{}

// Name returns the provider name.
func (OffsetProvider) Name() string { return "byte-offset" }

// CanHandle reports whether the finding has explicit byte-offset range information.
func (OffsetProvider) CanHandle(f finding.Finding) bool {
	if !f.HasCodeChange() {
		return false
	}

	return f.Range != nil &&
		f.Range.Start.Offset >= 0 && f.Range.End.Offset >= 0 &&
		f.Range.Start.Offset < f.Range.End.Offset
}

// Edits produces byte-level edits from offset-based range information.
func (OffsetProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	start := f.Range.Start.Offset
	end := f.Range.End.Offset

	if start < 0 || end < start || end > len(content) {
		return nil, nil
	}

	if f.BeforeCode != "" {
		before := []byte(f.BeforeCode)
		if !bytes.Equal(content[start:end], before) {
			return nil, nil
		}
	}

	return []FixEdit{newReplacementEdit(start, end-start, f)}, nil
}

// LineProvider handles findings with line/column information by converting
// to byte offsets. It supports range-based, insertion, and replacement fixes.
type LineProvider struct{}

// Name returns the provider name.
func (LineProvider) Name() string { return "line-column" }

// CanHandle reports whether the finding has line/column position info.
func (LineProvider) CanHandle(f finding.Finding) bool {
	if !f.HasCodeChange() {
		return false
	}

	return f.Position.Line > 0
}

// Edits produces byte-level edits from line/column information.
func (p LineProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	return p.EditsWithLineIndex(content, buildLineOffsetIndex(content), f)
}

// EditsWithLineIndex produces byte-level edits using a pre-built line offset
// index, avoiding an O(n) rebuild per finding.
func (LineProvider) EditsWithLineIndex(content []byte, idx []int, f finding.Finding) ([]FixEdit, error) {
	if f.Range != nil && f.Range.HasEnd() && f.Range.Start.Line > 0 && f.Range.End.Line > 0 {
		return lineProviderRangeEdits(content, f, idx)
	}

	if f.BeforeCode == "" && f.AfterCode != "" {
		return lineProviderInsertionEdit(content, f, idx)
	}

	if f.BeforeCode != "" {
		return lineProviderReplacementEdit(content, f, idx)
	}

	return nil, nil
}

func lineProviderRangeEdits(content []byte, f finding.Finding, idx []int) ([]FixEdit, error) {
	start, err := indexLineColToOffset(idx, len(content), f.Range.Start.Line, f.Range.Start.Column)
	if err != nil {
		return nil, ErrPositionUnresolvable
	}

	end, err := indexLineColToOffset(idx, len(content), f.Range.End.Line, f.Range.End.Column)
	if err != nil {
		return nil, ErrPositionUnresolvable
	}

	if end < start || end > len(content) {
		return nil, nil
	}

	if f.BeforeCode != "" {
		rangeContent := content[start:end]
		before := []byte(f.BeforeCode)

		loc := bytes.Index(rangeContent, before)
		if loc < 0 {
			return nil, nil
		}

		return []FixEdit{newReplacementEdit(start+loc, len(before), f)}, nil
	}

	return []FixEdit{newReplacementEdit(start, end-start, f)}, nil
}

func lineProviderInsertionEdit(content []byte, f finding.Finding, idx []int) ([]FixEdit, error) {
	offset, err := indexLineColToOffset(idx, len(content), f.Position.Line, f.Position.Column)
	if err != nil {
		return nil, ErrPositionUnresolvable
	}

	replacement := append([]byte(f.AfterCode), '\n')

	return []FixEdit{{Offset: offset, Length: 0, Replacement: replacement, Source: f}}, nil
}

func lineProviderReplacementEdit(content []byte, f finding.Finding, idx []int) ([]FixEdit, error) {
	offset, err := indexLineColToOffset(idx, len(content), f.Position.Line, f.Position.Column)
	if err != nil {
		return nil, ErrPositionUnresolvable
	}

	before := []byte(f.BeforeCode)
	end := offset + len(before)

	if end > len(content) {
		return nil, nil
	}

	if !bytes.Equal(content[offset:end], before) {
		return nil, nil
	}

	return []FixEdit{newReplacementEdit(offset, len(before), f)}, nil
}

// SubstringProvider is a fallback provider that locates BeforeCode in the content
// using substring matching. It is less accurate than OffsetProvider and LineProvider
// because substring positions can be ambiguous when the same text appears multiple times.
//
// Strongly prefer registering a domain-specific FixProvider for your language or format.
type SubstringProvider struct{}

// Name returns the provider name.
func (SubstringProvider) Name() string { return "substring" }

// CanHandle reports whether the finding has BeforeCode for substring matching.
func (SubstringProvider) CanHandle(f finding.Finding) bool {
	return f.HasCodeChange() && f.BeforeCode != ""
}

// Edits locates BeforeCode in the content using substring matching and produces edits.
func (p SubstringProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	return p.EditsWithLineIndex(content, buildLineOffsetIndex(content), f)
}

// EditsWithLineIndex locates BeforeCode using a pre-built line offset index,
// avoiding an O(n) rebuild per finding when multiple occurrences require
// disambiguation by line proximity.
func (SubstringProvider) EditsWithLineIndex(content []byte, lineIndex []int, f finding.Finding) ([]FixEdit, error) {
	before := []byte(f.BeforeCode)

	occurrences := findAllOccurrences(content, before)
	if len(occurrences) == 0 {
		return nil, nil
	}

	best := occurrences[0]

	if f.Position.Line > 0 && len(occurrences) > 1 {
		// Use binary search on the pre-built line offset index for each
		// occurrence, reducing per-occurrence cost from O(F) to O(log F).
		bestDist := offsetLineDistance(lineIndex, best, f.Position.Line)

		for _, idx := range occurrences[1:] {
			dist := offsetLineDistance(lineIndex, idx, f.Position.Line)
			if dist < bestDist {
				bestDist = dist
				best = idx
			}
		}
	}

	return []FixEdit{newReplacementEdit(best, len(before), f)}, nil
}

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
