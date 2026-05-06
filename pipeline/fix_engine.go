package pipeline

import (
	"cmp"
	"slices"

	"github.com/larsartmann/go-finding"
)

// FixEngine applies byte-level edits to file content.
// It delegates to registered FixProviders to convert findings into edits,
// then applies them deterministically in descending offset order so that
// earlier edits don't shift the byte positions of later ones.
type FixEngine struct {
	providers []FixProvider
}

// NewFixEngine creates an engine with the default text-based providers:
// OffsetProvider (byte offsets), LineProvider (line/column), SubstringProvider (fallback).
func NewFixEngine() *FixEngine {
	return &FixEngine{
		providers: []FixProvider{
			&OffsetProvider{},
			&LineProvider{},
			&SubstringProvider{},
		},
	}
}

// NewFixEngineWithProviders creates an engine with custom providers.
// Providers are tried in order; the first that CanHandle a finding is used.
func NewFixEngineWithProviders(providers ...FixProvider) *FixEngine {
	return &FixEngine{providers: providers}
}

// Providers returns a copy of the registered providers list.
func (e *FixEngine) Providers() []FixProvider {
	return slices.Clone(e.providers)
}

// Apply applies findings to content and returns the modified content,
// the successfully applied findings, and the count of applied fixes.
func (e *FixEngine) Apply(
	content []byte,
	fixes []finding.Finding,
) ([]byte, []finding.Finding, int) {
	if len(fixes) == 0 {
		return content, nil, 0
	}

	var allEdits []FixEdit

	for _, f := range fixes {
		if f.BeforeCode == "" && f.AfterCode == "" {
			continue
		}

		edits := e.resolveEdits(content, f)
		allEdits = append(allEdits, edits...)
	}

	if len(allEdits) == 0 {
		return content, nil, 0
	}

	// Sort descending by offset so later edits don't shift earlier ones.
	slices.SortFunc(allEdits, func(a, b FixEdit) int {
		return cmp.Compare(b.Offset, a.Offset)
	})

	return e.applyEdits(content, allEdits)
}

// resolveEdits tries each provider in order and returns edits from the first match.
func (e *FixEngine) resolveEdits(content []byte, f finding.Finding) []FixEdit {
	for _, p := range e.providers {
		if !p.CanHandle(f) {
			continue
		}

		edits, err := p.Edits(content, f)
		if err != nil || len(edits) == 0 {
			continue
		}

		return edits
	}

	return nil
}

// applyEdits applies non-overlapping edits to content, returning the result,
// the applied findings, and the count.
func (*FixEngine) applyEdits(
	content []byte,
	edits []FixEdit,
) ([]byte, []finding.Finding, int) {
	var applied []finding.Finding
	result := content
	frontier := len(content) + 1

	for _, edit := range edits {
		if err := edit.Validate(); err != nil {
			continue
		}

		if edit.EndOffset() > len(result) {
			continue
		}

		if edit.EndOffset() > frontier {
			continue
		}

		var buf []byte
		buf = append(buf, result[:edit.Offset]...)
		buf = append(buf, edit.Replacement...)
		buf = append(buf, result[edit.EndOffset():]...)
		result = buf
		frontier = edit.Offset

		applied = append(applied, edit.Source)
	}

	return result, applied, len(applied)
}
