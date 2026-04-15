package finding

import (
	"errors"
	"fmt"
)

// ErrorCategory categorizes errors for programmatic handling.
type ErrorCategory string

const (
	// ErrCategoryValidation indicates validation errors.
	ErrCategoryValidation ErrorCategory = "validation"
	// ErrCategoryIO indicates file system or network errors.
	ErrCategoryIO ErrorCategory = "io"
	// ErrCategoryParse indicates parsing errors.
	ErrCategoryParse ErrorCategory = "parse"
	// ErrCategoryConflict indicates conflicting operations.
	ErrCategoryConflict ErrorCategory = "conflict"
	// ErrCategoryInternal indicates internal logic errors.
	ErrCategoryInternal ErrorCategory = "internal"
)

// IsValid returns true if the error category is a non-empty string.
func (c ErrorCategory) IsValid() bool {
	return c != ""
}

// FindingError provides structured error information with context.
type FindingError struct {
	Category ErrorCategory // Category of error
	Finding  *Finding      // Associated finding (may be nil)
	Message  string        // Human-readable message
	Cause    error         // Underlying cause (may be nil)
	File     string        // File path (if applicable)
	Position *Position     // Position in file (if applicable)
}

// Error implements the error interface.
func (e *FindingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Category, e.Message, e.Cause)
	}

	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

// Unwrap returns the underlying cause for error inspection.
func (e *FindingError) Unwrap() error {
	return e.Cause
}

// IsCategory returns true if the error matches the given category.
func (e *FindingError) IsCategory(cat ErrorCategory) bool {
	return e.Category == cat
}

// WithFinding sets the finding on a copy of the FindingError and returns it.
func (e *FindingError) WithFinding(f Finding) *FindingError {
	clone := *e
	clone.Finding = &f
	clone.File = f.Position.File
	clone.Position = &f.Position

	return &clone
}

// WithPosition sets the position on a copy of the FindingError and returns it.
func (e *FindingError) WithPosition(pos Position) *FindingError {
	clone := *e
	clone.Position = &pos
	clone.File = pos.File

	return &clone
}

// NewValidationError creates a validation error.
func NewValidationError(message string, cause error) *FindingError {
	return &FindingError{
		Category: ErrCategoryValidation,
		Message:  message,
		Cause:    cause,
	}
}

// NewIOError creates an IO error.
func NewIOError(message string, cause error) *FindingError {
	return &FindingError{
		Category: ErrCategoryIO,
		Message:  message,
		Cause:    cause,
	}
}

// NewParseError creates a parse error.
func NewParseError(message string, cause error) *FindingError {
	return &FindingError{
		Category: ErrCategoryParse,
		Message:  message,
		Cause:    cause,
	}
}

// NewConflictError creates a conflict error.
func NewConflictError(message string, cause error) *FindingError {
	return &FindingError{
		Category: ErrCategoryConflict,
		Message:  message,
		Cause:    cause,
	}
}

// NewInternalError creates an internal error.
func NewInternalError(message string, cause error) *FindingError {
	return &FindingError{
		Category: ErrCategoryInternal,
		Message:  message,
		Cause:    cause,
	}
}

// IsFindingError returns true if err is a *FindingError.
func IsFindingError(err error) bool {
	_, ok := errors.AsType[*FindingError](err)
	return ok
}

// GetCategory returns the category of the error, or empty string if not a FindingError.
func GetCategory(err error) ErrorCategory {
	if findingErr, ok := errors.AsType[*FindingError](err); ok {
		return findingErr.Category
	}

	return ""
}

// IsCategory returns true if err is a FindingError with the given category.
func IsCategory(err error, cat ErrorCategory) bool {
	return GetCategory(err) == cat
}
