package finding

import "strings"

// linterCategories maps well-known linter/analyzer names to their default Category.
// Extend at runtime via RegisterLinterCategory.
var linterCategories = map[string]Category{
	// Security
	"gosec":      CategorySecurity,
	"noctx":      CategorySecurity,
	"errchkjson": CategorySecurity,

	// Correctness
	"govet":        CategoryCorrectness,
	"staticcheck":  CategoryCorrectness,
	"errcheck":     CategoryCorrectness,
	"nilerr":       CategoryCorrectness,
	"ineffassign":  CategoryCorrectness,
	"unconvert":    CategoryCorrectness,
	"bodyclose":    CategoryCorrectness,
	"contextcheck": CategoryCorrectness,
	"durationcheck": CategoryCorrectness,
	"typecheck":    CategoryCorrectness,
	"gosimple":     CategoryCorrectness,
	"deadcode":     CategoryCorrectness,
	"varcheck":     CategoryCorrectness,

	// Performance
	"prealloc":   CategoryPerformance,
	"perfsprint": CategoryPerformance,
	"unparam":    CategoryPerformance,

	// Complexity
	"gocyclo":        CategoryComplexity,
	"cyclop":         CategoryComplexity,
	"gocognit":       CategoryComplexity,
	"maintidx":       CategoryComplexity,
	"funlen":         CategoryComplexity,
	"nestif":         CategoryComplexity,
	"interfacebloat": CategoryComplexity,
	"gocritic":       CategoryComplexity,

	// Duplication
	"dupl":    CategoryDuplication,
	"goconst": CategoryDuplication,

	// Error handling
	"wrapcheck":  CategoryErrorHandling,
	"errorlint":  CategoryErrorHandling,
	"errname":    CategoryErrorHandling,
	"nilnil":     CategoryErrorHandling,

	// Style
	"misspell":  CategoryStyle,
	"revive":    CategoryStyle,
	"gofmt":     CategoryStyle,
	"goimports": CategoryStyle,
	"gci":       CategoryStyle,
	"wsl_v5":    CategoryStyle,
	"dupword":   CategoryStyle,
	"godot":     CategoryStyle,
	"lll":       CategoryStyle,
	"whitespace": CategoryStyle,
	"nlreturn":  CategoryStyle,
	"golint":    CategoryStyle,

	// Testing
	"paralleltest":       CategoryTesting,
	"thelper":            CategoryTesting,
	"testifylint":        CategoryTesting,
	"ginkgolinter":       CategoryTesting,
	"tparallel":          CategoryTesting,
	"testpackage":        CategoryTesting,
	"testableexamples":   CategoryTesting,

	// Type safety
	"exhaustive":     CategoryTypeSafety,
	"exhaustruct":    CategoryTypeSafety,
	"forcetypeassert": CategoryTypeSafety,
	"musttag":        CategoryTypeSafety,
	"gochecksumtype": CategoryTypeSafety,
	"copyloopvar":    CategoryTypeSafety,
	"intrange":       CategoryTypeSafety,

	// Structure
	"sloglint":      CategoryStructure,
	"loggercheck":   CategoryStructure,
	"unused":        CategoryUnused,

	// Configuration
	"gomodguard_v2": CategoryConfiguration,
}

// CategoryForLinter returns the default Category for a well-known linter or analyzer name.
// The lookup is case-insensitive. If the name is not registered, it returns CategoryCorrectness.
// Register custom mappings with RegisterLinterCategory.
func CategoryForLinter(name string) Category {
	if cat, ok := linterCategories[strings.ToLower(name)]; ok {
		return cat
	}

	return CategoryCorrectness
}

// RegisterLinterCategory registers or overrides the Category for a linter name.
// The name is stored in lowercase for case-insensitive lookup.
// This is safe for concurrent use via init-time or early-program setup.
func RegisterLinterCategory(name string, cat Category) {
	linterCategories[strings.ToLower(name)] = cat
}
