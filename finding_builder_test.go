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

	if string(f.Related[idx].FindingID) != want {
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

func TestBuilder_GroupID(t *testing.T) {
	t.Parallel()

	f, err := NewBuilder("clone-detected", "art-dupl", "duplicate code", SeverityWarning, Pos("a.go", 1, 1)).
		WithGroupID("group-1").
		Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if f.GroupID != GroupID("group-1") {
		t.Errorf("GroupID = %q, want %q", f.GroupID, "group-1")
	}
}

func TestBuilder_Chaining(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)

	f, err := NewBuilder("r", "t", "m", SeverityWarning, pos).
		WithRelated(RelatedRef{FindingID: "r1", Relation: RelationRelated}).
		WithRelated(RelatedRef{FindingID: "r2", Relation: RelationRelated}).
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

	AssertPanics(t, "expected panic for invalid builder", func() {
		b.MustBuild()
	})
}

func TestBuilder_BuildOrDefault_Valid(t *testing.T) {
	t.Parallel()

	pos := Pos("main.go", 42, 5)

	f := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, pos).
		BuildOrDefault()

	if f.Rule != "nilcheck" {
		t.Errorf("Rule = %q, want %q", f.Rule, "nilcheck")
	}

	if f.ToolName != "govet" {
		t.Errorf("ToolName = %q, want %q", f.ToolName, "govet")
	}

	if f.Message != "possible nil deref" {
		t.Errorf("Message = %q, want %q", f.Message, "possible nil deref")
	}

	if f.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestBuilder_BuildOrDefault_Invalid(t *testing.T) {
	t.Parallel()

	b := &Builder{f: Finding{}}

	f := b.BuildOrDefault()

	if f.Rule != "" {
		t.Errorf("Rule = %q, want empty", f.Rule)
	}

	if f.ID != "" {
		t.Errorf("ID = %q, want empty", f.ID)
	}
}

func TestTemplate_Basic(t *testing.T) {
	t.Parallel()

	tmpl := NewTemplate("my-tool").
		WithCategory(CategorySecurity).
		WithFixStrategy(FixStrategySuggest).
		WithTags(Tag("security"), Tag("auto-fix"))

	f := tmpl.Build("SQL-injection", "unescaped query", SeverityError, Pos("db.go", 10, 5))

	if f.ToolName != "my-tool" {
		t.Errorf("ToolName = %q, want %q", f.ToolName, "my-tool")
	}

	if f.Category != CategorySecurity {
		t.Errorf("Category = %v, want %v", f.Category, CategorySecurity)
	}

	if f.FixStrategy != FixStrategySuggest {
		t.Errorf("FixStrategy = %v, want %v", f.FixStrategy, FixStrategySuggest)
	}

	if len(f.Tags) != 2 {
		t.Fatalf("Tags length = %d, want 2", len(f.Tags))
	}

	if f.Tags[0] != Tag("security") || f.Tags[1] != Tag("auto-fix") {
		t.Errorf("Tags = %v, want [security, auto-fix]", f.Tags)
	}

	if f.Rule != "SQL-injection" {
		t.Errorf("Rule = %q, want %q", f.Rule, "SQL-injection")
	}

	if f.Message != "unescaped query" {
		t.Errorf("Message = %q, want %q", f.Message, "unescaped query")
	}

	if f.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestTemplate_MultipleFindings(t *testing.T) {
	t.Parallel()

	tmpl := NewTemplate("linter").WithCategory(CategoryStyle)

	f1 := tmpl.Build("R1", "message 1", SeverityInfo, Pos("a.go", 1, 1))
	f2 := tmpl.Build("R2", "message 2", SeverityWarning, Pos("b.go", 2, 3))

	if f1.Rule != "R1" || f2.Rule != "R2" {
		t.Errorf("Rules = %q, %q; want R1, R2", f1.Rule, f2.Rule)
	}

	if f1.ToolName != "linter" || f2.ToolName != "linter" {
		t.Errorf("ToolName mismatch")
	}

	if f1.Category != CategoryStyle || f2.Category != CategoryStyle {
		t.Errorf("Category not stamped")
	}

	if f1.ID == f2.ID {
		t.Error("findings should have different IDs")
	}
}

func TestTemplate_NoCategory(t *testing.T) {
	t.Parallel()

	tmpl := NewTemplate("tool")

	f := tmpl.Build("R1", "msg", SeverityInfo, Pos("a.go", 1, 1))

	if f.Category != "" {
		t.Errorf("Category = %q, want empty", f.Category)
	}

	if f.FixStrategy != FixStrategyNone {
		t.Errorf("FixStrategy = %v, want %v", f.FixStrategy, FixStrategyNone)
	}
}

func TestTemplate_InvalidInput(t *testing.T) {
	t.Parallel()

	tmpl := NewTemplate("tool")

	f := tmpl.Build("", "", SeverityError, Position{})

	if f.ID != "" {
		t.Error("invalid input should return zero-value Finding")
	}
}

func TestTemplate_Builder_AllowsChainingConfidenceAndSuggestion(t *testing.T) {
	t.Parallel()

	tmpl := NewTemplate("my-linter").
		WithCategory(CategoryStyle).
		WithFixStrategy(FixStrategySuggest)

	f := tmpl.Builder("R1", "bad pattern", SeverityWarning, Pos("demo.go", 42, 3)).
		WithConfidence(ConfidenceHigh).
		WithSuggestion("use humanize.Bytes instead").
		MustBuild()

	if f.ToolName != "my-linter" {
		t.Errorf("ToolName = %q, want %q", f.ToolName, "my-linter")
	}

	if f.Category != CategoryStyle {
		t.Errorf("Category = %v, want %v", f.Category, CategoryStyle)
	}

	if f.FixStrategy != FixStrategySuggest {
		t.Errorf("FixStrategy = %v, want %v", f.FixStrategy, FixStrategySuggest)
	}

	if f.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %v, want %v", f.Confidence, ConfidenceHigh)
	}

	if f.Suggestion != "use humanize.Bytes instead" {
		t.Errorf("Suggestion = %q, want %q", f.Suggestion, "use humanize.Bytes instead")
	}

	if f.Rule != "R1" {
		t.Errorf("Rule = %q, want %q", f.Rule, "R1")
	}

	if f.Message != "bad pattern" {
		t.Errorf("Message = %q, want %q", f.Message, "bad pattern")
	}
}

func TestTemplate_Builder_DelegatesToBuild(t *testing.T) {
	t.Parallel()

	tmpl := NewTemplate("tool").WithCategory(CategoryStyle)

	built := tmpl.Build("R1", "msg", SeverityInfo, Pos("a.go", 1, 1))
	buildered := tmpl.Builder("R1", "msg", SeverityInfo, Pos("a.go", 1, 1)).BuildOrDefault()

	if built.ToolName != buildered.ToolName {
		t.Errorf("ToolName mismatch: Build=%q, Builder=%q", built.ToolName, buildered.ToolName)
	}

	if built.Category != buildered.Category {
		t.Errorf("Category mismatch: Build=%v, Builder=%v", built.Category, buildered.Category)
	}

	if built.Rule != buildered.Rule {
		t.Errorf("Rule mismatch: Build=%q, Builder=%q", built.Rule, buildered.Rule)
	}
}

func TestTemplate_Builder_WithoutChaining(t *testing.T) {
	t.Parallel()

	tmpl := NewTemplate("tool").WithCategory(CategoryStyle)

	f := tmpl.Builder("R1", "msg", SeverityInfo, Pos("a.go", 1, 1)).BuildOrDefault()

	if f.ID == "" {
		t.Error("Builder without chaining should produce valid finding")
	}

	if f.Category != CategoryStyle {
		t.Errorf("Category = %v, want %v", f.Category, CategoryStyle)
	}
}

func TestTemplate_WithGroupID(t *testing.T) {
	t.Parallel()

	tmpl := NewTemplate("dupl").
		WithGroupID(GroupID("clone-42"))

	pos := Pos(FilePath("a.go"), 1, 1)

	f := tmpl.Build(RuleName("dup-block"), "duplicated block", SeverityWarning, pos)
	if f.GroupID != GroupID("clone-42") {
		t.Errorf("Build GroupID = %q, want clone-42", f.GroupID)
	}

	b := tmpl.Builder(RuleName("dup-block"), "duplicated block", SeverityWarning, pos)
	built := b.MustBuild()
	if built.GroupID != GroupID("clone-42") {
		t.Errorf("Builder GroupID = %q, want clone-42", built.GroupID)
	}

	ungrouped := NewTemplate("solo").Build(RuleName("r"), "m", SeverityInfo, pos)
	if ungrouped.GroupID != "" {
		t.Errorf("template without WithGroupID stamped GroupID = %q, want empty", ungrouped.GroupID)
	}
}
