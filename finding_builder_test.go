package finding

import (
	"errors"
	"testing"
)

const (
	findBuilderTestCauses    = "causes"
	findBuilderTestKey       = "key"
	findBuilderTestEmptyRule = "empty rule"
	findBuilderTestMutated   = "mutated"
)

func assertRelatedID(t *testing.T, f Finding, idx int, want string) {
	t.Helper()

	if f.Related[idx].FindingID != want {
		t.Errorf("Related[%d] = %q, want %q", idx, f.Related[idx].FindingID, want)
	}
}

func missingFieldBuilder(pos Position, zero func(*Finding)) *Builder {
	f := Finding{Rule: "r", ToolName: "t", Message: "m", Severity: SeverityError, Position: pos}
	zero(&f)

	return &Builder{f: f}
}

func TestBuilder_Minimal(t *testing.T) {
	t.Parallel()

	pos := Pos("main.go", 42, 5)
	b := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, pos)

	f, err := b.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	AssertFindingFields(t, f, "nilcheck", "govet", "possible nil deref", SeverityError, pos)

	if f.ID == "" {
		t.Error("ID should not be empty")
	}

	if f.FixStrategy != FixStrategyNone {
		t.Errorf("FixStrategy = %v, want %v", f.FixStrategy, FixStrategyNone)
	}
}

func TestBuilder_Full(t *testing.T) {
	t.Parallel()

	pos := Pos("main.go", 42, 5)
	rng := NewRange("main.go", 42, 5, 42, 10)

	f, err := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, pos).
		WithID("custom-id").
		WithCategory(CategorySecurity).
		WithTags(Tag("nil-deref"), Tag("security"), Tag("injection")).
		WithFixStrategy(FixStrategyDirect).
		WithSuggestion("Add nil check").
		WithBeforeCode("x.foo").
		WithAfterCode("x.foo()").
		WithRange(rng).
		WithSnippet("x.foo\n").
		WithConfidence(1.5).
		WithRelated(RelatedRef{FindingID: "other", Relation: findBuilderTestCauses}).
		WithSuppression(Suppression{Kind: SuppressionInSource, Rule: "R1"}).
		WithMetadata(map[string]string{findBuilderTestKey: "value"}).
		Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if f.ID != "custom-id" {
		t.Errorf("ID = %q, want %q", f.ID, "custom-id")
	}

	if f.Category != CategorySecurity {
		t.Errorf("Category = %v, want %v", f.Category, CategorySecurity)
	}

	if len(f.Tags) != 3 || f.Tags[0] != Tag("nil-deref") || f.Tags[1] != Tag("security") ||
		f.Tags[2] != Tag("injection") {
		t.Errorf("Tags = %v, want [nil-deref, security, injection]", f.Tags)
	}

	if f.FixStrategy != FixStrategyDirect {
		t.Errorf("FixStrategy = %v, want %v", f.FixStrategy, FixStrategyDirect)
	}

	if f.Suggestion != "Add nil check" {
		t.Errorf("Suggestion = %q, want %q", f.Suggestion, "Add nil check")
	}

	if f.BeforeCode != "x.foo" {
		t.Errorf("BeforeCode = %q, want %q", f.BeforeCode, "x.foo")
	}

	if f.AfterCode != "x.foo()" {
		t.Errorf("AfterCode = %q, want %q", f.AfterCode, "x.foo()")
	}

	if f.Range == nil {
		t.Error("Range should not be nil")
	}

	if f.Snippet != "x.foo\n" {
		t.Errorf("Snippet = %q, want %q", f.Snippet, "x.foo\n")
	}

	if f.Confidence < 1.0-1e-9 || f.Confidence > 1.0+1e-9 {
		t.Errorf("Confidence = %v, want ~1.0", f.Confidence)
	}

	if len(f.Related) != 1 {
		t.Errorf("Related length = %d, want 1", len(f.Related))
	}

	if f.Suppression == nil {
		t.Error("Suppression should not be nil")
	}

	if f.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %q, want %q", f.Metadata["key"], "value")
	}
}

func TestBuilder_Chaining(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)

	f, err := NewBuilder("r", "t", "m", SeverityWarning, pos).
		WithRelated(RelatedRef{FindingID: "r1"}).
		WithRelated(RelatedRef{FindingID: "r2"}).
		Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if len(f.Related) != 2 {
		t.Fatalf("Related length = %d, want 2", len(f.Related))
	}

	assertRelatedID(t, f, 0, "r1")
	assertRelatedID(t, f, 1, "r2")
}

func TestBuilder_MetadataMerge(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)
	b := NewBuilder("r", "t", "m", SeverityInfo, pos).
		WithMetadata(map[string]string{"a": "1"}).
		WithMetadata(map[string]string{"b": "2"})

	f, err := b.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if f.Metadata["a"] != "1" {
		t.Errorf("Metadata[a] = %q, want %q", f.Metadata["a"], "1")
	}

	if f.Metadata["b"] != "2" {
		t.Errorf("Metadata[b] = %q, want %q", f.Metadata["b"], "2")
	}
}

func TestBuilder_Build_InvalidPanics(t *testing.T) {
	t.Parallel()

	b := &Builder{f: Finding{}}

	_, err := b.Build()
	if err == nil {
		t.Fatal("expected error for invalid builder")
	}

	if !errors.Is(err, ErrValidation) {
		t.Errorf("error = %v, want ErrValidation", err)
	}
}

func TestBuilder_Build_MissingFields(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)

	tests := []struct {
		name    string
		builder *Builder
	}{
		{"empty rule", missingFieldBuilder(pos, func(f *Finding) { f.Rule = "" })},
		{"empty tool", missingFieldBuilder(pos, func(f *Finding) { f.ToolName = "" })},
		{"empty message", missingFieldBuilder(pos, func(f *Finding) { f.Message = "" })},
		{"empty severity", missingFieldBuilder(pos, func(f *Finding) { f.Severity = "" })},
		{
			"empty position",
			&Builder{f: Finding{Rule: "r", ToolName: "t", Message: "m", Severity: SeverityError}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.builder.Build()
			if err == nil {
				t.Fatal("expected error for missing fields")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("error = %v, want ErrValidation", err)
			}
		})
	}
}

func TestBuilder_Immutability(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)
	b := NewBuilder("r", "t", "m", SeverityInfo, pos).
		WithMetadata(map[string]string{"key": "original"})

	f1, err := b.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	b.f.Metadata["key"] = "mutated"

	f2, err := b.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if f1.Metadata["key"] != "original" {
		t.Errorf("f1 Metadata[key] = %q, want %q", f1.Metadata["key"], "original")
	}

	if f2.Metadata["key"] != "mutated" {
		t.Errorf("f2 Metadata[key] = %q, want %q", f2.Metadata["key"], "mutated")
	}
}

func TestBuilder_MustBuild(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)

	f := NewBuilder("r", "t", "m", SeverityInfo, pos).MustBuild()
	if f.Rule != "r" {
		t.Errorf("Rule = %q, want %q", f.Rule, "r")
	}
}

func TestBuilder_MustBuild_PanicsOnInvalid(t *testing.T) {
	t.Parallel()

	b := &Builder{f: Finding{}}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid builder")
		}
	}()

	b.MustBuild()
}
