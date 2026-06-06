package finding

import (
	"testing"
	"time"
)

func keyTestFinding(id string) Finding {
	return Finding{
		ID:       id,
		Rule:     exportTestR001,
		ToolName: "test",
		Message:  "msg",
		Severity: SeverityError,
		Position: Position{File: benchFile, Line: 10, Column: 5},
	}
}

func TestClone(t *testing.T) {
	const mutated = "changed"

	t.Parallel()

	expires := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	original := Finding{
		ID:          "test:" + exportTestR001 + ":" + benchFile + ":10:5",
		Rule:        exportTestR001,
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityError,
		Position:    Position{File: benchFile, Line: 10, Column: 5},
		Category:    CategorySecurity,
		FixStrategy: FixStrategyDirect,
		Suggestion:  exportTestFixIt,
		BeforeCode:  exportTestOld,
		AfterCode:   exportTestNew,
		Range:       NewRangePtr(benchFile, 10, 5, 10, 20),
		Snippet:     "code here",
		Confidence:  0.95,
		Related:     []RelatedRef{{FindingID: "other:1", Relation: findBuilderTestCauses}},
		Suppression: &Suppression{
			Kind:      SuppressionInSource,
			Reason:    suppTestIntentional,
			ExpiresAt: &expires,
		},
		Tags:     []Tag{TagSecurity, exportTestInjection},
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

	clone.Tags[0] = mutated
	if original.Tags[0] == mutated {
		t.Error("mutating clone Tags should not affect original")
	}

	*clone.Suppression.ExpiresAt = time.Time{}
	if original.Suppression.ExpiresAt.IsZero() {
		t.Error("mutating clone Suppression.ExpiresAt should not affect original")
	}
}

func TestCloneEmpty(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:       "test:" + exportTestR001 + ":" + benchFile + ":1:1",
		Rule:     exportTestR001,
		ToolName: "test",
		Message:  "msg",
		Severity: SeverityInfo,
		Position: Position{File: benchFile, Line: 1},
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
			f:    keyTestFinding("my-id"),
			want: "my-id",
		},
		{
			name: "falls back to composite key when ID empty",
			f:    keyTestFinding(""),
			want: "test\x00file.go\x00R001\x00msg",
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
			want: "test\x00\x00\x00",
		},
		{
			name: "ID with only whitespace is used as-is",
			f:    keyTestFinding("   "),
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

	f := keyTestFinding("")

	key1 := f.Key()

	key2 := f.Key()
	if key1 != key2 {
		t.Error("Key() should be stable across calls")
	}

	f2 := f

	f2.Message = "other"
	if f2.Key() == key1 {
		t.Error("different Message should produce different Key")
	}

	f3 := f

	f3.Rule = "R002"
	if f3.Key() == key1 {
		t.Error("different Rule should produce different Key")
	}

	f4 := f

	f4.Position.File = "other.go"
	if f4.Key() == key1 {
		t.Error("different Position.File should produce different Key")
	}

	f5 := f

	f5.ToolName = "other-tool"
	if f5.Key() == key1 {
		t.Error("different ToolName should produce different Key")
	}
}

func TestEqual_FieldMismatch(t *testing.T) {
	t.Parallel()

	base := Finding{
		ID: "id", Rule: "r", ToolName: "t", Message: "m",
		Severity: SeverityError, Position: Position{File: filterTestFileA, Line: 1},
		Confidence: 0.5,
	}

	tests := []struct {
		name      string
		modify    func(*Finding)
		wantEqual bool
	}{
		{exportTestEqual, nil, true},
		{"different ID", func(f *Finding) { f.ID = "x" }, false},
		{"different Rule", func(f *Finding) { f.Rule = "x" }, false},
		{"different ToolName", func(f *Finding) { f.ToolName = "x" }, false},
		{"different Message", func(f *Finding) { f.Message = "x" }, false},
		{"different Severity", func(f *Finding) { f.Severity = SeverityWarning }, false},
		{"different Category", func(f *Finding) { f.Category = "x" }, false},
		{"different Tags", func(f *Finding) { f.Tags = []Tag{"x"} }, false},
		{"different FixStrategy", func(f *Finding) { f.FixStrategy = FixStrategyDirect }, false},
		{"different Suggestion", func(f *Finding) { f.Suggestion = "x" }, false},
		{"different BeforeCode", func(f *Finding) { f.BeforeCode = "x" }, false},
		{"different AfterCode", func(f *Finding) { f.AfterCode = "x" }, false},
		{"different Snippet", func(f *Finding) { f.Snippet = "x" }, false},
		{"different Confidence", func(f *Finding) { f.Confidence = 0.9 }, false},
		{"different Position", func(f *Finding) { f.Position = Position{File: filterTestFileB} }, false},
		{"different Range", func(f *Finding) { f.Range = NewRangePtr("a.go", 1, 1, 1, 5) }, false},
		{"different Related", func(f *Finding) { f.Related = []RelatedRef{{FindingID: "x"}} }, false},
		{"different Metadata", func(f *Finding) { f.Metadata = map[string]string{"k": "v"} }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			other := base
			if tt.modify != nil {
				tt.modify(&other)
			}

			if got := base.Equal(other); got != tt.wantEqual {
				t.Errorf("Equal() = %v, want %v for case %q", got, tt.wantEqual, tt.name)
			}
		})
	}
}
