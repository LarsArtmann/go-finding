package finding

import (
	"context"
	"testing"
)

// TestSplitBrain_FixStrategyNormalization verifies that empty FixStrategy
// is normalized to FixStrategyNone, resolving split-brain #2.
func TestSplitBrain_FixStrategyNormalization(t *testing.T) {
	t.Parallel()

	// Builder normalizes empty strategy to "none"
	f, err := NewBuilder("r", "test", "m", SeverityError, Pos("main.go", 1, 1)).
		WithFixStrategy("").
		Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	if f.FixStrategy != FixStrategyNone {
		t.Errorf("after Build(), FixStrategy = %q, want %q", f.FixStrategy, FixStrategyNone)
	}

	// Empty and "none" should be Equal (Equal normalizes both sides)
	empty := Finding{FixStrategy: ""}

	none := Finding{FixStrategy: FixStrategyNone}
	if !empty.Equal(none) {
		t.Error("Finding{Strategy:\"\"}.Equal(Finding{Strategy:\"none\"}) = false, want true")
	}
}

// TestSplitBrain_HasFixRespectsValidate verifies that HasFix() never returns
// true for findings that would fail validation, resolving split-brain #3.
func TestSplitBrain_HasFixRespectsValidate(t *testing.T) {
	t.Parallel()

	// Direct without code: invalid per Validate, should not claim HasFix
	directNoCode := Finding{FixStrategy: FixStrategyDirect}
	if directNoCode.HasFix() {
		t.Error("FixStrategyDirect without code: HasFix() = true, want false")
	}

	if directNoCode.Validate() == nil {
		t.Error("FixStrategyDirect without code: Validate() should fail")
	}

	// Direct with code: valid, HasFix should be true
	directWithCode := Finding{FixStrategy: FixStrategyDirect, AfterCode: "fixed()"}
	if !directWithCode.HasFix() {
		t.Error("FixStrategyDirect with AfterCode: HasFix() = false, want true")
	}

	// IsAutoFixable implies HasFix (lattice property)
	autoFix := Finding{
		FixStrategy: FixStrategyDirect,
		BeforeCode:  "old",
		AfterCode:   "new",
	}
	if autoFix.IsAutoFixable() && !autoFix.HasFix() {
		t.Error("IsAutoFixable()=true but HasFix()=false — violates lattice")
	}
}

// TestSplitBrain_SARIFRegionConsistency verifies that the SARIF location region
// and fix region agree for multi-line findings, resolving split-brain #7.
func TestSplitBrain_SARIFRegionConsistency(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:          "test:ml:main.go:10:5",
		Rule:        "ml",
		ToolName:    "test",
		Message:     "multi-line",
		Severity:    SeverityError,
		Position:    Pos("main.go", 10, 5),
		Range:       NewRangePtr("main.go", 10, 5, 20, 10),
		FixStrategy: FixStrategyDirect,
		BeforeCode:  "old",
		AfterCode:   "new",
	}

	report := NewReport(ToolInfo{Name: "test"})
	report.AddFinding(f)

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF(): %v", err)
	}

	findings, err := FindingsFromSARIF(context.Background(), sarif)
	if err != nil {
		t.Fatalf("FindingsFromSARIF(): %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	rt := findings[0]
	if rt.Range == nil {
		t.Fatal("round-tripped Range is nil")
	}

	if rt.Range.End.Line != 20 {
		t.Errorf("round-tripped Range.End.Line = %d, want 20", rt.Range.End.Line)
	}

	if rt.Range.End.Column != 10 {
		t.Errorf("round-tripped Range.End.Column = %d, want 10", rt.Range.End.Column)
	}
}

// TestSplitBrain_SuppressionRejectsUnknownKind verifies that Suppression
// validation rejects unknown kinds, resolving split-brain #9.
func TestSplitBrain_SuppressionRejectsUnknownKind(t *testing.T) {
	t.Parallel()

	bad := Suppression{Kind: "bogus", Rule: "R1"}
	if bad.IsValid() {
		t.Error("Suppression{Kind:\"bogus\"}.IsValid() = true, want false")
	}

	good := Suppression{Kind: SuppressionInSource, Rule: "R1"}
	if !good.IsValid() {
		t.Error("Suppression{Kind:SuppressionInSource}.IsValid() = false, want true")
	}
}

// TestSplitBrain_PositionConsistency verifies that Position constructors
// produce consistent IsZero/HasOffset results, resolving split-brain #1.
func TestSplitBrain_PositionConsistency(t *testing.T) {
	t.Parallel()

	// Constructor-built Position (no offset provided) should have consistent state
	pos := Pos("main.go", 10, 5)
	if pos.HasOffset() {
		t.Error("Pos() without offset: HasOffset() = true, want false (Offset should be -1)")
	}

	// Position with explicit offset 0 should report HasOffset=true
	withOffset := Position{File: "main.go", Line: 10, Column: 5, Offset: 0}
	if !withOffset.HasOffset() {
		t.Error("Position{Offset:0}.HasOffset() = false, want true (byte 0 is valid)")
	}

	// Unset sentinel Position should be IsZero=true, HasOffset=false
	unset := Position{File: "", Line: 0, Column: 0, Offset: -1}
	if !unset.IsZero() {
		t.Error("Position{Offset:-1}.IsZero() = false, want true")
	}

	if unset.HasOffset() {
		t.Error("Position{Offset:-1}.HasOffset() = true, want false")
	}
}

// TestSplitBrain_CategoryTagsConsistency verifies that Category/Tags conflicts
// are caught by validation, resolving split-brain #4.
func TestSplitBrain_CategoryTagsConsistency(t *testing.T) {
	t.Parallel()

	// Conflicting: Category=security, Tags=[performance] (no security in tags)
	conflicting := Finding{
		ID:          "test:r:main.go:1:1",
		Rule:        "r",
		ToolName:    "test",
		Message:     "m",
		Severity:    SeverityError,
		Position:    Pos("main.go", 1, 1),
		Category:    CategorySecurity,
		Tags:        []Tag{TagPerformance},
		FixStrategy: FixStrategyNone,
	}
	err := conflicting.Validate()
	if err == nil {
		t.Error("conflicting Category/Tags: Validate() = nil, want error")
	}

	// Consistent: Category=security, Tags=[security, performance]
	consistent := Finding{
		ID:          "test:r:main.go:1:1",
		Rule:        "r",
		ToolName:    "test",
		Message:     "m",
		Severity:    SeverityError,
		Position:    Pos("main.go", 1, 1),
		Category:    CategorySecurity,
		Tags:        []Tag{TagSecurity, TagPerformance},
		FixStrategy: FixStrategyNone,
	}
	err = consistent.Validate()
	if err != nil {
		t.Errorf("consistent Category/Tags: Validate() = %v, want nil", err)
	}
}
