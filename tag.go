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
