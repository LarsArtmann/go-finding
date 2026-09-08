package pipeline

import (
	"github.com/larsartmann/go-finding"
)

// FixOutcomeStatus classifies what happened to a single finding during fix
// application. Outcomes let callers distinguish applied fixes from silent
// no-ops (refused findings), conflicts, and provider failures.
//
// See docs/DOMAIN_LANGUAGE.md ("Fix Application") for the shared vocabulary.
type FixOutcomeStatus string

const (
	// FixOutcomeApplied means at least one edit for the finding was applied.
	FixOutcomeApplied FixOutcomeStatus = "applied"
	// FixOutcomeNoChange means the finding had no code change to apply
	// (HasCodeChange() is false).
	FixOutcomeNoChange FixOutcomeStatus = "no-change"
	// FixOutcomeRefused means every provider that could handle the finding
	// returned zero edits without an error — the finding matched, but the
	// provider chose not to edit.
	FixOutcomeRefused FixOutcomeStatus = "refused"
	// FixOutcomeConflict means the finding's edits overlapped another edit
	// and were skipped.
	FixOutcomeConflict FixOutcomeStatus = "conflict"
	// FixOutcomeInvalid means a provider produced edits, but they were
	// dropped as invalid or out of bounds without an error.
	FixOutcomeInvalid FixOutcomeStatus = "invalid"
	// FixOutcomeFailed means a provider returned an error while resolving
	// the finding to edits. Err carries the wrapped provider error.
	FixOutcomeFailed FixOutcomeStatus = "failed"
)

// FixOutcome is the per-finding result of fix application.
type FixOutcome struct {
	Finding finding.Finding
	Status  FixOutcomeStatus
	// Err is non-nil if and only if Status is FixOutcomeFailed.
	Err error
}

// FixApplyResult is the complete result of one FixEngine application.
// Outcomes contains exactly one entry per input finding, in input order.
type FixApplyResult struct {
	// Content is the input content after applying all non-conflicting edits.
	Content []byte
	// Applied lists findings whose edits were applied, in application order.
	Applied []finding.Finding
	// AppliedEdits lists the applied edits, in application order.
	AppliedEdits []FixEdit
	// Conflicts describes findings skipped because their edits overlapped
	// earlier edits.
	Conflicts []Conflict
	// Outcomes holds one entry per input finding, in input order.
	Outcomes []FixOutcome
	// Errors collects provider resolution errors. Each error also appears in
	// the corresponding outcome with Status FixOutcomeFailed.
	Errors []error
}

// HasErrors reports whether any provider error occurred.
func (r FixApplyResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// OutcomeFor returns the outcome recorded for the given finding ID, or nil
// if no such finding was part of the run.
func (r FixApplyResult) OutcomeFor(id finding.ID) *FixOutcome {
	for i := range r.Outcomes {
		if r.Outcomes[i].Finding.ID == id {
			return &r.Outcomes[i]
		}
	}

	return nil
}

// OutcomeCounts returns the number of outcomes per status.
func (r FixApplyResult) OutcomeCounts() map[FixOutcomeStatus]int {
	counts := make(map[FixOutcomeStatus]int)
	for _, o := range r.Outcomes {
		counts[o.Status]++
	}

	return counts
}
