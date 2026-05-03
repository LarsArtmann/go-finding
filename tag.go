package finding

// Tag is a sub-classification label for a finding.
type Tag string

// Standard tag constants for common classification labels.
const (
	TagSecurity      Tag = "security"
	TagPerformance   Tag = "performance"
	TagStyle         Tag = "style"
	TagCorrectness   Tag = "correctness"
	TagBug           Tag = "bug"
	TagDeprecated    Tag = "deprecated"
	TagDocumentation Tag = "documentation"
	TagComplexity    Tag = "complexity"
	TagTest          Tag = "test"
	TagBuild         Tag = "build"
)

// IsStandard returns true if the tag is one of the predefined standard constants.
func (t Tag) IsStandard() bool {
	switch t {
	case TagSecurity, TagPerformance, TagStyle, TagCorrectness,
		TagBug, TagDeprecated, TagDocumentation, TagComplexity,
		TagTest, TagBuild:
		return true
	}

	return false
}

// IsValid returns true if the tag is a non-empty string.
// Custom tags are valid. Use IsStandard to check for predefined constants only.
func (t Tag) IsValid() bool {
	return t != ""
}

// String returns the string representation of the tag.
func (t Tag) String() string {
	return string(t)
}
