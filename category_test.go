package finding

import (
	"slices"
	"testing"
)

func TestCategory_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cat  Category
		want bool
	}{
		{CategorySecurity, true},
		{CategoryStyle, true},
		{Category(""), false},
		{Category("custom"), true},
		{Category("go-vet"), true},
		{Category("Security"), false},
		{Category("some_thing"), false},
		{Category("UPPERCASE"), false},
		{Category("has space"), false},
	}

	for _, tt := range tests {
		if got := tt.cat.IsValid(); got != tt.want {
			t.Errorf("Category(%q).IsValid() = %v, want %v", tt.cat, got, tt.want)
		}
	}
}

func TestCategory_IsStandard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cat  Category
		want bool
	}{
		{CategorySecurity, true},
		{CategoryStyle, true},
		{CategoryPerformance, true},
		{CategoryCorrectness, true},
		{CategoryComplexity, true},
		{CategoryDuplication, true},
		{CategoryErrorHandling, true},
		{CategoryMigration, true},
		{CategoryTypeSafety, true},
		{CategoryStructure, true},
		{CategoryConfiguration, true},
		{CategoryDocumentation, true},
		{CategoryTesting, true},
		{Category(""), false},
		{Category("custom"), false},
		{Category("Security"), false},
	}

	for _, tt := range tests {
		if got := tt.cat.IsStandard(); got != tt.want {
			t.Errorf("Category(%q).IsStandard() = %v, want %v", tt.cat, got, tt.want)
		}
	}
}

func TestCategory_Constants(t *testing.T) {
	t.Parallel()

	if CategorySecurity != "security" {
		t.Errorf("CategorySecurity = %q, want %q", CategorySecurity, "security")
	}

	if CategoryErrorHandling != "error-handling" {
		t.Errorf("CategoryErrorHandling = %q, want %q", CategoryErrorHandling, "error-handling")
	}
}

func TestCategory_IsSecurity(t *testing.T) {
	t.Parallel()

	if !CategorySecurity.IsSecurity() {
		t.Error("CategorySecurity.IsSecurity() = false, want true")
	}

	if CategoryStyle.IsSecurity() {
		t.Error("CategoryStyle.IsSecurity() = true, want false")
	}

	if Category("").IsSecurity() {
		t.Error("empty Category.IsSecurity() = true, want false")
	}
}

func TestParseCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  Category
		err   bool
	}{
		{"security", CategorySecurity, false},
		{"style", CategoryStyle, false},
		{"correctness", CategoryCorrectness, false},
		{"performance", CategoryPerformance, false},
		{"complexity", CategoryComplexity, false},
		{"error-handling", CategoryErrorHandling, false},
		{"type-safety", CategoryTypeSafety, false},
		{"custom-category", Category("custom-category"), false},
		{"", Category(""), true},
		{"Security", Category(""), true},
		{"has space", Category(""), true},
		{"UPPERCASE", Category(""), true},
		{"under_score", Category(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, err := ParseCategory(tt.input)
			if tt.err {
				if err == nil {
					t.Errorf("ParseCategory(%q) = %q, expected error", tt.input, got)
				}
			} else {
				if err != nil {
					t.Errorf("ParseCategory(%q) unexpected error: %v", tt.input, err)
				}

				if got != tt.want {
					t.Errorf("ParseCategory(%q) = %q, want %q", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestMustParseCategory(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		got := MustParseCategory("security")
		if got != CategorySecurity {
			t.Errorf("MustParseCategory(%q) = %q, want %q", "security", got, CategorySecurity)
		}
	})

	t.Run("panics on invalid", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("MustParseCategory(%q) expected panic", "INVALID")
			}
		}()

		MustParseCategory("INVALID")
	})
}

func TestCategory_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		c     Category
		other Category
		want  int
	}{
		{"less than", CategoryCorrectness, CategorySecurity, -1},
		{"equal", CategorySecurity, CategorySecurity, 0},
		{"greater than", CategoryStyle, CategorySecurity, 1},
		{"empty less than", Category(""), CategorySecurity, -1},
		{"both empty equal", Category(""), Category(""), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.c.Compare(tt.other); got != tt.want {
				t.Errorf("Category(%q).Compare(%q) = %d, want %d", tt.c, tt.other, got, tt.want)
			}
		})
	}
}

func TestCategory_Compare_SortStable(t *testing.T) {
	t.Parallel()

	cats := []Category{
		CategorySecurity,
		CategoryCorrectness,
		CategoryStyle,
		CategoryDocumentation,
		CategoryBestPractice,
	}
	want := []Category{
		CategoryBestPractice,
		CategoryCorrectness,
		CategoryDocumentation,
		CategorySecurity,
		CategoryStyle,
	}

	slices.SortFunc(cats, func(a, b Category) int { return a.Compare(b) })

	for i, c := range cats {
		if c != want[i] {
			t.Fatalf("index %d: got %q, want %q", i, c, want[i])
		}
	}
}
