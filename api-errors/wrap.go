package apierrors

import (
	"fmt"
	"os"
)

// WrapError wraps an existing error with a new message and enhances the stack trace.
// It collects the stack trace at the point of the original error and where the error is being wrapped,
// then combines them to provide a more comprehensive trace.
//
// Parameters:
//   - err: The original error to be wrapped.
//   - message: A new message to provide additional context to the error.
//
// Returns:
//   - An error of type *AppError that includes the original error, the new message, and the enhanced stack trace.
func WrapError(err error, message string) error {
	// Collect the stack trace at the point of the original error
	originalStackTrace := collectStackTrace()

	// Collect the stack trace where the error is being wrapped
	newStackTrace := collectStackTrace()
	newStackTrace.enhanceWithCause(originalStackTrace)

	// Return a wrappedError containing the new message, original error, and enhanced stack trace
	return &AppError{
		OriginalError: err,
		Message:       message,
		Stack:         newStackTrace,
	}
}

// Decorate takes an error and prints it in a structured, human-readable format.
// Decorate prints detailed information about an error to the standard output.
// If the error is of type *AppError, it prints custom fields such as the message,
// original error, and stack trace if available. Otherwise, it prints the error directly.
//
// Parameters:
//   - err (error): The error to be decorated and printed.
func Decorate(err error) {
	// Check if the error is of type *wrappedError to access custom fields.
	if wrappedErr, ok := err.(*AppError); ok {
		fmt.Fprintln(os.Stdout, "----------------------------")
		fmt.Fprintln(os.Stdout, "Error: ", wrappedErr.Message)
		fmt.Fprintln(os.Stdout, "Cause: ", wrappedErr.OriginalError)

		// Print the stack trace if available
		if wrappedErr.Stack != nil {
			fmt.Fprintln(os.Stdout, wrappedErr.Stack.String())
		} else {
			fmt.Fprintln(os.Stdout, "No stack trace available.")
		}
		fmt.Fprintln(os.Stdout, "----------------------------")
	} else {
		// If the error is not a wrappedError, print the error directly.
		fmt.Fprintln(os.Stdout, "Error: ", err)
	}
}