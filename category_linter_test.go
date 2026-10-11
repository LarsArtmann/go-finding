package finding

import (
	"slices"
	"strings"
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

	// Parallel tests assert built-in lookups against the shared
	// DefaultLinterRegistry, which also persists across -count iterations,
	// so the global function is pinned with a unique name and no pre-state
	// assertion; override-with-known-prestate runs on an isolated registry.
	const uniqueLinter = "test-register-linter-category-unique"

	RegisterLinterCategory(uniqueLinter, CategorySecurity)

	if got := CategoryForLinter(uniqueLinter); got != CategorySecurity {
		t.Errorf("after registration: CategoryForLinter(%q) = %q, want %q", uniqueLinter, got, CategorySecurity)
	}

	r := &LinterRegistry{}
	r.Register("govet", CategoryCorrectness)
	r.Register("govet", CategorySecurity)

	if got := r.Lookup("govet", CategoryCorrectness); got != CategorySecurity {
		t.Errorf("after override: Lookup(%q) = %q, want %q", "govet", got, CategorySecurity)
	}
}

// FuzzCategoryForLinter verifies that CategoryForLinter never panics on
// arbitrary input and always returns a valid Category. It also checks
// case-insensitivity against an isolated registry (not the global one, which
// other tests mutate concurrently).
func FuzzCategoryForLinter(f *testing.F) {
	f.Add("gosec")
	f.Add("GOVET")
	f.Add("")
	f.Add("golangci-lint")
	f.Add("some-very-long-linter-name-with-unicode-→")

	f.Fuzz(func(t *testing.T, name string) {
		// Global lookup must never panic and always return a valid category.
		cat := CategoryForLinter(name)
		if !cat.IsValid() && cat != "" {
			t.Fatalf("CategoryForLinter(%q) = %q: not valid and not empty", name, cat)
		}

		// Case-insensitivity property on an isolated registry.
		reg := NewLinterRegistry(map[string]Category{
			"samplelinter": CategoryPerformance,
		})

		lower := reg.Lookup(strings.ToLower(name), CategoryCorrectness)
		orig := reg.Lookup(name, CategoryCorrectness)

		if lower != orig {
			t.Fatalf("Lookup not case-insensitive for %q: lower=%q orig=%q", name, lower, orig)
		}
	})
}

func TestDefaultLinterRegistry_Expanded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		linter string
		want   Category
	}{
		// Existing entries still work
		{"gosec", CategorySecurity},
		{"govet", CategoryCorrectness},
		{"staticcheck", CategoryCorrectness},

		// New entries added in v1.3.0
		{"gofumpt", CategoryStyle},
		{"nolintlint", CategoryStyle},
		{"depguard", CategoryBestPractice},
		{"nakedret", CategoryCorrectness},
		{"rowserrcheck", CategoryCorrectness},
		{"sqlclosecheck", CategoryCorrectness},
		{"wastedassign", CategoryCorrectness},
		{"bidichk", CategorySecurity},
		{"nosprintfhost", CategorySecurity},
		{"tagliatelle", CategoryTypeSafety},
		{"mirror", CategoryBestPractice},
		{"nilnesserr", CategoryCorrectness},
		{"recvcheck", CategoryCorrectness},
	}

	for _, tt := range tests {
		t.Run(tt.linter, func(t *testing.T) {
			t.Parallel()

			got := CategoryForLinter(tt.linter)
			if got != tt.want {
				t.Errorf("CategoryForLinter(%q) = %v, want %v", tt.linter, got, tt.want)
			}
		})
	}
}

func TestLinterRegistry_Names(t *testing.T) {
	t.Parallel()

	r := &LinterRegistry{}
	if got := r.Names(); len(got) != 0 {
		t.Errorf("empty registry Names() = %v, want empty", got)
	}

	r.Register("zeta", CategoryStyle)
	r.Register("Alpha", CategoryCorrectness)
	r.Register("mid", CategorySecurity)

	want := []string{"alpha", "mid", "zeta"}
	if got := r.Names(); !slices.Equal(got, want) {
		t.Errorf("Names() = %v, want %v (sorted, lowercased keys)", got, want)
	}
}

func TestLinterRegistry_Clone(t *testing.T) {
	t.Parallel()

	r := &LinterRegistry{}
	r.Register("gosec", CategorySecurity)

	clone := r.Clone()
	clone.Register("misspell", CategoryStyle)
	r.Register("dupl", CategoryDuplication)

	if got := r.Lookup("misspell", CategoryUnused); got != CategoryUnused {
		t.Errorf("clone registration leaked into original: got %q", got)
	}

	if got := clone.Lookup("dupl", CategoryUnused); got != CategoryUnused {
		t.Errorf("original registration leaked into clone: got %q", got)
	}

	if got := clone.Lookup("gosec", CategoryUnused); got != CategorySecurity {
		t.Errorf("clone lost original mapping: got %q", got)
	}
}
