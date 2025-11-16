package route

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// contextKey for test context values
type testContextKey string

const traceIDKey testContextKey = "trace-id"

// TestFromGinCtxPreservesTraceContext tests that fromGinCtx correctly preserves
// trace context and other values from the request context
func TestFromGinCtxPreservesTraceContext(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	// Create Gin context with trace info in request context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Add trace context to request
	ctx := context.WithValue(req.Context(), traceIDKey, "trace-abc-123")
	ctx = context.WithValue(ctx, testContextKey("request-id"), "req-456")
	req = req.WithContext(ctx)

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Create route.Context and populate from Gin context
	routeCtx := getContext()
	defer putContext(routeCtx)

	routeCtx.fromGinCtx(c)

	// Verify trace context is preserved
	if traceID, ok := routeCtx.Value(traceIDKey).(string); !ok || traceID != "trace-abc-123" {
		t.Errorf("Trace ID not preserved! Expected 'trace-abc-123', got: %v", traceID)
		t.Error("❌ Context propagation BROKEN in fromGinCtx!")
	} else {
		t.Log("✅ Trace ID preserved in route.Context")
	}

	// Verify request ID is preserved
	if reqID, ok := routeCtx.Value(testContextKey("request-id")).(string); !ok || reqID != "req-456" {
		t.Errorf("Request ID not preserved! Expected 'req-456', got: %v", reqID)
	} else {
		t.Log("✅ Request ID preserved in route.Context")
	}

	// Verify context is cancellable (has cancel func)
	if routeCtx.cancel == nil {
		t.Error("❌ Cancel function not set!")
	} else {
		t.Log("✅ Context is cancellable")
	}

	t.Log("✅ fromGinCtx() correctly preserves all context values from request")
}

// TestFromGinCtxWithCancel verifies that WithCancel preserves parent context values
func TestFromGinCtxWithCancel(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Add multiple values to request context
	ctx := context.WithValue(req.Context(), testContextKey("key1"), "value1")
	ctx = context.WithValue(ctx, testContextKey("key2"), "value2")
	ctx = context.WithValue(ctx, testContextKey("key3"), "value3")
	req = req.WithContext(ctx)

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	routeCtx := getContext()
	defer putContext(routeCtx)

	routeCtx.fromGinCtx(c)

	// Verify ALL values are accessible
	values := []struct {
		key   testContextKey
		value string
	}{
		{testContextKey("key1"), "value1"},
		{testContextKey("key2"), "value2"},
		{testContextKey("key3"), "value3"},
	}

	for _, v := range values {
		if val, ok := routeCtx.Value(v.key).(string); !ok || val != v.value {
			t.Errorf("Value for %s not preserved! Expected %s, got: %v", v.key, v.value, val)
		}
	}

	t.Log("✅ WithCancel preserves all parent context values")
}

// TestContextDelegation verifies that Context methods delegate to c.Ctx
func TestContextDelegation(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	ctx := context.WithValue(req.Context(), testContextKey("test-key"), "test-value")
	req = req.WithContext(ctx)

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	routeCtx := getContext()
	defer putContext(routeCtx)

	routeCtx.fromGinCtx(c)

	// Test Value() delegation
	if val, ok := routeCtx.Value(testContextKey("test-key")).(string); !ok || val != "test-value" {
		t.Error("Value() method doesn't delegate to c.Ctx correctly")
	}

	// Test Done() delegation
	if routeCtx.Done() == nil {
		t.Error("Done() method doesn't return channel from c.Ctx")
	}

	// Test Err() delegation (should be nil for non-cancelled context)
	if routeCtx.Err() != nil {
		t.Errorf("Err() should be nil for non-cancelled context, got: %v", routeCtx.Err())
	}

	t.Log("✅ Context methods correctly delegate to c.Ctx")
}

// TestContextCancellation verifies cancel function works
func TestContextCancellation(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	routeCtx := getContext()
	// Don't defer putContext here, we'll call cancel manually

	routeCtx.fromGinCtx(c)

	// Cancel the context
	if routeCtx.cancel != nil {
		routeCtx.cancel()
	}

	// Verify context is cancelled
	if routeCtx.Err() == nil {
		t.Error("Context should be cancelled but Err() is nil")
	}

	// Verify Done channel is closed
	select {
	case <-routeCtx.Done():
		t.Log("✅ Context cancellation works correctly")
	default:
		t.Error("Done channel should be closed after cancel()")
	}

	putContext(routeCtx)
}
