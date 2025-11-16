package middlewares

import (
	"fmt"
	"net/http"

	"MgApplication/api-server/router-adapter"
)

// BodyLimiter returns a middleware that limits request body size
// If the body exceeds the limit, it returns 413 Request Entity Too Large
func BodyLimiter(limit int64) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Wrap the request body with a size limiter
		ctx.Request.Body = http.MaxBytesReader(ctx.Response, ctx.Request.Body, limit)

		// Call next middleware/handler
		err := next()

		// Check if error is due to body size limit
		if err != nil && err.Error() == "http: request body too large" {
			return ctx.JSON(http.StatusRequestEntityTooLarge, map[string]string{
				"error": "request body too large",
				"limit": formatBytes(limit),
			})
		}

		return err
	}
}

// formatBytes formats bytes into human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return fmt.Sprintf("%d %s", bytes/div, units[exp])
}
