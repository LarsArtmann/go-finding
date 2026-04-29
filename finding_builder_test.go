package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Minimal(t *testing.T) {
	t.Parallel()

	pos := Pos("main.go", 42, 5)
	b := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, pos)
	f, err := b.Build()
	require.NoError(t, err)

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

	f, err := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, pos).
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

	require.NoError(t, err)
	assert.Equal(t, "custom-id", f.ID)
	assert.Equal(t, CategorySecurity, f.Category)
	assert.Equal(t, "nil-deref", f.Tag)
	assert.Equal(t, FixStrategyDirect, f.FixStrategy)
	assert.Equal(t, "Add nil check", f.Suggestion)
	assert.Equal(t, "x.foo", f.BeforeCode)
	assert.Equal(t, "x.foo()", f.AfterCode)
	assert.NotNil(t, f.Range)
	assert.Equal(t, "x.foo\n", f.Snippet)
	assert.InDelta(t, 1.0, f.Confidence, 1e-9)
	assert.Len(t, f.Related, 1)
	assert.NotNil(t, f.Suppression)
	assert.Equal(t, "value", f.Metadata["key"])
}

func TestBuilder_Chaining(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)
	f, err := NewBuilder("r", "t", "m", SeverityWarning, pos).
		WithRelated(RelatedRef{FindingID: "r1"}).
		WithRelated(RelatedRef{FindingID: "r2"}).
		Build()

	require.NoError(t, err)
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

	f, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "1", f.Metadata["a"])
	assert.Equal(t, "2", f.Metadata["b"])
}

func TestBuilder_Build_InvalidPanics(t *testing.T) {
	t.Parallel()

	b := &Builder{f: Finding{}} // invalid: no required fields

	_, err := b.Build()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidBuilder)
}

func TestBuilder_Build_MissingFields(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)

	tests := []struct {
		name    string
		builder *Builder
	}{
		{
			"empty rule",
			&Builder{
				f: Finding{
					Rule: "", ToolName: "t", Message: "m",
					Severity: SeverityError, Position: pos,
				},
			},
		},
		{
			"empty tool",
			&Builder{
				f: Finding{
					Rule: "r", ToolName: "", Message: "m",
					Severity: SeverityError, Position: pos,
				},
			},
		},
		{
			"empty message",
			&Builder{
				f: Finding{
					Rule: "r", ToolName: "t", Message: "",
					Severity: SeverityError, Position: pos,
				},
			},
		},
		{
			"empty severity",
			&Builder{
				f: Finding{
					Rule: "r", ToolName: "t", Message: "m",
					Severity: "", Position: pos,
				},
			},
		},
		{
			"empty position",
			&Builder{
				f: Finding{
					Rule: "r", ToolName: "t", Message: "m",
					Severity: SeverityError,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.builder.Build()
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidBuilder)
		})
	}
}

func TestBuilder_Immutability(t *testing.T) {
	t.Parallel()

	pos := Pos("a.go", 1, 1)
	b := NewBuilder("r", "t", "m", SeverityInfo, pos).
		WithMetadata(map[string]string{"key": "original"})

	f1, err := b.Build()
	require.NoError(t, err)
	b.f.Metadata["key"] = "mutated"
	f2, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "original", f1.Metadata["key"])
	assert.Equal(t, "mutated", f2.Metadata["key"])
}
