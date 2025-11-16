package routeradapter_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"MgApplication/api-server/router-adapter"
)

// TestRouterContextDynamicRead tests that RouterContext.Context() reads from Request.Context() dynamically
func TestRouterContextDynamicRead(t *testing.T) {
	// Create a request with initial context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Create RouterContext
	rctx := routeradapter.NewRouterContext(w, req)

	// Verify initial context is empty (no custom values)
	if val := rctx.Context().Value("test-key"); val != nil {
		t.Errorf("Expected context to be empty initially, but found value: %v", val)
	}

	// Update the request context directly
	ctx := context.WithValue(req.Context(), "test-key", "test-value")
	req = req.WithContext(ctx)

	// Update the request in RouterContext
	rctx.Request = req

	// Now verify RouterContext.Context() returns the updated context
	if val, ok := rctx.Context().Value("test-key").(string); !ok || val != "test-value" {
		t.Errorf("Expected RouterContext.Context() to return updated context with value 'test-value', got: %v", val)
	} else {
		t.Log("✓ RouterContext.Context() correctly reads from Request.Context() dynamically")
	}
}

// TestRouterContextSetContext tests that SetContext updates both internal and request context
func TestRouterContextSetContextMethod(t *testing.T) {
	// Create a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Create RouterContext
	rctx := routeradapter.NewRouterContext(w, req)

	// Set a new context using SetContext
	newCtx := context.WithValue(context.Background(), "user-id", "12345")
	rctx.SetContext(newCtx)

	// Verify RouterContext.Context() returns the new context
	if val, ok := rctx.Context().Value("user-id").(string); !ok || val != "12345" {
		t.Errorf("Expected RouterContext.Context() to have user-id=12345, got: %v", val)
	}

	// Verify Request.Context() also has the new context
	if val, ok := rctx.Request.Context().Value("user-id").(string); !ok || val != "12345" {
		t.Errorf("Expected Request.Context() to have user-id=12345, got: %v", val)
	}

	t.Log("✓ SetContext correctly updates both internal and request context")
}
