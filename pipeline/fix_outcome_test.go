package pipeline

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

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

func TestFixOutcome_JSON_RoundTrip(t *testing.T) {
	g := NewParallelGomega(t)

	outcome := FixOutcome{
		Finding: makeFixFinding("1", "old()", "new()", "a.go", 2),
		Status:  FixOutcomeFailed,
		Err:     errors.New("resolve edits failed"),
	}

	data, err := json.Marshal(outcome, json.Deterministic(true))
	g.Expect(err).NotTo(HaveOccurred())

	again, err := json.Marshal(outcome, json.Deterministic(true))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(again)).To(Equal(string(data)))

	var parsed FixOutcome
	g.Expect(json.Unmarshal(data, &parsed)).To(Succeed())
	g.Expect(parsed.Finding.Equal(outcome.Finding)).To(BeTrue())
	g.Expect(parsed.Status).To(Equal(FixOutcomeFailed))
	g.Expect(parsed.Err).To(MatchError("resolve edits failed"))
}

func TestFixApplyResult_JSON_RoundTrip(t *testing.T) {
	g := NewParallelGomega(t)

	result := NewFixEngine().ApplyWithOutcomes(
		[]byte("old()\n"),
		[]finding.Finding{
			makeFixFinding("1", "old()", "new()", "a.go", 1),
			makeFixFinding("2", "never()", "x()", "a.go", 5),
		},
	)
	g.Expect(result.OutcomeCounts()[FixOutcomeApplied]).To(Equal(1))
	g.Expect(result.OutcomeCounts()[FixOutcomeFailed]).To(Equal(1))

	data, err := json.Marshal(result, json.Deterministic(true))
	g.Expect(err).NotTo(HaveOccurred())

	again, err := json.Marshal(result, json.Deterministic(true))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(again)).To(Equal(string(data)))

	var parsed FixApplyResult
	g.Expect(json.Unmarshal(data, &parsed)).To(Succeed())

	g.Expect(string(parsed.Content)).To(Equal("new()\n"))
	g.Expect(parsed.Applied).To(HaveLen(1))
	g.Expect(parsed.Outcomes).To(HaveLen(2))
	g.Expect(parsed.OutcomeCounts()).To(Equal(result.OutcomeCounts()))
	g.Expect(parsed.Errors).To(HaveLen(1))
	g.Expect(parsed.Errors[0].Error()).To(Equal(result.Errors[0].Error()))
}

// FuzzFixOutcomeUnmarshalJSON verifies that UnmarshalJSON never panics on
// malformed wire data, and that any outcome it accepts re-marshals cleanly.
func FuzzFixOutcomeUnmarshalJSON(f *testing.F) {
	f.Add([]byte(`{"finding":{"id":"a"},"status":"applied"}`))
	f.Add([]byte(`{"status":"failed","error":"boom"}`))
	f.Add([]byte(`{"status":42}`))
	f.Add([]byte(`{"finding":null,"status":""}`))
	f.Add([]byte(`not json at all`))
	f.Add([]byte(`{"finding":{"severity":"critical","line":-5},"status":"conflict"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var o FixOutcome
		if err := o.UnmarshalJSON(data); err != nil {
			return
		}

		// Accepted input must round-trip through the deterministic marshaler.
		remarshaled, err := o.MarshalJSON()
		if err != nil {
			t.Fatalf("accepted input re-marshal failed: %v\ninput: %s", err, data)
		}

		var again FixOutcome
		if err := again.UnmarshalJSON(remarshaled); err != nil {
			t.Fatalf("re-unmarshal of marshaled output failed: %v\ninput: %s", err, data)
		}
	})
}

// FuzzFixApplyResultUnmarshalJSON verifies that UnmarshalJSON never panics on
// malformed wire data, and that any result it accepts re-marshals cleanly.
func FuzzFixApplyResultUnmarshalJSON(f *testing.F) {
	f.Add([]byte(`{"outcomes":[{"finding":{"id":"a"},"status":"applied"}]}`))
	f.Add([]byte(`{"content":"aGVsbG8=","applied":[],"errors":["boom"]}`))
	f.Add([]byte(`{"appliedEdits":[{"offset":1,"length":2,"replacement":"eHg="}]}`))
	f.Add([]byte(`{"outcomes":[{"status":"failed","error":123}]}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`{"content":[1,2,3]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var r FixApplyResult
		if err := r.UnmarshalJSON(data); err != nil {
			return
		}

		remarshaled, err := r.MarshalJSON()
		if err != nil {
			t.Fatalf("accepted input re-marshal failed: %v\ninput: %s", err, data)
		}

		var again FixApplyResult
		if err := again.UnmarshalJSON(remarshaled); err != nil {
			t.Fatalf("re-unmarshal of marshaled output failed: %v\ninput: %s", err, data)
		}
	})
}

// TestFixApplyResult_GoldenWireBytes pins the exact JSON wire format of a
// representative FixApplyResult. This is the documented byte-level contract:
// field names, error-as-message-string, and content encoding must not drift
// silently. Update this golden only via an intentional, changelogged change.
func TestFixApplyResult_GoldenWireBytes(t *testing.T) {
	g := NewParallelGomega(t)

	f := finding.Finding{
		ID:          "gold-1",
		Rule:        "r1",
		ToolName:    "gold",
		Message:     "msg",
		Severity:    finding.SeverityWarning,
		Position:    finding.Position{File: "a.go", Line: 3, Offset: 10},
		BeforeCode:  "old",
		AfterCode:   "new",
		FixStrategy: finding.FixStrategyDirect,
	}

	res := FixApplyResult{
		Content: []byte("new content"),
		Applied: []finding.Finding{f},
		AppliedEdits: []FixEdit{{
			Offset:      0,
			Length:      3,
			Replacement: []byte("new"),
			Source:      f,
		}},
		Outcomes: []FixOutcome{
			{Finding: f, Status: FixOutcomeApplied},
			{Finding: f, Status: FixOutcomeFailed, Err: errors.New("provider boom")},
		},
		Errors: []error{errors.New("provider boom")},
	}

	data, err := res.MarshalJSON()
	g.Expect(err).NotTo(HaveOccurred())

	golden := `{"content":"bmV3IGNvbnRlbnQ=","applied":[{"id":"gold-1","rule":"r1","toolName":"gold","message":"msg","severity":"warning","position":{"file":"a.go","line":3,"column":0,"offset":10},"fixStrategy":"direct","beforeCode":"old","afterCode":"new","confidence":0}],"appliedEdits":[{"offset":0,"length":3,"replacement":"bmV3"}],"outcomes":[{"finding":{"id":"gold-1","rule":"r1","toolName":"gold","message":"msg","severity":"warning","position":{"file":"a.go","line":3,"column":0,"offset":10},"fixStrategy":"direct","beforeCode":"old","afterCode":"new","confidence":0},"status":"applied"},{"finding":{"id":"gold-1","rule":"r1","toolName":"gold","message":"msg","severity":"warning","position":{"file":"a.go","line":3,"column":0,"offset":10},"fixStrategy":"direct","beforeCode":"old","afterCode":"new","confidence":0},"status":"failed","error":"provider boom"}],"errors":["provider boom"]}`

	g.Expect(string(data)).To(Equal(golden), "wire format drifted from the golden bytes")

	var restored FixApplyResult
	g.Expect(restored.UnmarshalJSON(data)).To(Succeed())
	g.Expect(restored.Outcomes).To(HaveLen(2))
	g.Expect(restored.Outcomes[1].Status).To(Equal(FixOutcomeFailed))
	g.Expect(restored.Outcomes[1].Err.Error()).To(Equal("provider boom"))
}

// TestApplyWithOutcomes_Property_OutcomeCountsSum verifies the core outcome
// invariant: for any input set, the outcome counts sum to exactly the number
// of input findings — every finding gets precisely one outcome, whatever its
// status. Property-style: 100 randomized corpora from a fixed seed.
func TestApplyWithOutcomes_Property_OutcomeCountsSum(t *testing.T) {
	g := NewParallelGomega(t)

	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic property corpus

	content := []byte("package main\n\nold()\nold()\nold()\nold()\nold()\nold()\nold()\nold()\n")

	for iter := range 100 {
		n := rng.Intn(20)
		fixes := make([]finding.Finding, 0, n)

		for i := range n {
			var f finding.Finding

			switch rng.Intn(4) {
			case 0: // applies
				f = finding.Finding{
					ID:         finding.ID(fmt.Sprintf("f%d-%d-applied", iter, i)),
					BeforeCode: "old()",
					AfterCode:  "new()",
					Position:   finding.Pos("prop.go", 3+rng.Intn(8), 1),
					Range: &finding.Range{
						Start: finding.Position{File: "prop.go", Offset: 13},
						End:   finding.Position{File: "prop.go", Offset: 18},
					},
				}
			case 1: // refused (pattern absent)
				f = finding.Finding{
					ID:         finding.ID(fmt.Sprintf("f%d-%d-refused", iter, i)),
					BeforeCode: "absent()",
					AfterCode:  "new()",
					Position:   finding.Pos("prop.go", 3, 1),
				}
			case 2: // no change
				f = finding.Finding{
					ID:       finding.ID(fmt.Sprintf("f%d-%d-nochange", iter, i)),
					Message:  "report only",
					Position: finding.Pos("prop.go", 3, 1),
				}
			default: // no change, different shape
				f = finding.Finding{
					ID:       finding.ID(fmt.Sprintf("f%d-%d-plain", iter, i)),
					Rule:     "r",
					Position: finding.Pos("prop.go", 4, 1),
				}
			}

			fixes = append(fixes, f)
		}

		result := NewFixEngine().ApplyWithOutcomes(content, fixes)

		g.Expect(result.Outcomes).To(HaveLen(len(fixes)),
			"iter %d: one outcome per input finding", iter)

		sum := 0
		for _, count := range result.OutcomeCounts() {
			sum += count
		}

		g.Expect(sum).To(Equal(len(fixes)),
			"iter %d: outcome counts must sum to the number of input findings", iter)
	}
}
