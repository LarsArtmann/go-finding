package pipeline

import (
	"bytes"

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

	return f.Range != nil && f.Range.Length() > 0
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
	start, err := resolveLineCol(idx, len(content), f.Range.Start.Line, f.Range.Start.Column)
	if err != nil {
		return nil, err
	}

	end, err := resolveLineCol(idx, len(content), f.Range.End.Line, f.Range.End.Column)
	if err != nil {
		return nil, err
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

// lineProviderOffset resolves the byte offset for a finding's Position
// using the pre-built line offset index. Shared by insertion and
// replacement edit helpers below.
func lineProviderOffset(content []byte, idx []int, f finding.Finding) (int, error) {
	return resolveLineCol(idx, len(content), f.Position.Line, f.Position.Column)
}

func lineProviderInsertionEdit(content []byte, f finding.Finding, idx []int) ([]FixEdit, error) {
	offset, err := lineProviderOffset(content, idx, f)
	if err != nil {
		return nil, err
	}

	replacement := append([]byte(f.AfterCode), '\n')

	return []FixEdit{{Offset: offset, Length: 0, Replacement: replacement, Source: f}}, nil
}

func lineProviderReplacementEdit(content []byte, f finding.Finding, idx []int) ([]FixEdit, error) {
	offset, err := lineProviderOffset(content, idx, f)
	if err != nil {
		return nil, err
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
// disambiguation by position proximity.
//
// Disambiguation priority when multiple occurrences exist:
//  1. If Position.Line > 0 AND Position.Column > 0: the occurrence whose byte
//     offset is closest to the target line+column offset wins. This resolves
//     ambiguity between occurrences on the same line.
//  2. Else if Position.Line > 0: the occurrence on the closest line wins.
//  3. Otherwise: the first occurrence wins.
func (SubstringProvider) EditsWithLineIndex(content []byte, lineIndex []int, f finding.Finding) ([]FixEdit, error) {
	before := []byte(f.BeforeCode)

	occurrences := findAllOccurrences(content, before)
	if len(occurrences) == 0 {
		return nil, nil
	}

	if len(occurrences) == 1 {
		return []FixEdit{newReplacementEdit(occurrences[0], len(before), f)}, nil
	}

	best := pickNearestOccurrence(lineIndex, occurrences, f)

	return []FixEdit{newReplacementEdit(best, len(before), f)}, nil
}

// pickNearestOccurrence selects the occurrence closest to the finding's position.
// When both line and column are available, distance is measured in bytes from
// the target offset — this disambiguates occurrences on the same line.
// When only line is available, distance is measured in lines.
// When neither is available, the first occurrence is returned.
func pickNearestOccurrence(lineIndex, occurrences []int, f finding.Finding) int {
	best := occurrences[0]

	if f.Position.Line <= 0 || f.Position.Line > len(lineIndex) {
		return best
	}

	// When column is known, use byte-offset distance to the target point.
	// This naturally accounts for intra-line proximity.
	if f.Position.Column > 0 {
		target := lineIndex[f.Position.Line-1] + (f.Position.Column - 1)
		bestDist := absInt(occurrences[0] - target)

		for _, off := range occurrences[1:] {
			d := absInt(off - target)
			if d < bestDist {
				bestDist = d
				best = off
			}
		}

		return best
	}

	// Line-only fallback: minimize line distance.
	bestDist := offsetLineDistance(lineIndex, occurrences[0], f.Position.Line)

	for _, off := range occurrences[1:] {
		d := offsetLineDistance(lineIndex, off, f.Position.Line)
		if d < bestDist {
			bestDist = d
			best = off
		}
	}

	return best
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
