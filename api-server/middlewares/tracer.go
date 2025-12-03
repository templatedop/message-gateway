package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"MgApplication/api-server/middlewares/reqid"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// RequestTracerMiddlewareConfig is the configuration for the RequestTracerMiddleware
type RequestTracerMiddlewareConfig struct {
	Skipper                     func(*gin.Context) bool
	TracerProvider              oteltrace.TracerProvider
	TextMapPropagator           propagation.TextMapPropagator
	RequestUriPrefixesToExclude []string
}

func RequestTracerMiddleware(servicename string, config RequestTracerMiddlewareConfig) gin.HandlerFunc {

	if config.Skipper == nil {
		config.Skipper = func(*gin.Context) bool { return false }
	}
	if config.TracerProvider == nil {
		config.TracerProvider = otel.GetTracerProvider()
	}
	if config.TextMapPropagator == nil {
		config.TextMapPropagator = otel.GetTextMapPropagator()
	}

	tracer := config.TracerProvider.Tracer(servicename)

	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}

		// Skip tracing if request URI matches excluded prefixes
		for _, prefix := range config.RequestUriPrefixesToExclude {
			if prefix != "" && strings.HasPrefix(c.Request.RequestURI, prefix) {
				c.Next()
				return
			}
		}

		ctx := c.Request.Context()

		// Generate or extract request ID
		requestID := c.GetHeader("Traceparent")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store request ID in context
		ctx = context.WithValue(ctx, reqid.CtxRequestIdKey{}, requestID)

		// Extract trace context from headers
		carrier := propagation.HeaderCarrier(c.Request.Header)
		ctx = config.TextMapPropagator.Extract(ctx, carrier)

		// Inject trace context into response headers
		responseHeaders := propagation.HeaderCarrier(c.Writer.Header())
		config.TextMapPropagator.Inject(ctx, responseHeaders)

		// Build span name
		spanName := c.FullPath()
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		}

		// Build span options
		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(httpconv.ServerRequest("", c.Request)...),
			oteltrace.WithAttributes(semconv.HTTPRoute(c.FullPath())),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		// Start span - single creation point
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// Set trace ID headers
		traceID := span.SpanContext().TraceID().String()
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		status := c.Writer.Status()
		attrs := semconv.HTTPStatusCode(status)
		span.SetAttributes(attrs)

		if status >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d: %s", status, http.StatusText(status)))

		}
	}
}
