package pipeline

import (
	"cmp"
	"fmt"
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
	applied, _, result, _ := e.ApplyWithConflicts(content, fixes)

	return result, applied, len(applied)
}

// ApplyWithConflicts applies findings and returns applied findings, conflicts, modified content,
// and any provider errors encountered during edit resolution.
// Conflicts are findings whose edits overlap with earlier edits — they are skipped.
func (e *FixEngine) ApplyWithConflicts(
	content []byte,
	fixes []finding.Finding,
) ([]finding.Finding, []ConflictInfo, []byte, []error) {
	if len(fixes) == 0 {
		return nil, nil, content, nil
	}

	var allEdits []FixEdit
	var resolveErrors []error

	for _, f := range fixes {
		if !f.HasCodeChange() {
			continue
		}

		edits, err := e.resolveEdits(content, f)
		if err != nil {
			resolveErrors = append(resolveErrors, err)
		}
		allEdits = append(allEdits, edits...)
	}

	if len(allEdits) == 0 {
		return nil, nil, content, resolveErrors
	}

	// Sort descending by offset so later edits don't shift earlier ones.
	slices.SortFunc(allEdits, func(a, b FixEdit) int {
		return cmp.Compare(b.Offset, a.Offset)
	})

	applied, conflicts, result := e.applyEditsWithConflicts(content, allEdits)

	return applied, conflicts, result, resolveErrors
}

// resolveEdits tries each provider in order and returns edits from the first match.
// If a provider that CanHandle'd the finding returns an error, it is collected.
// Returns the edits from the first successful provider, or the first provider error if all fail.
func (e *FixEngine) resolveEdits(content []byte, f finding.Finding) ([]FixEdit, error) {
	var firstErr error

	for _, p := range e.providers {
		if !p.CanHandle(f) {
			continue
		}

		edits, err := p.Edits(content, f)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("provider %s: %w", p.Name(), err)
			}

			continue
		}

		if len(edits) > 0 {
			return edits, nil
		}
	}

	return nil, firstErr
}

// applyEditsWithConflicts applies edits and tracks which were skipped due to overlaps.
func (*FixEngine) applyEditsWithConflicts(
	content []byte,
	edits []FixEdit,
) ([]finding.Finding, []ConflictInfo, []byte) {
	var applied []finding.Finding
	var appliedEdits []FixEdit
	var conflicts []ConflictInfo
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
			var conflictsWith []finding.Finding
			for _, prev := range appliedEdits {
				if edit.Overlaps(prev) {
					conflictsWith = append(conflictsWith, prev.Source)
				}
			}

			conflicts = append(conflicts, ConflictInfo{
				Finding:       edit.Source,
				ConflictsWith: conflictsWith,
				Reason:        ReasonOverlappingEdit,
			})

			continue
		}

		var buf []byte
		buf = append(buf, result[:edit.Offset]...)
		buf = append(buf, edit.Replacement...)
		buf = append(buf, result[edit.EndOffset():]...)
		result = buf
		frontier = edit.Offset

		applied = append(applied, edit.Source)
		appliedEdits = append(appliedEdits, edit)
	}

	return applied, conflicts, result
}
