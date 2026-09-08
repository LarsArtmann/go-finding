package pipeline

import (
	"errors"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

// erroringProvider fails resolution for findings whose rule matches.
type erroringProvider struct{}

func (erroringProvider) Name() string { return "erroring" }

func (erroringProvider) CanHandle(f finding.Finding) bool {
	return f.Rule == "unresolvable"
}

func (erroringProvider) Edits(_ []byte, _ finding.Finding) ([]FixEdit, error) {
	return nil, errors.New("cannot resolve this finding")
}

// TestFixEngine_ApplyWithOutcomes_MixedStatuses verifies that per-finding
// outcomes distinguish applied, no-change, refused, conflict, and failed
// findings (issue #27).
func TestFixEngine_ApplyWithOutcomes_MixedStatuses(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("alpha\nbeta\ngamma\n")

	fixes := []finding.Finding{
		makeFixFinding("applied", "alpha", "ALPHA", "", 0),
		{ID: "no-change", Rule: "r", ToolName: "tool", Message: "m", Position: finding.Position{Line: 1}},
		makeFixFinding("refused", "missing-text", "NEW", "", 0),
		makeFixFinding("failed", "missing-text", "BETA", "", 0),
		makeFixFinding("conflict-a", "gamma", "GAMMA-A", "", 0),
		makeFixFinding("conflict-b", "gamma", "GAMMA-B", "", 0),
	}

	fixes[3].Rule = "unresolvable"

	// conflict-a and conflict-b target the same text; the engine's default
	// providers resolve both to edits at the same offset, so one applies and
	// the other conflicts. The erroring provider is tried first but only
	// claims findings with rule "unresolvable".
	engine := NewFixEngineWithProviders(erroringProvider{}, &OffsetProvider{}, &LineProvider{}, &SubstringProvider{})
	result := engine.ApplyWithOutcomes(content, fixes)

	statuses := map[string]FixOutcomeStatus{}
	for _, o := range result.Outcomes {
		statuses[string(o.Finding.ID)] = o.Status
	}

	g.Expect(statuses).To(Equal(map[string]FixOutcomeStatus{
		"applied":    FixOutcomeApplied,
		"no-change":  FixOutcomeNoChange,
		"refused":    FixOutcomeRefused,
		"failed":     FixOutcomeFailed,
		"conflict-a": FixOutcomeApplied,
		"conflict-b": FixOutcomeConflict,
	}))

	g.Expect(result.HasErrors()).To(BeTrue())
	g.Expect(result.Errors).To(HaveLen(1))
	g.Expect(result.Errors[0].Error()).To(ContainSubstring("cannot resolve this finding"))

	g.Expect(result.Content).To(Equal([]byte("ALPHA\nbeta\nGAMMA-A\n")))

	outcome := result.OutcomeFor("failed")
	g.Expect(outcome).NotTo(BeNil())
	g.Expect(outcome.Err).To(HaveOccurred())
}

// TestFixEngine_ApplyWithOutcomes_Empty verifies the zero-input contract:
// content passes through unchanged and no outcomes are recorded.
func TestFixEngine_ApplyWithOutcomes_Empty(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("unchanged\n")

	engine := NewFixEngine()
	result := engine.ApplyWithOutcomes(content, nil)

	g.Expect(result.Content).To(Equal(content))
	g.Expect(result.Outcomes).To(BeEmpty())
	g.Expect(result.Applied).To(BeEmpty())
	g.Expect(result.HasErrors()).To(BeFalse())
	g.Expect(result.OutcomeCounts()).To(BeEmpty())
	g.Expect(result.OutcomeFor("x")).To(BeNil())
}

// TestFixEngine_ApplyWithConflicts_LegacyParity verifies that the legacy
// ApplyWithConflicts return shape is preserved when all providers fail:
// resolve errors are returned even though no edits were produced.
func TestFixEngine_ApplyWithConflicts_LegacyParity(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("alpha\n")

	fixes := []finding.Finding{makeFixFinding("bad", "alpha", "ALPHA", "", 0)}
	fixes[0].Rule = "unresolvable"

	engine := NewFixEngineWithProviders(erroringProvider{})
	applied, edits, conflicts, result, errs := engine.ApplyWithConflicts(content, fixes)

	g.Expect(applied).To(BeEmpty())
	g.Expect(edits).To(BeEmpty())
	g.Expect(conflicts).To(BeEmpty())
	g.Expect(result).To(Equal(content))
	g.Expect(errs).To(HaveLen(1))
}

// TestFixOutcome_StatusValues verifies the outcome status strings stay stable
// (consumers may persist or log them).
func TestFixOutcome_StatusValues(t *testing.T) {
	g := NewParallelGomega(t)

	g.Expect(string(FixOutcomeApplied)).To(Equal("applied"))
	g.Expect(string(FixOutcomeNoChange)).To(Equal("no-change"))
	g.Expect(string(FixOutcomeRefused)).To(Equal("refused"))
	g.Expect(string(FixOutcomeConflict)).To(Equal("conflict"))
	g.Expect(string(FixOutcomeInvalid)).To(Equal("invalid"))
	g.Expect(string(FixOutcomeFailed)).To(Equal("failed"))
}

// TestFixEngine_ApplyWithOutcomes_InvalidEditReported verifies that a provider
// producing an out-of-bounds edit yields an invalid outcome instead of a
// silent drop.
func TestFixEngine_ApplyWithOutcomes_InvalidEditReported(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("short")

	outOfBounds := newReplacementEdit(100, 5, makeFixFinding("oob", "short", "longer", "", 0))
	engine := NewFixEngineWithProviders(staticEditProvider{edits: []FixEdit{outOfBounds}})

	result := engine.ApplyWithOutcomes(content, []finding.Finding{makeFixFinding("oob", "short", "longer", "", 0)})

	g.Expect(result.Outcomes).To(HaveLen(1))
	g.Expect(result.Outcomes[0].Status).To(Equal(FixOutcomeInvalid))
	g.Expect(result.Content).To(Equal(content))
	g.Expect(result.HasErrors()).To(BeFalse())
}

// staticEditProvider always returns the given edits.
type staticEditProvider struct {
	edits []FixEdit
}

func (staticEditProvider) Name() string { return "static" }

func (staticEditProvider) CanHandle(finding.Finding) bool { return true }

func (s staticEditProvider) Edits(_ []byte, _ finding.Finding) ([]FixEdit, error) {
	return s.edits, nil
}

// TestFixApplyResult_OutcomeCounts verifies counting across statuses.
func TestFixApplyResult_OutcomeCounts(t *testing.T) {
	g := NewParallelGomega(t)

	result := FixApplyResult{
		Outcomes: []FixOutcome{
			{Status: FixOutcomeApplied},
			{Status: FixOutcomeApplied},
			{Status: FixOutcomeRefused},
			{Status: FixOutcomeFailed, Err: errors.New("x")},
		},
	}

	counts := result.OutcomeCounts()
	g.Expect(counts[FixOutcomeApplied]).To(Equal(2))
	g.Expect(counts[FixOutcomeRefused]).To(Equal(1))
	g.Expect(counts[FixOutcomeFailed]).To(Equal(1))
}
