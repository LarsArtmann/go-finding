package finding

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindingErrorError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *FindingError
		expected string
	}{
		{
			name:     "with cause",
			err:      NewValidationError("invalid finding", errors.New("missing ID")),
			expected: "[validation] invalid finding: missing ID",
		},
		{
			name:     "without cause",
			err:      NewIOError("file not found", nil),
			expected: "[io] file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFindingErrorUnwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("underlying error")
	err := NewValidationError("validation failed", cause)

	assert.ErrorIs(t, err, cause)
}

func TestFindingErrorWithFinding(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:       "test:1",
		Position: Position{File: "test.go", Line: 10, Column: 5},
	}

	err := NewValidationError("invalid finding", nil).WithFinding(f)

	assert.NotNil(t, err.Finding)
	assert.Equal(t, f.ID, err.Finding.ID)

	assertFindingErrorFile(t, err, "test.go")

	assert.NotNil(t, err.Position)
	assert.Equal(t, 10, err.Position.Line)
}

func TestFindingErrorWithPosition(t *testing.T) {
	t.Parallel()

	pos := Position{File: "test.go", Line: 20, Column: 10}
	err := NewIOError("read failed", nil).WithPosition(pos)

	assert.NotNil(t, err.Position)
	assert.Equal(t, 20, err.Position.Line)

	assertFindingErrorFile(t, err, "test.go")
}

func TestIsFindingError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "FindingError",
			err:      NewValidationError("test", nil),
			expected: true,
		},
		{
			name:     "wrapped FindingError",
			err:      fmt.Errorf("wrapped: %w", NewValidationError("test", nil)),
			expected: true,
		},
		{
			name:     "regular error",
			err:      errors.New("regular"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsFindingError(tt.err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		{
			name:     "validation error",
			err:      NewValidationError("test", nil),
			expected: ErrCategoryValidation,
		},
		{
			name:     "io error",
			err:      NewIOError("test", nil),
			expected: ErrCategoryIO,
		},
		{
			name:     "regular error",
			err:      errors.New("regular"),
			expected: "",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetCategory(tt.err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsCategory(t *testing.T) {
	t.Parallel()

	err := NewValidationError("test", nil)

	assert.True(t, IsCategory(err, ErrCategoryValidation))
	assert.False(t, IsCategory(err, ErrCategoryIO))
	assert.False(t, IsCategory(errors.New("regular"), ErrCategoryValidation))
}

func TestErrorCategoryConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *FindingError
		expected ErrorCategory
	}{
		{"validation", NewValidationError("test", nil), ErrCategoryValidation},
		{"io", NewIOError("test", nil), ErrCategoryIO},
		{"parse", NewParseError("test", nil), ErrCategoryParse},
		{"conflict", NewConflictError("test", nil), ErrCategoryConflict},
		{"internal", NewInternalError("test", nil), ErrCategoryInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.err.Category)
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    *FindingError
		target error
		want   bool
	}{
		{"validation matches", NewValidationError("test", nil), ErrValidation, true},
		{"io matches", NewIOError("test", nil), ErrIO, true},
		{"parse matches", NewParseError("test", nil), ErrParse, true},
		{"conflict matches", NewConflictError("test", nil), ErrConflict, true},
		{"internal matches", NewInternalError("test", nil), ErrInternal, true},
		{"validation wrong sentinel", NewValidationError("test", nil), ErrIO, false},
		{"io wrong sentinel", NewIOError("test", nil), ErrParse, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, errors.Is(tt.err, tt.target))
		})
	}
}

func TestSentinelErrors_Wrapped(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("wrapped: %w", NewValidationError("test", nil))
	assert.ErrorIs(t, err, ErrValidation)
}

func TestFindingError_Is_UnknownCategory(t *testing.T) {
	t.Parallel()

	err := &FindingError{Category: ErrorCategory("custom"), Message: "custom error"}
	assert.NotErrorIs(t, err, ErrValidation)
	assert.NotErrorIs(t, err, ErrInternal)
}
