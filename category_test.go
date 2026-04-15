package finding

import "testing"

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
		{Category("Security"), true},
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
