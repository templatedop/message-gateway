package routeradapter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthCheckCreation tests creating a health check manager
func TestHealthCheckCreation(t *testing.T) {
	hc := NewHealthCheck()
	if hc == nil {
		t.Fatal("NewHealthCheck returned nil")
	}

	// Should start in healthy state
	if hc.IsShuttingDown() {
		t.Error("New health check should not be shutting down")
	}
}

// TestHealthCheckShutdown tests marking service as shutting down
func TestHealthCheckShutdown(t *testing.T) {
	hc := NewHealthCheck()

	// Mark as shutting down
	hc.MarkShuttingDown()

	if !hc.IsShuttingDown() {
		t.Error("Health check should be shutting down after MarkShuttingDown()")
	}
}

// TestHealthzHandlerHealthy tests health check endpoint when healthy
func TestHealthzHandlerHealthy(t *testing.T) {
	hc := NewHealthCheck()
	middleware := HealthzHandler(hc)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	ctx := NewRouterContext(w, req)

	// Should handle /healthz path
	nextCalled := false
	err := middleware(ctx, func() error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if nextCalled {
		t.Error("Next handler should not be called for /healthz path")
	}

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !contains(body, "ok") && !contains(body, "\"status\":\"ok\"") {
		t.Errorf("Expected 'ok' in response body, got: %s", body)
	}

	t.Logf("Health check response: %s", body)
}

// TestHealthzHandlerShuttingDown tests health check when shutting down
func TestHealthzHandlerShuttingDown(t *testing.T) {
	hc := NewHealthCheck()
	hc.MarkShuttingDown()

	middleware := HealthzHandler(hc)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	ctx := NewRouterContext(w, req)

	err := middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return 503 when shutting down
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	body := w.Body.String()
	if !contains(body, "shutting down") {
		t.Errorf("Expected 'shutting down' in response body, got: %s", body)
	}

	t.Logf("Shutdown response: %s", body)
}

// TestHealthzHandlerNonHealthzPath tests that middleware passes through non-healthz requests
func TestHealthzHandlerNonHealthzPath(t *testing.T) {
	hc := NewHealthCheck()
	middleware := HealthzHandler(hc)

	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	ctx := NewRouterContext(w, req)

	nextCalled := false
	err := middleware(ctx, func() error {
		nextCalled = true
		return ctx.JSON(200, map[string]string{"status": "users"})
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !nextCalled {
		t.Error("Next handler should be called for non-healthz paths")
	}
}

// TestHealthzHandlerWrongMethod tests that only GET requests are handled
func TestHealthzHandlerWrongMethod(t *testing.T) {
	hc := NewHealthCheck()
	middleware := HealthzHandler(hc)

	req := httptest.NewRequest("POST", "/healthz", nil)
	w := httptest.NewRecorder()
	ctx := NewRouterContext(w, req)

	nextCalled := false
	err := middleware(ctx, func() error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !nextCalled {
		t.Error("Next handler should be called for non-GET methods")
	}
}

// TestHealthzHandlerNilHealthCheck tests that handler works with nil health check
func TestHealthzHandlerNilHealthCheck(t *testing.T) {
	middleware := HealthzHandler(nil)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	ctx := NewRouterContext(w, req)

	err := middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return 200 OK with default health check
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
