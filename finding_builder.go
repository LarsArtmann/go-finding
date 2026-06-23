package finding

import "maps"

// Builder provides a fluent API for constructing Finding values.
// Use NewBuilder with the required fields, then chain With* methods
// for optional fields, and call Build to obtain the result.
//
// Example:
//
//	f := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, Pos("main.go", 42, 5)).
//		WithFixStrategy(FixStrategyDirect).
//		WithBeforeCode("x.foo").
//		WithAfterCode("x.foo()").
//		Build()
type Builder struct {
	f Finding
}

// NewBuilder creates a builder seeded with the required fields.
// The ID is auto-generated from the provided arguments.
func NewBuilder(rule RuleName, toolName ToolName, message string, severity Severity, pos Position) *Builder {
	return &Builder{f: NewFinding(rule, toolName, message, severity, pos, 0)}
}

// WithID overrides the auto-generated ID.
func (b *Builder) WithID(id ID) *Builder {
	b.f.ID = id

	return b
}

// WithCategory sets the category.
func (b *Builder) WithCategory(cat Category) *Builder {
	b.f.Category = cat

	return b
}

// WithTags sets multiple tags.
func (b *Builder) WithTags(tags ...Tag) *Builder {
	b.f.Tags = append(b.f.Tags, tags...)

	return b
}

// WithFixStrategy sets the fix strategy.
func (b *Builder) WithFixStrategy(fs FixStrategy) *Builder {
	b.f.FixStrategy = fs

	return b
}

// WithSuggestion sets the human-readable fix suggestion.
func (b *Builder) WithSuggestion(s string) *Builder {
	b.f.Suggestion = s

	return b
}

// WithBeforeCode sets the code before the fix.
func (b *Builder) WithBeforeCode(code string) *Builder {
	b.f.BeforeCode = code

	return b
}

// WithAfterCode sets the code after the fix.
func (b *Builder) WithAfterCode(code string) *Builder {
	b.f.AfterCode = code

	return b
}

// WithRange sets the source range.
func (b *Builder) WithRange(r Range) *Builder {
	b.f.Range = &r

	return b
}

// WithSnippet sets the surrounding code context.
func (b *Builder) WithSnippet(s string) *Builder {
	b.f.Snippet = s

	return b
}

// WithConfidence sets the confidence level (clamped to [0.0, 1.0]).
func (b *Builder) WithConfidence(c Confidence) *Builder {
	b.f.Confidence = c.Clamp()

	return b
}

// WithRelated appends related references.
func (b *Builder) WithRelated(refs ...RelatedRef) *Builder {
	b.f.Related = append(b.f.Related, refs...)

	return b
}

// WithSuppression sets the suppression info.
func (b *Builder) WithSuppression(s Suppression) *Builder {
	b.f.Suppression = &s

	return b
}

// WithMetadata copies the given metadata into the finding.
func (b *Builder) WithMetadata(m map[string]string) *Builder {
	if b.f.Metadata == nil {
		b.f.Metadata = make(map[string]string, len(m))
	}

	maps.Copy(b.f.Metadata, m)

	return b
}

// Build returns the constructed Finding.
// Returns a detailed validation error if required fields are missing or invalid.
// FixStrategy is normalized: empty string becomes FixStrategyNone.
func (b *Builder) Build() (Finding, error) {
	b.f = b.f.Normalized()

	err := b.f.Validate()
	if err != nil {
		return Finding{}, err
	}

	return b.f.Clone(), nil
}

// MustBuild returns the constructed Finding or panics if required fields are missing.
// Use this only when the builder is fully configured and invalid state is a programmer error.
func (b *Builder) MustBuild() Finding {
	f, err := b.Build()
	if err != nil {
		panic(err)
	}

	return f
}
