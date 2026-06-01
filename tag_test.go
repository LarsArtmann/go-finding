package finding

import "testing"

func TestTag_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tag  Tag
		want bool
	}{
		{TagSecurity, true},
		{TagBug, true},
		{Tag(""), false},
		{Tag("custom"), true},
		{Tag("go-vet"), true},
		{Tag("UPPER"), false},
		{Tag("has space"), false},
		{Tag("under_score"), false},
	}

	for _, tt := range tests {
		if got := tt.tag.IsValid(); got != tt.want {
			t.Errorf("Tag(%q).IsValid() = %v, want %v", tt.tag, got, tt.want)
		}
	}
}

func TestTag_IsStandard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tag  Tag
		want bool
	}{
		{TagSecurity, true},
		{TagPerformance, true},
		{TagStyle, true},
		{TagCorrectness, true},
		{TagBug, true},
		{TagDeprecated, true},
		{TagDocumentation, true},
		{TagComplexity, true},
		{TagTest, true},
		{TagBuild, true},
		{Tag(""), false},
		{Tag("custom"), false},
		{Tag("Security"), false},
	}

	for _, tt := range tests {
		if got := tt.tag.IsStandard(); got != tt.want {
			t.Errorf("Tag(%q).IsStandard() = %v, want %v", tt.tag, got, tt.want)
		}
	}
}

func TestTag_String(t *testing.T) {
	t.Parallel()

	if got := TagSecurity.String(); got != string(TagSecurity) {
		t.Errorf("TagSecurity.String() = %q, want %q", got, string(TagSecurity))
	}

	if got := Tag("custom").String(); got != "custom" {
		t.Errorf("Tag(\"custom\").String() = %q, want %q", got, "custom")
	}

	if got := Tag("").String(); got != "" {
		t.Errorf("Tag(\"\").String() = %q, want %q", got, "")
	}
}

func TestTag_Constants(t *testing.T) {
	t.Parallel()

	_ = TagSecurity
	_ = TagBug
	_ = TagTest
	_ = TagPerformance
	_ = TagStyle
	_ = TagCorrectness
	_ = TagDeprecated
	_ = TagDocumentation
	_ = TagComplexity
	_ = TagBuild
}
