package pipeline

import (
	"bytes"
	"errors"
	"fmt"

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

// OffsetProvider handles findings with byte-offset Range information.
// This is the most accurate text-based provider — findings with byte offsets
// bypass line/column conversion entirely.
type OffsetProvider struct{}

// Name returns the provider name.
func (OffsetProvider) Name() string { return "byte-offset" }

// CanHandle reports whether the finding has explicit byte-offset range information.
func (OffsetProvider) CanHandle(f finding.Finding) bool {
	if f.BeforeCode == "" && f.AfterCode == "" {
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

	return []FixEdit{{
		Offset:      start,
		Length:      end - start,
		Replacement: []byte(f.AfterCode),
		Source:      f,
	}}, nil
}

// LineProvider handles findings with line/column information by converting
// to byte offsets. It supports range-based, insertion, and replacement fixes.
type LineProvider struct{}

// Name returns the provider name.
func (LineProvider) Name() string { return "line-column" }

// CanHandle reports whether the finding has line/column position info.
func (LineProvider) CanHandle(f finding.Finding) bool {
	if f.BeforeCode == "" && f.AfterCode == "" {
		return false
	}

	return f.Position.Line > 0
}

// Edits produces byte-level edits from line/column information.
func (LineProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	idx := buildLineOffsetIndex(content)

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

		return []FixEdit{{
			Offset:      start + loc,
			Length:      len(before),
			Replacement: []byte(f.AfterCode),
			Source:      f,
		}}, nil
	}

	return []FixEdit{{
		Offset:      start,
		Length:      end - start,
		Replacement: []byte(f.AfterCode),
		Source:      f,
	}}, nil
}

func lineProviderInsertionEdit(content []byte, f finding.Finding, idx []int) ([]FixEdit, error) {
	offset, err := indexLineColToOffset(idx, len(content), f.Position.Line, f.Position.Column)
	if err != nil {
		return nil, ErrPositionUnresolvable
	}

	replacement := append([]byte(f.AfterCode), '\n')

	return []FixEdit{{
		Offset:      offset,
		Length:      0,
		Replacement: replacement,
		Source:      f,
	}}, nil
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

	return []FixEdit{{
		Offset:      offset,
		Length:      len(before),
		Replacement: []byte(f.AfterCode),
		Source:      f,
	}}, nil
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
	return f.BeforeCode != ""
}

// Edits locates BeforeCode in the content using substring matching and produces edits.
func (SubstringProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	before := []byte(f.BeforeCode)

	occurrences := findAllOccurrences(content, before)
	if len(occurrences) == 0 {
		return nil, nil
	}

	best := occurrences[0]

	if f.Position.Line > 0 && len(occurrences) > 1 {
		bestDist := offsetLineDistance(content, best, f.Position.Line)
		for _, idx := range occurrences[1:] {
			dist := offsetLineDistance(content, idx, f.Position.Line)
			if dist < bestDist {
				bestDist = dist
				best = idx
			}
		}
	}

	return []FixEdit{{
		Offset:      best,
		Length:      len(before),
		Replacement: []byte(f.AfterCode),
		Source:      f,
	}}, nil
}

var (
	// ErrPositionUnresolvable indicates a line/column could not be mapped to a byte offset.
	ErrPositionUnresolvable = errors.New("position unresolvable in content")

	errInvalidLine   = errors.New("invalid line number")
	errLineBeyondEOF = errors.New("line beyond end of file")
	errColumnBeyond  = errors.New("column beyond end of line")
)

// lineColToOffset converts a 1-based line and column to a 0-based byte offset.
// Builds a line offset index per call; for batch processing, prefer
// indexLineColToOffset with a pre-built index.
//
//nolint:unparam // col is always 1 in current callers but is part of the general-purpose API
func lineColToOffset(content []byte, line, col int) (int, error) {
	return indexLineColToOffset(buildLineOffsetIndex(content), len(content), line, col)
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
	lineCount := 1
	for _, b := range content {
		if b == '\n' {
			lineCount++
		}
	}

	index := make([]int, 0, lineCount)
	index = append(index, 0) // line 1 starts at offset 0

	for i, b := range content {
		if b == '\n' && i+1 < len(content) {
			index = append(index, i+1)
		}
	}

	return index
}

// findAllOccurrences returns all starting byte positions of needle in haystack.
func findAllOccurrences(haystack, needle []byte) []int {
	if len(needle) == 0 {
		return nil
	}

	var results []int
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

// offsetLineDistance counts newlines before a byte offset and returns the
// absolute difference from targetLine.
func offsetLineDistance(content []byte, offset, targetLine int) int {
	line := 1

	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}

	diff := line - targetLine
	if diff < 0 {
		return -diff
	}

	return diff
}
