package finding

// Category classifies the domain of a finding.
type Category string

// Standard category constants for findings.
const (
	CategorySecurity      Category = "security"
	CategoryStyle         Category = "style"
	CategoryPerformance   Category = "performance"
	CategoryCorrectness   Category = "correctness"
	CategoryComplexity    Category = "complexity"
	CategoryDuplication   Category = "duplication"
	CategoryErrorHandling Category = "error-handling"
	CategoryMigration     Category = "migration"
	CategoryTypeSafety    Category = "type-safety"
	CategoryStructure     Category = "structure"
	CategoryConfiguration Category = "configuration"
	CategoryDocumentation Category = "documentation"
	CategoryTesting       Category = "testing"
)

// IsValid returns true if the category is a standard value.
func (c Category) IsValid() bool {
	switch c {
	case CategorySecurity, CategoryStyle, CategoryPerformance, CategoryCorrectness,
		CategoryComplexity, CategoryDuplication, CategoryErrorHandling, CategoryMigration,
		CategoryTypeSafety, CategoryStructure, CategoryConfiguration, CategoryDocumentation,
		CategoryTesting:
		return true
	}

	return false
}
