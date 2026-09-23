// Package goast provides a Go AST-aware [pipeline.FixProvider] for .go source files.
//
// It uses [go/parser] and [go/ast] to produce precise byte-level edits,
// eliminating the ambiguity of substring matching when the same text
// appears multiple times in a file.
//
// Register it before the text-based providers for maximum accuracy:
//
//	cfg := pipeline.Config{
//	    FixProviders: []pipeline.FixProvider{
//	        &pipeline.OffsetProvider{},   // exact byte offsets (no parse needed)
//	        &goast.Provider{},            // AST-aware for .go files
//	        &pipeline.LineProvider{},     // line/column for non-Go files
//	        &pipeline.SubstringProvider{}, // last-resort substring fallback
//	    },
//	}
//
// The parsed AST is cached per unique file content (identified by FNV-1a hash),
// so processing N findings in the same file incurs only a single parse.
package goast

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"strings"
	"sync"
	"unsafe"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/gotoken"
	"github.com/larsartmann/go-finding/lockutil"
	"github.com/larsartmann/go-finding/pipeline"
)

// Provider is a Go AST-aware [pipeline.FixProvider].
//
// It outperforms [pipeline.SubstringProvider] for Go files by using AST
// structure to disambiguate which occurrence of BeforeCode to replace.
// When a finding has a line/column position, the provider locates the
// innermost AST node at that position and searches for BeforeCode within
// that node's byte range — not the entire file.
type Provider struct {
	mu    sync.Mutex
	cache parseCache
}

type parseCache struct {
	hash     uint64
	fset     *token.FileSet
	file     *ast.File
	contentP unsafe.Pointer // pointer identity for fast cache hit without hashing
}

// Name returns the provider name used for identification and CLI registration.
func (*Provider) Name() string { return "go-ast" }

// CanHandle reports whether the finding targets a .go file and has a code change.
func (*Provider) CanHandle(f finding.Finding) bool {
	return f.HasCodeChange() && strings.HasSuffix(string(f.Position.File), ".go")
}

// Compile-time interface assertion — catches missing methods at build time.
var _ pipeline.FixProvider = (*Provider)(nil)

// Edits converts a finding into byte-level edits using the parsed Go AST.
func (p *Provider) Edits(content []byte, f finding.Finding) ([]pipeline.FixEdit, error) {
	fset, file, ok := p.parse(content, string(f.Position.File))
	if !ok {
		return nil, nil
	}

	before := []byte(f.BeforeCode)

	switch {
	case f.Range != nil && f.Range.Length() > 0:
		return offsetRangeEdits(content, f, before)

	case f.Position.Line > 0:
		return p.positionEdits(content, fset, file, f, before)

	case f.BeforeCode != "":
		return beforeCodeOnlyEdits(content, f, before)

	default:
		return nil, nil
	}
}

// parse returns the cached AST for the given content, parsing on first access.
// Returns ok=false (not an error) when the file cannot be parsed — the engine
// falls through to the next provider (e.g., SubstringProvider).
func (p *Provider) parse(content []byte, filename string) (*token.FileSet, *ast.File, bool) {
	// Fast path: check pointer identity first to avoid hashing the entire
	// content on every call. This is the common case when processing N
	// findings in the same file within a single ApplyWithConflicts call.
	contentPtr := unsafe.Pointer(unsafe.SliceData(content)) //nolint:gosec // cache key, never dereferenced

	type result struct {
		fset *token.FileSet
		file *ast.File
		ok   bool
	}

	res := lockutil.Locked(&p.mu, func() result {
		if p.cache.contentP == contentPtr && p.cache.file != nil {
			return result{fset: p.cache.fset, file: p.cache.file, ok: true}
		}

		// Content changed (different pointer): verify with hash.
		h := fnv.New64a()
		_, _ = h.Write(content) //nolint:erraudit // hash.Write is documented to never return an error
		hash := h.Sum64()

		if p.cache.hash == hash && p.cache.file != nil {
			p.cache.contentP = contentPtr

			return result{fset: p.cache.fset, file: p.cache.file, ok: true}
		}

		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
		if err != nil || file == nil {
			p.cache = parseCache{} //nolint:exhaustruct_v5 // intentionally zero: reset to avoid re-attempting known-bad content

			return result{fset: nil, file: nil, ok: false}
		}

		p.cache = parseCache{hash: hash, fset: fset, file: file, contentP: contentPtr}

		return result{fset: fset, file: file, ok: true}
	})

	return res.fset, res.file, res.ok
}

// offsetRangeEdits handles findings with explicit byte-offset ranges.
func offsetRangeEdits(content []byte, f finding.Finding, before []byte) ([]pipeline.FixEdit, error) {
	start := f.Range.Start.Offset
	end := f.Range.End.Offset

	if end > len(content) {
		return nil, nil
	}

	if len(before) > 0 && !bytes.Equal(content[start:end], before) {
		return nil, nil
	}

	return []pipeline.FixEdit{newEdit(start, end-start, f)}, nil
}

// positionEdits handles findings with line/column information.
// It converts the position to a byte offset via the token.FileSet, then:
//   - For replacements (BeforeCode set): tries exact match at the offset,
//     falls back to searching within the innermost AST node.
//   - For insertions (AfterCode set, no BeforeCode): inserts at the offset.
func (*Provider) positionEdits(
	content []byte,
	fset *token.FileSet,
	file *ast.File,
	f finding.Finding,
	before []byte,
) ([]pipeline.FixEdit, error) {
	tokenFile := fset.File(file.Pos())
	if tokenFile == nil {
		return nil, nil
	}

	offset, ok := gotoken.LineColToOffset(tokenFile, f.Position.Line, f.Position.Column, len(content))
	if !ok {
		return nil, nil
	}

	if len(before) > 0 {
		if edit := matchAtOffset(content, offset, before, f); edit != nil {
			return []pipeline.FixEdit{*edit}, nil
		}

		node := gotoken.FindInnermostNode(fset, file, offset)
		if node != nil {
			if edit := matchInNode(content, fset, node, before, f); edit != nil {
				return []pipeline.FixEdit{*edit}, nil
			}
		}

		return nil, nil
	}

	if f.AfterCode != "" {
		return []pipeline.FixEdit{{
			Offset:      offset,
			Length:      0,
			Replacement: append([]byte(f.AfterCode), '\n'),
			Source:      f,
		}}, nil
	}

	return nil, nil
}

// beforeCodeOnlyEdits handles findings with BeforeCode but no position info.
// It searches the entire content (same as SubstringProvider without line proximity).
func beforeCodeOnlyEdits(content []byte, f finding.Finding, before []byte) ([]pipeline.FixEdit, error) {
	idx := bytes.Index(content, before)
	if idx < 0 {
		return nil, nil
	}

	return []pipeline.FixEdit{newEdit(idx, len(before), f)}, nil
}

// matchAtOffset checks whether before matches content at exactly offset.
func matchAtOffset(content []byte, offset int, before []byte, f finding.Finding) *pipeline.FixEdit {
	end := offset + len(before)
	if end > len(content) {
		return nil
	}

	if !bytes.Equal(content[offset:end], before) {
		return nil
	}

	edit := newEdit(offset, len(before), f)

	return &edit
}

// matchInNode searches for before within the byte range of the given AST node.
func matchInNode(
	content []byte,
	fset *token.FileSet,
	node ast.Node,
	before []byte,
	f finding.Finding,
) *pipeline.FixEdit {
	start, end, ok := gotoken.NodeByteRange(fset, node)
	if !ok || end > len(content) || start > end {
		return nil
	}

	relIdx := bytes.Index(content[start:end], before)
	if relIdx < 0 {
		return nil
	}

	edit := newEdit(start+relIdx, len(before), f)

	return &edit
}

func newEdit(offset, length int, f finding.Finding) pipeline.FixEdit {
	return pipeline.FixEdit{
		Offset:      offset,
		Length:      length,
		Replacement: []byte(f.AfterCode),
		Source:      f,
	}
}

var _ pipeline.FixProvider = (*Provider)(nil)
