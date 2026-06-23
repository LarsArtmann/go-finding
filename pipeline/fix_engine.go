package pipeline

import (
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
// Provider errors from edit resolution are discarded; use ApplyWithConflicts
// to access them.
func (e *FixEngine) Apply(
	content []byte,
	fixes []finding.Finding,
) ([]byte, []finding.Finding, int) {
	applied, _, conflicts, result, _ := e.ApplyWithConflicts(content, fixes)
	_ = conflicts

	return result, applied, len(applied)
}

// ApplyWithConflicts applies findings and returns applied findings, applied edits,
// conflicts, modified content, and any provider errors encountered during edit resolution.
// Conflicts are findings whose edits overlap with earlier edits — they are skipped.
func (e *FixEngine) ApplyWithConflicts(
	content []byte,
	fixes []finding.Finding,
) ([]finding.Finding, []FixEdit, []Conflict, []byte, []error) {
	if len(fixes) == 0 {
		return nil, nil, nil, content, nil
	}

	var (
		allEdits      []FixEdit
		resolveErrors []error
		lineIndex     []int // lazily built by resolveEdits when a lineIndexAware provider handles a finding
	)

	for _, f := range fixes {
		if !f.HasCodeChange() {
			continue
		}

		edits, err := e.resolveEdits(content, &lineIndex, f)
		if err != nil {
			resolveErrors = append(resolveErrors, err)
		}

		allEdits = append(allEdits, edits...)
	}

	if len(allEdits) == 0 {
		return nil, nil, nil, content, resolveErrors
	}

	// Sort descending by offset so later edits don't shift earlier ones.
	sortEditsDescending(allEdits)

	applied, appliedEdits, conflicts, result := e.applyEditsWithConflicts(content, allEdits)

	return applied, appliedEdits, conflicts, result, resolveErrors
}

// resolveEdits tries each provider in order and returns edits from the first match.
// If a provider that CanHandle'd the finding returns an error, it is collected.
// Returns the edits from the first successful provider, or the first provider error if all fail.
// If the provider implements lineIndexAware, the line offset index is lazily built
// on first access and reused for subsequent findings, avoiding O(n) rebuilds per finding.
func (e *FixEngine) resolveEdits(content []byte, lineIndex *[]int, f finding.Finding) ([]FixEdit, error) {
	var firstErr error

	for _, p := range e.providers {
		if !p.CanHandle(f) {
			continue
		}

		var edits []FixEdit

		var err error

		if la, ok := p.(lineIndexAware); ok {
			if *lineIndex == nil {
				*lineIndex = buildLineOffsetIndex(content)
			}

			edits, err = la.EditsWithLineIndex(content, *lineIndex, f)
		} else {
			edits, err = p.Edits(content, f)
		}

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
// Edits must be sorted descending by offset (highest first).
func (*FixEngine) applyEditsWithConflicts(
	content []byte,
	edits []FixEdit,
) ([]finding.Finding, []FixEdit, []Conflict, []byte) {
	var (
		applied      []finding.Finding
		appliedEdits []FixEdit
		conflicts    []Conflict
	)

	frontier := len(content) + 1

	// Phase 1: Walk edits in descending offset order, detecting conflicts
	// via the frontier boundary. Non-conflicting edits are collected for
	// a single-pass application in Phase 2.
	for _, edit := range edits {
		err := edit.Validate()
		if err != nil {
			continue
		}

		if edit.EndOffset() > len(content) {
			continue
		}

		if edit.EndOffset() > frontier {
			var conflictsWith []finding.Finding

			for _, prev := range appliedEdits {
				if edit.Overlaps(prev) {
					conflictsWith = append(conflictsWith, prev.Source)
				}
			}

			conflicts = append(conflicts, Conflict{
				Finding:       edit.Source,
				ConflictsWith: conflictsWith,
				Reason:        ReasonOverlappingEdit,
			})

			continue
		}

		applied = append(applied, edit.Source)
		appliedEdits = append(appliedEdits, edit)
		frontier = edit.Offset
	}

	// Phase 2: Apply all non-conflicting edits in a single buffer pass.
	// This is O(F + R) where F = file size and R = total replacement size,
	// instead of the previous O(N × F) which copied the entire content
	// per edit.
	result := applyEditsToContent(content, appliedEdits)

	return applied, appliedEdits, conflicts, result
}

// applyEditsToContent applies non-overlapping edits to content in a single pass.
// Edits must be sorted descending by offset (highest first) and should be
// non-conflicting (verified by the caller). Overlapping edits are skipped
// defensively to prevent panics.
func applyEditsToContent(content []byte, edits []FixEdit) []byte {
	if len(edits) == 0 {
		return content
	}

	// Pre-compute the final size to avoid reallocation.
	// This is an upper bound; skipped overlapping edits reduce actual size.
	finalSize := len(content)
	for _, edit := range edits {
		finalSize += len(edit.Replacement) - edit.Length
	}

	if finalSize < 0 {
		finalSize = 0
	}

	result := make([]byte, 0, finalSize)

	// Edits are sorted DESCENDING by offset. Iterate in REVERSE (ascending)
	// to build the output left-to-right in a single pass.
	prevEnd := 0

	for i := range slices.Backward(edits) {
		edit := edits[i]

		// Skip overlapping edits defensively (shouldn't happen in normal use).
		if edit.Offset < prevEnd {
			continue
		}

		result = append(result, content[prevEnd:edit.Offset]...)
		result = append(result, edit.Replacement...)
		prevEnd = edit.EndOffset()
	}

	result = append(result, content[prevEnd:]...)

	return result
}
