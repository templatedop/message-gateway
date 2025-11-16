package middlewares

import (
	"context"
	"time"

	"MgApplication/api-server/router-adapter"
	log "MgApplication/api-log"
)

// ctxLoggerKey is the context key for storing the logger
type ctxLoggerKey struct{}

// SetCtxLogger returns a middleware that sets up a context-aware logger
// This middleware extracts request metadata and creates a logger with that context
func SetCtxLogger() routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Get base logger
		baseLogger := log.GetBaseLoggerInstance()
		if baseLogger == nil {
			return next()
		}

		// Create logger with request metadata
		// Note: This is simplified - in production, you'd extract more metadata
		// from the RouterContext (like request ID, trace ID, etc.)

		// Store logger in context
		reqCtx := context.WithValue(ctx.Request.Context(), ctxLoggerKey{}, baseLogger)
		ctx.Request = ctx.Request.WithContext(reqCtx)

		return next()
	}
}

// RequestResponseLogger returns a middleware that logs HTTP requests and responses
func RequestResponseLogger() routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Skip healthz endpoint
		if ctx.Request.Method == "GET" && ctx.Request.URL.Path == "/healthz" {
			return next()
		}

		// Record start time
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery
		method := ctx.Request.Method

		// Execute next middleware/handler
		err := next()

		// Calculate duration
		duration := time.Since(start)

		// Build full path
		fullPath := path
		if raw != "" {
			fullPath = path + "?" + raw
		}

		// Get status code
		status := ctx.StatusCode()
		if status == 0 {
			status = 200
		}

		// Log request/response
		logger := log.GetBaseLoggerInstance()
		if logger != nil {
			zl := logger.ToZerolog()

			// Determine log level based on status code
			switch {
			case status >= 500:
				zl.Error().
					Str("method", method).
					Str("path", fullPath).
					Int("status", status).
					Dur("duration", duration).
					Msg("HTTP Request")
			case status >= 400:
				zl.Warn().
					Str("method", method).
					Str("path", fullPath).
					Int("status", status).
					Dur("duration", duration).
					Msg("HTTP Request")
			default:
				zl.Info().
					Str("method", method).
					Str("path", fullPath).
					Int("status", status).
					Dur("duration", duration).
					Msg("HTTP Request")
			}
		}

		return err
	}
}
