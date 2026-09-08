package pipeline

import (
	"encoding/json/v2"
	"errors"
	"fmt"

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

// fixOutcomeJSON is the JSON wire type for FixOutcome. Controlling the wire
// type keeps field tags independent of the domain type and carries the error
// as a message string (error values do not serialize).
type fixOutcomeJSON struct {
	Finding finding.Finding  `json:"finding"`
	Status  FixOutcomeStatus `json:"status"`
	Error   string           `json:"error,omitempty"`
}

// MarshalJSON implements json.Marshaler for FixOutcome. Output is
// deterministic (see scripts/json-deterministic-check.sh).
func (o FixOutcome) MarshalJSON() ([]byte, error) { //nolint:recvcheck // value receivers for read-only, pointer for UnmarshalJSON
	wire := fixOutcomeJSON{Finding: o.Finding, Status: o.Status}
	if o.Err != nil {
		wire.Error = o.Err.Error()
	}

	return json.Marshal(wire, json.Deterministic(true))
}

// UnmarshalJSON implements json.Unmarshaler for FixOutcome. The error string
// is restored as a plain error; error identity does not survive
// serialization, so match on the original error before marshaling.
func (o *FixOutcome) UnmarshalJSON(data []byte) error {
	var wire fixOutcomeJSON
	if err := json.Unmarshal(data, &wire); err != nil {
		return fmt.Errorf("unmarshal FixOutcome: %w", err)
	}

	o.Finding = wire.Finding
	o.Status = wire.Status

	if wire.Error != "" {
		o.Err = errors.New(wire.Error)
	}

	return nil
}

// fixApplyResultJSON is the JSON wire type for FixApplyResult.
type fixApplyResultJSON struct {
	// Content is the file content after applying all non-conflicting edits
	// (base64 in JSON).
	Content      []byte            `json:"content,omitempty"`
	Applied      []finding.Finding `json:"applied,omitempty"`
	AppliedEdits []FixEdit         `json:"appliedEdits,omitempty"`
	Conflicts    []Conflict        `json:"conflicts,omitempty"`
	Outcomes     []FixOutcome      `json:"outcomes,omitempty"`
	Errors       []string          `json:"errors,omitempty"`
}

// MarshalJSON implements json.Marshaler for FixApplyResult. Output is
// deterministic; provider errors are serialized as message strings.
func (r FixApplyResult) MarshalJSON() ([]byte, error) { //nolint:recvcheck // value receivers for read-only, pointer for UnmarshalJSON
	wire := fixApplyResultJSON{
		Content:      r.Content,
		Applied:      r.Applied,
		AppliedEdits: r.AppliedEdits,
		Conflicts:    r.Conflicts,
		Outcomes:     r.Outcomes,
	}

	for _, err := range r.Errors {
		if err != nil {
			wire.Errors = append(wire.Errors, err.Error())
		}
	}

	return json.Marshal(wire, json.Deterministic(true))
}

// UnmarshalJSON implements json.Unmarshaler for FixApplyResult. Errors are
// restored as plain errors carrying the original message text.
func (r *FixApplyResult) UnmarshalJSON(data []byte) error {
	var wire fixApplyResultJSON
	if err := json.Unmarshal(data, &wire); err != nil {
		return fmt.Errorf("unmarshal FixApplyResult: %w", err)
	}

	*r = FixApplyResult{
		Content:      wire.Content,
		Applied:      wire.Applied,
		AppliedEdits: wire.AppliedEdits,
		Conflicts:    wire.Conflicts,
		Outcomes:     wire.Outcomes,
	}

	for _, msg := range wire.Errors {
		r.Errors = append(r.Errors, errors.New(msg))
	}

	return nil
}
