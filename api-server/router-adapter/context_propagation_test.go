package routeradapter_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"MgApplication/api-server/route"
	"MgApplication/api-server/router-adapter"

	"github.com/gin-gonic/gin"

	// Import all adapters
	_ "MgApplication/api-server/router-adapter/echo"
	_ "MgApplication/api-server/router-adapter/fiber"
	_ "MgApplication/api-server/router-adapter/gin"
	_ "MgApplication/api-server/router-adapter/nethttp"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	traceIDKey     contextKey = "trace-id"
	requestIDKey   contextKey = "request-id"
	userIDKey      contextKey = "user-id"
	middlewareKey  contextKey = "middleware-called"
)

// TestContextPropagation tests that context is properly propagated through middleware chain
// This test only works with Gin adapter due to handler type compatibility
func TestContextPropagation(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	cfg := routeradapter.DefaultRouterConfig()
	cfg.Type = routeradapter.RouterTypeGin

	adapter, err := routeradapter.NewRouterAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Track if handler received updated context
	handlerCalled := false
	contextHasTraceID := false
	contextHasRequestID := false

	// Register middleware that adds values to context
	middleware := func(rctx *routeradapter.RouterContext, next func() error) error {
		// Get current context from request
		ctx := rctx.Request.Context()

		// Add trace ID to context
		ctx = context.WithValue(ctx, traceIDKey, "trace-123")

		// Add request ID to context
		ctx = context.WithValue(ctx, requestIDKey, "req-456")

		// Update request with new context
		rctx.Request = rctx.Request.WithContext(ctx)

		// Call next handler
		return next()
	}

	err = adapter.RegisterMiddleware(middleware)
	if err != nil {
		t.Fatalf("Failed to register middleware: %v", err)
	}

	// Register handler that checks context
	handler := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			handlerCalled = true

			// Get context - should have values from middleware
			ctx := c.Request.Context()

			// Check for trace ID
			if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID == "trace-123" {
				contextHasTraceID = true
			}

			// Check for request ID
			if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID == "req-456" {
				contextHasRequestID = true
			}

			c.JSON(200, gin.H{"status": "ok"})
		}
	}()

	err = adapter.RegisterRoute(route.Meta{
		Method: "GET",
		Path:   "/test",
		Func:   handler,
	})
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	adapter.ServeHTTP(w, req)

	// Verify handler was called
	if !handlerCalled {
		t.Error("Handler was not called")
	}

	// Verify context propagation
	if !contextHasTraceID {
		t.Error("Trace ID not found in context (context propagation failed)")
	}

	if !contextHasRequestID {
		t.Error("Request ID not found in context (context propagation failed)")
	}

	t.Logf("Context propagation ✓ (trace_id=%v, request_id=%v)",
		contextHasTraceID, contextHasRequestID)
}

// TestSetContext tests that SetContext properly updates both internal and request context
func TestSetContext(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	cfg := routeradapter.DefaultRouterConfig()
	cfg.Type = routeradapter.RouterTypeGin

	adapter, err := routeradapter.NewRouterAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	contextUpdated := false

	// Handler that uses SetContext
	handler := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			rctx := routeradapter.NewRouterContext(c.Writer, c.Request)

			// Update context using SetContext
			ctx := context.WithValue(context.Background(), middlewareKey, "set-context-test")
			rctx.SetContext(ctx)

			// Verify both rctx.Context() and rctx.Request.Context() are updated
			if val, ok := rctx.Context().Value(middlewareKey).(string); ok && val == "set-context-test" {
				if val2, ok2 := rctx.Request.Context().Value(middlewareKey).(string); ok2 && val2 == "set-context-test" {
					contextUpdated = true
				}
			}

			c.JSON(200, gin.H{"status": "ok"})
		}
	}()

	err = adapter.RegisterRoute(route.Meta{
		Method: "GET",
		Path:   "/test",
		Func:   handler,
	})
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	adapter.ServeHTTP(w, req)

	if !contextUpdated {
		t.Error("SetContext did not properly update both internal and request context")
	}

	t.Log("SetContext properly updates both contexts ✓")
}

// TestMiddlewareChainContext tests context propagation through multiple middlewares
func TestMiddlewareChainContext(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	cfg := routeradapter.DefaultRouterConfig()
	cfg.Type = routeradapter.RouterTypeGin

	adapter, err := routeradapter.NewRouterAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Track which middlewares were called and what context they saw
	middleware1Called := false
	middleware2Called := false
	middleware3Called := false
	middleware2SawValue1 := false
	middleware3SawValue1 := false
	middleware3SawValue2 := false
	handlerSawAllValues := false

	// Middleware 1: Adds value1
	middleware1 := func(rctx *routeradapter.RouterContext, next func() error) error {
		middleware1Called = true
		ctx := context.WithValue(rctx.Request.Context(), contextKey("value1"), "from-mw1")
		rctx.Request = rctx.Request.WithContext(ctx)
		return next()
	}

	// Middleware 2: Checks value1, adds value2
	middleware2 := func(rctx *routeradapter.RouterContext, next func() error) error {
		middleware2Called = true
		if val, ok := rctx.Request.Context().Value(contextKey("value1")).(string); ok && val == "from-mw1" {
			middleware2SawValue1 = true
		}
		ctx := context.WithValue(rctx.Request.Context(), contextKey("value2"), "from-mw2")
		rctx.Request = rctx.Request.WithContext(ctx)
		return next()
	}

	// Middleware 3: Checks value1 and value2, adds value3
	middleware3 := func(rctx *routeradapter.RouterContext, next func() error) error {
		middleware3Called = true
		if val, ok := rctx.Request.Context().Value(contextKey("value1")).(string); ok && val == "from-mw1" {
			middleware3SawValue1 = true
		}
		if val, ok := rctx.Request.Context().Value(contextKey("value2")).(string); ok && val == "from-mw2" {
			middleware3SawValue2 = true
		}
		ctx := context.WithValue(rctx.Request.Context(), contextKey("value3"), "from-mw3")
		rctx.Request = rctx.Request.WithContext(ctx)
		return next()
	}

	err = adapter.RegisterMiddleware(middleware1)
	if err != nil {
		t.Fatalf("Failed to register middleware1: %v", err)
	}

	err = adapter.RegisterMiddleware(middleware2)
	if err != nil {
		t.Fatalf("Failed to register middleware2: %v", err)
	}

	err = adapter.RegisterMiddleware(middleware3)
	if err != nil {
		t.Fatalf("Failed to register middleware3: %v", err)
	}

	// Handler checks all three values
	handler := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			ctx := c.Request.Context()
			val1, ok1 := ctx.Value(contextKey("value1")).(string)
			val2, ok2 := ctx.Value(contextKey("value2")).(string)
			val3, ok3 := ctx.Value(contextKey("value3")).(string)

			if ok1 && val1 == "from-mw1" && ok2 && val2 == "from-mw2" && ok3 && val3 == "from-mw3" {
				handlerSawAllValues = true
			}

			c.JSON(200, gin.H{"status": "ok"})
		}
	}()

	err = adapter.RegisterRoute(route.Meta{
		Method: "GET",
		Path:   "/test",
		Func:   handler,
	})
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	adapter.ServeHTTP(w, req)

	// Verify middleware chain
	if !middleware1Called || !middleware2Called || !middleware3Called {
		t.Error("Not all middlewares were called")
	}

	if !middleware2SawValue1 {
		t.Error("Middleware 2 did not see value from Middleware 1 (context not propagated)")
	}

	if !middleware3SawValue1 || !middleware3SawValue2 {
		t.Error("Middleware 3 did not see values from previous middlewares (context not propagated)")
	}

	if !handlerSawAllValues {
		t.Error("Handler did not see all context values from middleware chain")
	}

	t.Log("Context properly propagates through middleware chain ✓")
}
