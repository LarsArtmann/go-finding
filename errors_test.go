package finding

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
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
			g := NewWithT(t)

			got := tt.err.Error()
			g.Expect(got).To(Equal(tt.expected))
		})
	}
}

func TestFindingErrorUnwrap(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cause := errors.New("underlying error")
	err := NewValidationError("validation failed", cause)

	g.Expect(errors.Is(err, cause)).To(BeTrue())
}

func TestFindingErrorWithFinding(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{
		ID:       "test:1",
		Position: Position{File: "test.go", Line: 10, Column: 5},
	}

	err := NewValidationError("invalid finding", nil).WithFinding(f)

	g.Expect(err.Finding).NotTo(BeNil())
	g.Expect(err.Finding.ID).To(Equal(f.ID))

	assertFindingErrorFile(t, err, "test.go")

	g.Expect(err.Position).NotTo(BeNil())
	g.Expect(err.Position.Line).To(Equal(10))
}

func TestFindingErrorWithPosition(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	pos := Position{File: "test.go", Line: 20, Column: 10}
	err := NewIOError("read failed", nil).WithPosition(pos)

	g.Expect(err.Position).NotTo(BeNil())
	g.Expect(err.Position.Line).To(Equal(20))

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
			g := NewWithT(t)

			got := IsFindingError(tt.err)
			g.Expect(got).To(Equal(tt.expected))
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
			g := NewWithT(t)

			got := CategoryOf(tt.err)
			g.Expect(got).To(Equal(tt.expected))
		})
	}
}

func TestIsCategory(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	err := NewValidationError("test", nil)

	g.Expect(IsCategory(err, ErrCategoryValidation)).To(BeTrue())
	g.Expect(IsCategory(err, ErrCategoryIO)).To(BeFalse())
	g.Expect(IsCategory(errors.New("regular"), ErrCategoryValidation)).To(BeFalse())
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
			g := NewWithT(t)

			g.Expect(tt.err.Category).To(Equal(tt.expected))
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
			g := NewWithT(t)

			g.Expect(errors.Is(tt.err, tt.target)).To(Equal(tt.want))
		})
	}
}

func TestSentinelErrors_Wrapped(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	err := fmt.Errorf("wrapped: %w", NewValidationError("test", nil))
	g.Expect(errors.Is(err, ErrValidation)).To(BeTrue())
}

func TestFindingError_Is_UnknownCategory(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	err := &FindingError{Category: ErrorCategory("custom"), Message: "custom error"}
	g.Expect(errors.Is(err, ErrValidation)).To(BeFalse())
	g.Expect(errors.Is(err, ErrInternal)).To(BeFalse())
}
