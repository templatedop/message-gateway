package apierrors

import (
	"fmt"
	"runtime"
	"strings"
)

// stackTrace represents a series of stack frames collected during an error.
type stackTrace struct {
	Frames []stackFrame `json:"frames"`
}

// stackFrame represents a single frame in a stack trace.
// It contains the function name, file name, line number, pointer address, and function pointer reference.
type stackFrame struct {
	Function        string  `json:"function"`
	File            string  `json:"file"`
	Line            int     `json:"line"`
	Pointer         uintptr `json:"pointer"`
	FunctionPointer uintptr `json:"function_pointer"` // New field: pointer reference to the function.
}

// collectStackTrace captures the current stack trace starting from the caller
// of this function. The number of frames collected is controlled by the global
// StackTraceConfig.MaxDepth setting. It converts the program counters into a slice
// of stackFrame, which includes the function name, file name, line number,
// pointer address, and function pointer for each frame. The collected stack trace
// is returned as a pointer to a stackTrace struct.
//
// If stack traces are disabled globally (via StackTraceConfig.Enabled), this returns nil.
func collectStackTrace() *stackTrace {
	config := GetStackTraceConfig()

	// If stack traces are disabled, return nil immediately (zero overhead)
	if !config.Enabled || config.MaxDepth <= 0 {
		return nil
	}

	// Allocate buffer based on configured max depth
	// Use a reasonable maximum to prevent excessive memory allocation
	maxDepth := config.MaxDepth
	if maxDepth > 100 {
		maxDepth = 100 // Cap at 100 frames for safety
	}

	// Allocate slice with exact size needed
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(3, pcs)

	// If no frames were captured, return nil
	if n == 0 {
		return nil
	}

	// Convert program counters to stack frames
	frames := make([]stackFrame, 0, n)
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			file, line := fn.FileLine(pc)
			frames = append(frames, stackFrame{
				Function:        fn.Name(),
				File:            file,
				Line:            line,
				Pointer:         pc,
				FunctionPointer: fn.Entry(), // Include the pointer to the start of the function.
			})
		}
	}

	return &stackTrace{Frames: frames}
}

// String returns a formatted string representation of the stack trace.
// The output includes the function names, their respective file locations,
// and pointer addresses for each frame in the stack trace, separated by a line of dashes.
func (st *stackTrace) String() string {
	var sb strings.Builder
	sb.WriteString("Stack Trace:\n")
	sb.WriteString("****************STACK TRACE START**********************")
	sb.WriteString("\n")
	for i, frame := range st.Frames {
		sb.WriteString(fmt.Sprintf("#%d - Function: %s\n", i+1, frame.Function))
		sb.WriteString(fmt.Sprintf("     Location: %s:%d\n", frame.File, frame.Line))
		sb.WriteString(fmt.Sprintf("     Pointer: 0x%x\n", frame.Pointer))
		sb.WriteString(fmt.Sprintf("     Function Pointer: 0x%x\n", frame.FunctionPointer)) // Include the function pointer reference.
		sb.WriteString("\n")
		sb.WriteString("---------------------------------------------------\n")
	}
	sb.WriteString("****************STACK TRACE END**********************")
	return sb.String()
}

// collectStackTraceConditional collects a stack trace only if the configuration
// allows it for the given HTTP status code.
//
// This function respects the CollectFor5xxOnly configuration:
//   - If CollectFor5xxOnly is true, stack traces are only collected for status codes >= 500
//   - If CollectFor5xxOnly is false, stack traces are collected for all errors (when enabled)
//
// Returns nil if stack traces should not be collected for this error.
func collectStackTraceConditional(statusCode int) *stackTrace {
	// Check if we should collect stack trace for this status code
	if !shouldCollectStackTrace(statusCode) {
		return nil
	}

	// Collect and return the stack trace
	return collectStackTrace()
}

// enhanceWithCause appends the frames from the original stackTrace to the current stackTrace.
// This method is used to enhance the current stackTrace with additional context from another stackTrace.
func (st *stackTrace) enhanceWithCause(original *stackTrace) {
	if st != nil && original != nil {
		st.Frames = append(st.Frames, original.Frames...)
	}
}
