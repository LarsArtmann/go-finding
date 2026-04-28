package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		{ErrorCategory(""), false},
		{ErrorCategory("bogus"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.cat), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.cat.IsValid())
		})
	}
}
