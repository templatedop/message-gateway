package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"MgApplication/api-server/router-adapter"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracingConfig configures the tracing middleware
type TracingConfig struct {
	// ServiceName is the name of the service for tracing
	ServiceName string

	// Skipper allows skipping tracing for specific requests
	Skipper func(ctx *routeradapter.RouterContext) bool

	// TracerProvider is the OpenTelemetry tracer provider
	// If nil, uses otel.GetTracerProvider()
	TracerProvider oteltrace.TracerProvider

	// TextMapPropagator is used for context propagation
	// If nil, uses otel.GetTextMapPropagator()
	TextMapPropagator propagation.TextMapPropagator

	// ExcludePaths is a list of path prefixes to exclude from tracing
	ExcludePaths []string

	// RequestIDKey is the context key for storing request ID
	RequestIDKey string
}

// DefaultTracingConfig returns default tracing configuration
func DefaultTracingConfig(serviceName string) TracingConfig {
	return TracingConfig{
		ServiceName:       serviceName,
		Skipper:           nil,
		TracerProvider:    nil,
		TextMapPropagator: nil,
		ExcludePaths:      []string{"/healthz", "/metrics"},
		RequestIDKey:      "request-id",
	}
}

// Tracing returns a middleware that implements OpenTelemetry distributed tracing
func Tracing(config TracingConfig) routeradapter.MiddlewareFunc {
	// Set defaults
	if config.Skipper == nil {
		config.Skipper = func(*routeradapter.RouterContext) bool { return false }
	}
	if config.TracerProvider == nil {
		config.TracerProvider = otel.GetTracerProvider()
	}
	if config.TextMapPropagator == nil {
		config.TextMapPropagator = otel.GetTextMapPropagator()
	}
	if config.ServiceName == "" {
		config.ServiceName = "unknown-service"
	}

	tracer := config.TracerProvider.Tracer(config.ServiceName)

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Skip if configured
		if config.Skipper(ctx) {
			return next()
		}

		// Skip if path is in exclude list
		requestPath := ctx.Request.URL.Path
		for _, prefix := range config.ExcludePaths {
			if prefix != "" && strings.HasPrefix(requestPath, prefix) {
				return next()
			}
		}

		// Extract or generate request ID
		requestID := ctx.Header("Traceparent")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store request ID in context
		reqCtx := context.WithValue(ctx.Request.Context(), config.RequestIDKey, requestID)

		// Extract trace context from headers
		carrier := propagation.HeaderCarrier(ctx.Request.Header)
		reqCtx = config.TextMapPropagator.Extract(reqCtx, carrier)

		// Inject trace context into response headers
		responseHeaders := propagation.HeaderCarrier(ctx.Response.Header())
		config.TextMapPropagator.Inject(reqCtx, responseHeaders)

		// Prepare span options
		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(httpconv.ServerRequest(config.ServiceName, ctx.Request)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		// Determine span name
		spanName := ctx.Request.URL.Path
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", ctx.Request.Method)
		}

		// Handle custom trace ID from X-Trace-ID header
		var span oteltrace.Span
		ctxTraceID := ctx.Header("X-Trace-ID")

		if ctxTraceID != "" {
			// Parse and use custom trace ID
			tid, err := oteltrace.TraceIDFromHex(ctxTraceID)
			if err == nil {
				newSpanContext := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
					TraceID:    tid,
					Remote:     true,
					TraceFlags: oteltrace.FlagsSampled,
				})
				reqCtx = oteltrace.ContextWithRemoteSpanContext(reqCtx, newSpanContext)
				reqCtx, span = tracer.Start(reqCtx, spanName, opts...)
			} else {
				// Fall back to normal span creation if trace ID is invalid
				reqCtx, span = tracer.Start(reqCtx, spanName, opts...)
			}
		} else {
			// Start a new span without custom trace ID
			reqCtx, span = tracer.Start(reqCtx, spanName, opts...)
		}
		defer span.End()

		// Get trace ID and set headers
		spanContext := span.SpanContext()
		traceID := spanContext.TraceID().String()

		ctx.SetHeader("X-Trace-ID", traceID)
		ctx.Request.Header.Set("X-Trace-ID", traceID)

		// Update request context
		ctx.Request = ctx.Request.WithContext(reqCtx)

		// Execute next middleware/handler
		err := next()

		// Record response status in span
		statusCode := ctx.StatusCode()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		attrs := semconv.HTTPStatusCode(statusCode)
		span.SetAttributes(attrs)

		// Mark span as error if status code >= 400
		if statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d: %s", statusCode, http.StatusText(statusCode)))
		}

		// Record error in span if present
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}
