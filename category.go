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
	CategoryUnused        Category = "unused"
)

// IsStandard returns true if the category is one of the predefined standard constants.
func (c Category) IsStandard() bool {
	switch c {
	case CategorySecurity, CategoryStyle, CategoryPerformance, CategoryCorrectness,
		CategoryComplexity, CategoryDuplication, CategoryErrorHandling, CategoryMigration,
		CategoryTypeSafety, CategoryStructure, CategoryConfiguration, CategoryDocumentation,
		CategoryTesting, CategoryUnused:
		return true
	}

	return false
}

// IsValid returns true if the category is a non-empty string.
// Custom categories (e.g. "go-vet") are valid. Use IsStandard to check
// for predefined constants only.
func (c Category) IsValid() bool {
	return c != ""
}

// String returns the string representation of the category.
func (c Category) String() string {
	return string(c)
}
