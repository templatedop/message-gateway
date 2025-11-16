package router

import (
	"github.com/gin-gonic/gin"
	//log "MgApplication/api-log"
	trace "MgApplication/api-trace"

	oteltrace "go.opentelemetry.io/otel/trace"
)

const TracerName = "gin-server"

type CtxRequestIdKey struct{}

func CtxRequestId(c *gin.Context) string {

	if rid, ok := c.Request.Context().Value(CtxRequestIdKey{}).(string); ok {
		return rid
	} else {
		return ""
	}
}

// func CtxLogger(c *gin.Context) *log.Logger {
// 	return log.GetBaseLoggerInstance().ToZerolog().
// }

func CtxTracer(c *gin.Context) oteltrace.Tracer {
	return trace.CtxTracerProvider(c.Request.Context()).Tracer(TracerName)
}
