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
	result := e.apply(content, fixes, false)

	return result.Content, result.Applied, len(result.Applied)
}

// ApplyWithConflicts applies findings and returns applied findings, applied edits,
// conflicts, modified content, and any provider errors encountered during edit resolution.
// Conflicts are findings whose edits overlap with earlier edits — they are skipped.
func (e *FixEngine) ApplyWithConflicts(
	content []byte,
	fixes []finding.Finding,
) ([]finding.Finding, []FixEdit, []Conflict, []byte, []error) {
	result := e.apply(content, fixes, false)

	return result.Applied, result.AppliedEdits, result.Conflicts, result.Content, result.Errors
}

// ApplyWithOutcomes applies findings and returns the full per-finding result:
// applied findings, applied edits, conflicts, modified content, provider
// errors, and one FixOutcome per input finding (in input order) that
// distinguishes applied / no-change / refused / conflict / invalid / failed.
func (e *FixEngine) ApplyWithOutcomes(
	content []byte,
	fixes []finding.Finding,
) FixApplyResult {
	return e.apply(content, fixes, true)
}

// apply is the single implementation behind Apply, ApplyWithConflicts, and
// ApplyWithOutcomes. wantOutcomes controls per-finding outcome bookkeeping:
// the legacy entry points discard outcomes, so they skip allocating the
// outcomes slice (one full Finding copy per input) and the reconciliation
// pass, keeping their pre-outcome allocation profile.
func (e *FixEngine) apply(content []byte, fixes []finding.Finding, wantOutcomes bool) FixApplyResult {
	result := FixApplyResult{Content: content}
	if len(fixes) == 0 {
		return result
	}

	var (
		allEdits  []FixEdit
		lineIndex []int // lazily built by resolveEdits when a lineIndexAware provider handles a finding
		resolved  []int // outcome indices whose edits were collected for application
	)

	if wantOutcomes {
		result.Outcomes = make([]FixOutcome, 0, len(fixes))
	}

	addOutcome := func(f finding.Finding, status FixOutcomeStatus, err error) {
		if !wantOutcomes {
			return
		}

		result.Outcomes = append(result.Outcomes, FixOutcome{
			Finding: f,
			Status:  status,
			Err:     err,
		})
	}

	for _, f := range fixes {
		if !f.HasCodeChange() {
			addOutcome(f, FixOutcomeNoChange, nil)

			continue
		}

		edits, err := e.resolveEdits(content, &lineIndex, f)
		if err != nil {
			wrapped := fmt.Errorf("finding %s: %w", f.ID, err)
			result.Errors = append(result.Errors, wrapped)
			addOutcome(f, FixOutcomeFailed, wrapped)

			continue
		}

		if len(edits) == 0 {
			addOutcome(f, FixOutcomeRefused, nil)

			continue
		}

		if wantOutcomes {
			resolved = append(resolved, len(result.Outcomes))
		}

		addOutcome(f, FixOutcomeApplied, nil)
		allEdits = append(allEdits, edits...)
	}

	if len(allEdits) == 0 {
		return result
	}

	// Sort descending by offset so later edits don't shift earlier ones.
	sortEditsDescending(allEdits)

	result.Applied, result.AppliedEdits, result.Conflicts, result.Content = e.applyEditsWithConflicts(content, allEdits)

	if wantOutcomes {
		reconcileOutcomes(&result, resolved)
	}

	return result
}

// reconcileOutcomes finalizes provisional FixOutcomeApplied statuses against
// the actual application results: findings whose edits survived are applied,
// findings in Conflicts are conflicts, and findings whose edits were dropped
// as invalid or out of bounds are invalid.
func reconcileOutcomes(result *FixApplyResult, resolved []int) {
	appliedIDs := make(map[finding.ID]struct{}, len(result.Applied))
	for _, f := range result.Applied {
		appliedIDs[f.ID] = struct{}{}
	}

	conflictIDs := make(map[finding.ID]struct{}, len(result.Conflicts))
	for _, c := range result.Conflicts {
		conflictIDs[c.Finding.ID] = struct{}{}
	}

	for _, idx := range resolved {
		o := &result.Outcomes[idx]

		switch {
		case hasID(appliedIDs, o.Finding.ID):
			o.Status = FixOutcomeApplied
		case hasID(conflictIDs, o.Finding.ID):
			o.Status = FixOutcomeConflict
		default:
			o.Status = FixOutcomeInvalid
		}
	}
}

func hasID(set map[finding.ID]struct{}, id finding.ID) bool {
	_, ok := set[id]

	return ok
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
