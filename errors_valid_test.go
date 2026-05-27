package finding

import (
	"testing"
)

func TestErrorCategory_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cat  ErrorCategory
		want bool
	}{
		{ErrCategoryValidation, true},
		{ErrCategoryIO, true},
		{ErrCategoryConflict, true},
		{ErrCategoryInternal, true},
		{ErrorCategory("bogus"), true},
		{ErrorCategory("some-category"), true},
		{ErrorCategory(""), false},
		{ErrorCategory("Bogus"), false},
		{ErrorCategory("HAS_SPACE"), false},
		{ErrorCategory("has space"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.cat), func(t *testing.T) {
			t.Parallel()
			if tt.cat.IsValid() != tt.want {
				t.Errorf("IsValid() = %v, want %v", !tt.want, tt.want)
			}
		})
	}
}
