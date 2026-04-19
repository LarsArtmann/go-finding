package finding

import (
	"testing"
	"time"
)

func TestClone(t *testing.T) {
	const mutated = "changed"
	t.Parallel()

	expires := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	original := Finding{
		ID:          "test:R001:file.go:10:5",
		Rule:        "R001",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityError,
		Position:    Position{File: "file.go", Line: 10, Column: 5},
		Category:    CategorySecurity,
		Tag:         "injection",
		FixStrategy: FixStrategyDirect,
		Suggestion:  "fix it",
		BeforeCode:  "old",
		AfterCode:   "new",
		Range:       posRange("file.go", 10, 5, 10, 20),
		Snippet:     "code here",
		Confidence:  0.95,
		Related:     []RelatedRef{{FindingID: "other:1", Relation: "causes"}},
		Suppression: &Suppression{
			Kind:      SuppressionInSource,
			Reason:    "intentional",
			ExpiresAt: &expires,
		},
		Metadata: map[string]string{"key": "value"},
	}

	clone := original.Clone()

	if !clone.Equal(original) {
		t.Error("Clone should be Equal to original")
	}

	clone.Metadata["key"] = mutated
	if original.Metadata["key"] == mutated {
		t.Error("mutating clone Metadata should not affect original")
	}

	clone.Related[0] = RelatedRef{FindingID: mutated}
	if original.Related[0].FindingID == mutated {
		t.Error("mutating clone Related should not affect original")
	}

	clone.Range.End.Line = 999
	if original.Range.End.Line == 999 {
		t.Error("mutating clone Range should not affect original")
	}

	*clone.Suppression.ExpiresAt = time.Time{}
	if original.Suppression.ExpiresAt.IsZero() {
		t.Error("mutating clone Suppression.ExpiresAt should not affect original")
	}
}

func TestCloneEmpty(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID: "test:R001:file.go:1:1", Rule: "R001", ToolName: "test",
		Message: "msg", Severity: SeverityInfo,
		Position: Position{File: "file.go", Line: 1},
	}

	clone := f.Clone()
	if !clone.Equal(f) {
		t.Error("Clone of simple finding should be Equal")
	}
}

func posRange(file string, startLine, startCol, endLine, endCol int) *Range {
	return &Range{
		Start: Position{File: file, Line: startLine, Column: startCol},
		End:   Position{File: file, Line: endLine, Column: endCol},
	}
}
