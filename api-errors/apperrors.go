package apierrors

import (
	"bytes"
	"fmt"
)

// AppError represents a structured error with additional context information.
// It includes a stack trace, a user-facing message, an error code, the original error,
// and any field-specific errors that may have occurred.
type AppError struct {
	ID            string       `json:"id,omitempty"`
	Code          int          `json:"code"`
	Message       string       `json:"message"`
	FieldErrors   []FieldError `json:"field_errors,omitempty"`
	Stack         *stackTrace  `json:"-"`
	OriginalError error        `json:"-"`
}

// FieldError represents an error related to a specific field in a request or response.
// It contains information about the field name, the value that caused the error,
// and a message describing the error.
type FieldError struct {
	Field   string      `json:"field"`
	Tag     string      `json:"-"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

// NewAppError creates a new instance of AppError with the provided message, code, and original error.
// It conditionally collects a stack trace based on the global StackTraceConfig settings.
//
// Stack trace collection behavior:
//   - If StackTraceConfig.Enabled is false, no stack trace is collected (zero overhead).
//   - If StackTraceConfig.CollectFor5xxOnly is true, stack traces are only collected for status codes >= 500.
//   - Stack trace depth is limited by StackTraceConfig.MaxDepth.
//
// Parameters:
//   - message: A string representing the error message.
//   - code: An integer representing the HTTP status code (used to determine if stack trace should be collected).
//   - originalError: The original error that caused this error.
//
// Returns:
//
//	An instance of AppError containing the provided details and optionally a stack trace.
func NewAppError(message string, code int, originalError error) AppError {
	return AppError{
		Stack:         collectStackTraceConditional(code),
		Message:       message,
		Code:          code,
		OriginalError: originalError,
	}
}

// NewAppErrorWithId creates a new instance of AppError with the provided message, code, original error, and id.
// It conditionally collects a stack trace based on the global StackTraceConfig settings.
//
// Stack trace collection behavior:
//   - If StackTraceConfig.Enabled is false, no stack trace is collected (zero overhead).
//   - If StackTraceConfig.CollectFor5xxOnly is true, stack traces are only collected for status codes >= 500.
//   - Stack trace depth is limited by StackTraceConfig.MaxDepth.
//
// Parameters:
//   - message: A string representing the error message.
//   - code: An integer representing the HTTP status code (used to determine if stack trace should be collected).
//   - originalError: The original error that caused this error.
//   - id: A string representing the unique identifier for the error.
//
// Returns:
//
//	An instance of AppError containing the provided details and optionally a stack trace.
func NewAppErrorWithId(message string, code int, originalError error, id string) AppError {
	return AppError{
		Stack:         collectStackTraceConditional(code),
		Message:       message,
		Code:          code,
		OriginalError: originalError,
		ID:            id,
	}
}

// SetFieldErrors sets the field errors for the AppError instance.
//
// Parameters:
//
//	fieldErrors ([]FieldError): A slice of FieldError instances to be set.
//
// This method updates the FieldErrors property of the AppError instance with the provided slice of FieldError.
func (ae *AppError) SetFieldErrors(fieldErrors []FieldError) {
	ae.FieldErrors = fieldErrors
}

// AddFieldError appends a field-specific validation error to an existing AppError.
// AddFieldError creates a new FieldError and appends it to a slice of FieldError.
// It takes the following parameters:
// - field: 	The name of the field which failed the validation.
// - value: 	The value of the field which failed the validation.
// - message: 	A descriptive message explaining cause of the validation fail.
// - tag: 		A tag for the validation.
//
// Returns a slice of FieldError containing the newly created FieldError.
func (ae *AppError) NewFieldError(field string, value interface{}, message string, tag string) FieldError {
	return FieldError{
		Field:   field,
		Value:   value,
		Message: message,
		Tag:     tag,
	}
}

// Error returns a formatted error message string that includes the
// custom error message and the original error. It implements the
// error interface.
func (e *AppError) Error() string {
	return fmt.Sprintf("%s. %v", e.Message, e.OriginalError)
}

// Unwrap returns the original error wrapped by the AppError.
// This method allows for unwrapping the underlying error, enabling
// error inspection and comparison.
func (e *AppError) Unwrap() error {
	return e.OriginalError
}

// StackTrace returns the stack trace associated with the AppError.
// It provides detailed information about the sequence of function calls
// that led to the error, which can be useful for debugging purposes.
func (e *AppError) StackTrace() *stackTrace {
	return e.Stack
}

// Pretty returns a formatted string representation of the AppError, including
// the error message, cause, and stack trace if available.
// It returns the output as a bytes.Buffer, which can be used for logging.
func (e *AppError) Pretty() *bytes.Buffer {
	// Create a new buffer to store the formatted error output.
	var buf bytes.Buffer

	// Write the formatted error details into the buffer.
	fmt.Fprintln(&buf, "----------------------------")
	fmt.Fprintln(&buf, "Error:", e.Message)
	if e.OriginalError != nil {
		fmt.Fprintln(&buf, "Cause:", e.OriginalError)
	}
	// Print the stack trace if available.
	if e.Stack != nil {
		fmt.Fprintln(&buf, e.Stack.String())
	} else {
		fmt.Fprintln(&buf, "No stack trace available.")
	}
	fmt.Fprintln(&buf, "----------------------------")

	// Return the buffer.
	return &buf
}
