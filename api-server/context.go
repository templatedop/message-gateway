package router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	trace "MgApplication/api-trace"

	oteltrace "go.opentelemetry.io/otel/trace"
)

const TracerName = "gin-server"

type CtxRequestIdKey struct{}

// CtxRequestId retrieves the request ID from the context or generates a new one if missing.
// It also propagates the request ID back to the context and response headers for traceability.
func CtxRequestId(c *gin.Context) string {
	// First, check if request ID exists in context
	if rid, ok := c.Request.Context().Value(CtxRequestIdKey{}).(string); ok && rid != "" {
		return rid
	}

	// Second, check if request ID is in the header (X-Request-Id)
	rid := c.GetHeader("X-Request-Id")
	if rid == "" {
		// Generate a new request ID if not found
		rid = uuid.New().String()
	}

	// Store the request ID in context for downstream handlers
	ctx := context.WithValue(c.Request.Context(), CtxRequestIdKey{}, rid)
	c.Request = c.Request.WithContext(ctx)

	// Set the request ID in response header for client traceability
	c.Writer.Header().Set("X-Request-Id", rid)

	return rid
}

// func CtxLogger(c *gin.Context) *log.Logger {
// 	return log.GetBaseLoggerInstance().ToZerolog().
// }

func CtxTracer(c *gin.Context) oteltrace.Tracer {
	return trace.CtxTracerProvider(c.Request.Context()).Tracer(TracerName)
}
