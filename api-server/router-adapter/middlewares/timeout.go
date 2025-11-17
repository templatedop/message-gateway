package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "MgApplication/api-log"
	"MgApplication/api-server/router-adapter"
)

// Timeout returns a middleware that sets a timeout for request processing
// If the timeout is exceeded, it returns 504 Gateway Timeout
func Timeout(timeout time.Duration) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Create timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), timeout)
		defer cancel()

		// Update request context
		ctx.Request = ctx.Request.WithContext(timeoutCtx)

		// Channel to signal completion
		done := make(chan error, 1)
		var handlerError error

		// Run handler in goroutine
		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					handlerError = fmt.Errorf("handler panic: %v", r)
					log.Error(timeoutCtx, "Handler PANIC recovered: %v\n", r)
				}
			}()

			// Execute next middleware/handler
			err := next()
			done <- err
		}()

		// Wait for completion or timeout
		select {
		case err := <-done:
			// Handler completed normally
			if handlerError != nil && !ctx.IsResponseWritten() {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
			}
			return err

		case <-timeoutCtx.Done():
			// Timeout occurred
			if timeoutCtx.Err() == context.DeadlineExceeded {
				return ctx.JSON(http.StatusGatewayTimeout, map[string]string{
					"error":   "request timeout",
					"timeout": timeout.String(),
				})
			}
			return timeoutCtx.Err()
		}
	}
}
