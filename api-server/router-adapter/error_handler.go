package routeradapter

import (
	"fmt"
	"net/http"

	apierrors "MgApplication/api-errors"
)

// ErrorHandler handles errors in a framework-agnostic way
// Already defined in adapter.go as interface, this file contains implementations

// DefaultErrorHandler is a basic error handler that works with all frameworks
// It sends JSON error responses using the apierrors package
type DefaultErrorHandler struct{}

// HandleError implements ErrorHandler interface
func (h *DefaultErrorHandler) HandleError(ctx *RouterContext, err error) {
	if err == nil {
		return
	}

	// Check if response was already written
	if ctx.IsResponseWritten() {
		// Log error but can't send response
		fmt.Printf("Error after response written: %v\n", err)
		return
	}

	// Determine status code and message
	statusCode := http.StatusInternalServerError
	message := err.Error()
	errorCode := "internal_error"

	// Check if it's an API error with custom message and code
	if apiErr, ok := err.(*apierrors.AppError); ok {
		if apiErr.Message != "" {
			message = apiErr.Message
		}
		if apiErr.Code != "" {
			errorCode = apiErr.Code
		}
	}

	// Send JSON error response
	errorResponse := map[string]interface{}{
		"error":   message,
		"code":    errorCode,
		"status":  statusCode,
		"success": false,
	}

	// Ignore error from JSON encoding - if that fails, we can't do much
	_ = ctx.JSON(statusCode, errorResponse)
}

// GinErrorHandler handles errors for Gin framework
// Gin uses middleware-based error handling where errors are collected
// and handled at the end of the request chain
type GinErrorHandler struct {
	DefaultErrorHandler
}

// FiberErrorHandler handles errors for Fiber framework
// Fiber uses centralized error handling via app.config.ErrorHandler
type FiberErrorHandler struct {
	DefaultErrorHandler
}

// EchoErrorHandler handles errors for Echo framework
// Echo uses centralized error handling via echo.HTTPErrorHandler
type EchoErrorHandler struct {
	DefaultErrorHandler
}

// NetHTTPErrorHandler handles errors for net/http
// net/http requires manual error handling in each handler
type NetHTTPErrorHandler struct {
	DefaultErrorHandler
}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler() ErrorHandler {
	return &DefaultErrorHandler{}
}

// NewGinErrorHandler creates an error handler for Gin
func NewGinErrorHandler() ErrorHandler {
	return &GinErrorHandler{}
}

// NewFiberErrorHandler creates an error handler for Fiber
func NewFiberErrorHandler() ErrorHandler {
	return &FiberErrorHandler{}
}

// NewEchoErrorHandler creates an error handler for Echo
func NewEchoErrorHandler() ErrorHandler {
	return &EchoErrorHandler{}
}

// NewNetHTTPErrorHandler creates an error handler for net/http
func NewNetHTTPErrorHandler() ErrorHandler {
	return &NetHTTPErrorHandler{}
}
