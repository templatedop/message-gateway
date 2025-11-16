package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"MgApplication/api-server/router-adapter"
	log "MgApplication/api-log"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RecoveryConfig configures the recovery middleware
type RecoveryConfig struct {
	// EnableStackTrace enables detailed stack traces in debug mode
	EnableStackTrace bool

	// StackTraceHandler is called with the stack trace when a panic occurs
	// If nil, logs to default logger
	StackTraceHandler func(stack string)
}

// DefaultRecoveryConfig returns default recovery configuration
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		EnableStackTrace:  true,
		StackTraceHandler: nil,
	}
}

// Recovery returns a middleware that recovers from panics and returns 500 error
func Recovery(config ...RecoveryConfig) routeradapter.MiddlewareFunc {
	cfg := DefaultRecoveryConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				stack := string(debug.Stack())

				// Log panic
				logger := log.GetBaseLoggerInstance().ToZerolog()
				logger.Error().Msgf("Panic recovered: %v", r)

				// Log stack trace if enabled
				if cfg.EnableStackTrace {
					if cfg.StackTraceHandler != nil {
						cfg.StackTraceHandler(stack)
					} else {
						// Log stack trace line by line to avoid JSON-escaped newlines
						logger.Error().Msg("Stack trace:")
						for _, line := range splitLines(stack) {
							if line != "" {
								logger.Error().Msg("  " + line)
							}
						}
					}
				}

				// Record error in OpenTelemetry span if available
				span := trace.SpanFromContext(ctx.Context())
				if span.SpanContext().IsValid() {
					span.RecordError(fmt.Errorf("panic: %v", r))
					span.SetStatus(codes.Error, "panic recovered")
					span.SetAttributes(
						attribute.String("panic.value", fmt.Sprintf("%v", r)),
						attribute.String("panic.stack", stack),
					)
				}

				// Return 500 error response
				_ = ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error": "internal server error",
					"code":  "500",
				})
			}
		}()

		return next()
	}
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
