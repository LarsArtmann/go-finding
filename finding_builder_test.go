package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_Minimal(t *testing.T) {
	t.Parallel()

	pos := Pos("main.go", 42, 5)
	b := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, pos)
	f := b.Build()

	assert.Equal(t, "nilcheck", f.Rule)
	assert.Equal(t, "govet", f.ToolName)
	assert.Equal(t, "possible nil deref", f.Message)
	assert.Equal(t, SeverityError, f.Severity)
	assert.Equal(t, pos, f.Position)
	assert.NotEmpty(t, f.ID)
	assert.Equal(t, FixStrategyNone, f.FixStrategy)
}

func TestBuilder_Full(t *testing.T) {
	t.Parallel()

	pos := Pos("main.go", 42, 5)
	rng := NewRange("main.go", 42, 5, 42, 10)

	f := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, pos).
		WithID("custom-id").
		WithCategory(CategorySecurity).
		WithTag("nil-deref").
		WithFixStrategy(FixStrategyDirect).
		WithSuggestion("Add nil check").
		WithBeforeCode("x.foo").
		WithAfterCode("x.foo()").
		WithRange(rng).
		WithSnippet("x.foo\n").
		WithConfidence(1.5).
		WithRelated(RelatedRef{FindingID: "other", Relation: "causes"}).
		WithSuppression(Suppression{Kind: SuppressionInSource, Rule: "R1"}).
		WithMetadata(map[string]string{"key": "value"}).
		Build()

	assert.Equal(t, "custom-id", f.ID)
	assert.Equal(t, CategorySecurity, f.Category)
	assert.Equal(t, "nil-deref", f.Tag)
	assert.Equal(t, FixStrategyDirect, f.FixStrategy)
	assert.Equal(t, "Add nil check", f.Suggestion)
	assert.Equal(t, "x.foo", f.BeforeCode)
	assert.Equal(t, "x.foo()", f.AfterCode)
	assert.NotNil(t, f.Range)
	assert.Equal(t, "x.foo\n", f.Snippet)
	assert.Equal(t, 1.0, f.Confidence)
	assert.Len(t, f.Related, 1)
	assert.NotNil(t, f.Suppression)
	assert.Equal(t, "value", f.Metadata["key"])
}

func TestBuilder_Chaining(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)
	f := NewBuilder("r", "t", "m", SeverityWarning, pos).
		WithRelated(RelatedRef{FindingID: "r1"}).
		WithRelated(RelatedRef{FindingID: "r2"}).
		Build()

	assert.Len(t, f.Related, 2)
	assert.Equal(t, "r1", f.Related[0].FindingID)
	assert.Equal(t, "r2", f.Related[1].FindingID)
}

func TestBuilder_MetadataMerge(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)
	b := NewBuilder("r", "t", "m", SeverityInfo, pos).
		WithMetadata(map[string]string{"a": "1"}).
		WithMetadata(map[string]string{"b": "2"})

	f := b.Build()

	assert.Equal(t, "1", f.Metadata["a"])
	assert.Equal(t, "2", f.Metadata["b"])
}

func TestBuilder_Immutability(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)
	b := NewBuilder("r", "t", "m", SeverityInfo, pos).
		WithMetadata(map[string]string{"key": "original"})

	f1 := b.Build()
	b.f.Metadata["key"] = "mutated"
	f2 := b.Build()

	assert.Equal(t, "original", f1.Metadata["key"])
	assert.Equal(t, "mutated", f2.Metadata["key"])
}
