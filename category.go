package finding

// Standard category constants for findings.
const (
	CategorySecurity      = "security"
	CategoryStyle         = "style"
	CategoryPerformance   = "performance"
	CategoryCorrectness   = "correctness"
	CategoryComplexity    = "complexity"
	CategoryDuplication   = "duplication"
	CategoryErrorHandling = "error-handling"
	CategoryMigration     = "migration"
	CategoryTypeSafety    = "type-safety"
	CategoryStructure     = "structure"
	CategoryConfiguration = "configuration"
	CategoryDocumentation = "documentation"
	CategoryTesting       = "testing"
)

// IsStandardCategory returns true if the category is a standard value.
func IsStandardCategory(cat string) bool {
	switch cat {
	case CategorySecurity, CategoryStyle, CategoryPerformance, CategoryCorrectness,
		CategoryComplexity, CategoryDuplication, CategoryErrorHandling, CategoryMigration,
		CategoryTypeSafety, CategoryStructure, CategoryConfiguration, CategoryDocumentation,
		CategoryTesting:
		return true
	}

	return false
}
