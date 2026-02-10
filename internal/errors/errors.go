// Package errors provides custom error types for the application.
package errors

import (
	"errors"
	"fmt"
)

// Common errors that can be checked with errors.Is.
var (
	ErrProjectExists = errors.New("project already exists")
)

// ScaffoldError represents an error during scaffolding.
type ScaffoldError struct {
	Op  string // operation that failed
	Err error  // underlying error
}

func (e *ScaffoldError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("scaffold %s: %v", e.Op, e.Err)
	}
	return e.Err.Error()
}

func (e *ScaffoldError) Unwrap() error {
	return e.Err
}

// NewScaffoldError creates a new scaffold error.
func NewScaffoldError(op string, err error) *ScaffoldError {
	return &ScaffoldError{Op: op, Err: err}
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
	}
	return e.Message
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}
