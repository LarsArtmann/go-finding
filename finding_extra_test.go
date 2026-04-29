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
		Range:       NewRangePtr("file.go", 10, 5, 10, 20),
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

func TestEqualTimePtr(t *testing.T) {
	t.Parallel()

	now := time.Now()

	if !equalTimePtr(nil, nil) {
		t.Error("equalTimePtr(nil, nil) should be true")
	}

	if equalTimePtr(&now, nil) {
		t.Error("equalTimePtr(&now, nil) should be false")
	}

	if equalTimePtr(nil, &now) {
		t.Error("equalTimePtr(nil, &now) should be false")
	}

	other := now
	if !equalTimePtr(&now, &other) {
		t.Error("equalTimePtr(&now, &same) should be true")
	}
}

func TestFindingKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want string
	}{
		{
			name: "returns ID when set",
			f: Finding{
				ID:       "my-id",
				Rule:     "R001",
				ToolName: "test",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: "file.go", Line: 10, Column: 5},
			},
			want: "my-id",
		},
		{
			name: "falls back to composite key when ID empty",
			f: Finding{
				Rule:     "R001",
				ToolName: "test",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: "file.go", Line: 10, Column: 5},
			},
			want: "file.go\x00R001\x00msg",
		},
		{
			name: "empty file/rule/message still produces key",
			f: Finding{
				Rule:     "",
				ToolName: "test",
				Message:  "",
				Severity: SeverityError,
				Position: Position{File: "", Line: 0, Column: 0},
			},
			want: "\x00\x00",
		},
		{
			name: "ID with only whitespace is used as-is",
			f: Finding{
				ID:       "   ",
				Rule:     "R001",
				ToolName: "test",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: "file.go", Line: 10, Column: 5},
			},
			want: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.f.Key()
			if got != tt.want {
				t.Errorf("Key() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindingKeyStability(t *testing.T) {
	t.Parallel()

	// Key must be stable: same inputs → same output.
	f := Finding{
		Rule:     "R001",
		ToolName: "test",
		Message:  "msg",
		Severity: SeverityError,
		Position: Position{File: "file.go", Line: 10, Column: 5},
	}

	key1 := f.Key()
	key2 := f.Key()
	if key1 != key2 {
		t.Error("Key() should be stable across calls")
	}

	// Different message → different key.
	f2 := f
	f2.Message = "other"
	if f2.Key() == key1 {
		t.Error("different Message should produce different Key")
	}

	// Different rule → different key.
	f3 := f
	f3.Rule = "R002"
	if f3.Key() == key1 {
		t.Error("different Rule should produce different Key")
	}

	// Different file → different key.
	f4 := f
	f4.Position.File = "other.go"
	if f4.Key() == key1 {
		t.Error("different Position.File should produce different Key")
	}
}
