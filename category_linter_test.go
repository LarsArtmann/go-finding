package finding

import (
	"testing"
)

func TestCategoryForLinter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		linter   string
		expected Category
	}{
		{"gosec maps to security", "gosec", CategorySecurity},
		{"govet maps to correctness", "govet", CategoryCorrectness},
		{"staticcheck maps to correctness", "staticcheck", CategoryCorrectness},
		{"errcheck maps to correctness", "errcheck", CategoryCorrectness},
		{"gocyclo maps to complexity", "gocyclo", CategoryComplexity},
		{"dupl maps to duplication", "dupl", CategoryDuplication},
		{"wrapcheck maps to error-handling", "wrapcheck", CategoryErrorHandling},
		{"misspell maps to style", "misspell", CategoryStyle},
		{"prealloc maps to performance", "prealloc", CategoryPerformance},
		{"paralleltest maps to testing", "paralleltest", CategoryTesting},
		{"exhaustruct maps to type-safety", "exhaustruct", CategoryTypeSafety},
		{"unused maps to unused", "unused", CategoryUnused},
		{"unknown defaults to correctness", "some-unknown-linter", CategoryCorrectness},
		{"empty defaults to correctness", "", CategoryCorrectness},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CategoryForLinter(tt.linter)
			if got != tt.expected {
				t.Errorf("CategoryForLinter(%q) = %q, want %q", tt.linter, got, tt.expected)
			}
		})
	}
}

func TestCategoryForLinterCaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect Category
	}{
		{"lowercase", "gosec", CategorySecurity},
		{"uppercase", "GOSEC", CategorySecurity},
		{"mixed case", "GoSec", CategorySecurity},
		{"title case", "Staticcheck", CategoryCorrectness},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CategoryForLinter(tt.input)
			if got != tt.expect {
				t.Errorf("CategoryForLinter(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestRegisterLinterCategory(t *testing.T) {
	t.Parallel()

	original := CategoryForLinter("my-custom-linter")
	if original != CategoryCorrectness {
		t.Fatalf("unexpected pre-registration category: %q", original)
	}

	RegisterLinterCategory("my-custom-linter", CategorySecurity)

	got := CategoryForLinter("my-custom-linter")
	if got != CategorySecurity {
		t.Errorf("after registration: CategoryForLinter(%q) = %q, want %q", "my-custom-linter", got, CategorySecurity)
	}

	RegisterLinterCategory("my-custom-linter", CategoryCorrectness)
}

func TestRegisterLinterCategoryOverride(t *testing.T) {
	t.Parallel()

	original := CategoryForLinter("govet")
	if original != CategoryCorrectness {
		t.Fatalf("unexpected pre-registration category for govet: %q", original)
	}

	RegisterLinterCategory("govet", CategorySecurity)

	got := CategoryForLinter("govet")
	if got != CategorySecurity {
		t.Errorf("after override: CategoryForLinter(%q) = %q, want %q", "govet", got, CategorySecurity)
	}

	RegisterLinterCategory("govet", CategoryCorrectness)
}
