package prometheus

import (
	"net/http/httptest"
	"strings"
	"testing"

	"MgApplication/api-server/router-adapter"

	"github.com/prometheus/client_golang/prometheus"
)

// TestMetricsHandler tests the Prometheus metrics handler
func TestMetricsHandler(t *testing.T) {
	// Create a test registry
	registry := prometheus.NewRegistry()

	// Register some test metrics
	testCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter_total",
		Help: "A test counter",
	})
	registry.MustRegister(testCounter)
	testCounter.Inc()

	// Create middleware
	middleware := MetricsHandler("/metrics", registry)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	nextCalled := false
	err := middleware(ctx, func() error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if nextCalled {
		t.Error("Next handler should not be called for /metrics path")
	}

	// Check response
	body := w.Body.String()

	// Should contain Prometheus metrics format
	if !strings.Contains(body, "test_counter_total") {
		t.Errorf("Expected test_counter_total in metrics output, got:\n%s", body)
	}

	// Should contain the TYPE and HELP comments
	if !strings.Contains(body, "# HELP") || !strings.Contains(body, "# TYPE") {
		t.Error("Expected Prometheus metrics format with HELP and TYPE comments")
	}

	t.Logf("Metrics endpoint served successfully:\n%s", body)
}

// TestDefaultMetricsHandler tests the default metrics handler
func TestDefaultMetricsHandler(t *testing.T) {
	middleware := DefaultMetricsHandler()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	err := middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should serve metrics from default registry
	body := w.Body.String()
	if !strings.Contains(body, "# HELP") {
		t.Error("Expected Prometheus metrics format")
	}

	t.Logf("Default metrics handler served successfully")
}

// TestMetricsHandlerNonMatchingPath tests that handler passes through non-matching paths
func TestMetricsHandlerNonMatchingPath(t *testing.T) {
	middleware := MetricsHandler("/metrics")

	req := httptest.NewRequest("GET", "/api/other", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	nextCalled := false
	err := middleware(ctx, func() error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !nextCalled {
		t.Error("Next handler should be called for non-matching path")
	}
}

// TestMetricsHandlerCustomPath tests metrics handler with custom path
func TestMetricsHandlerCustomPath(t *testing.T) {
	middleware := MetricsHandler("/custom/metrics/path")

	req := httptest.NewRequest("GET", "/custom/metrics/path", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	err := middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should serve metrics
	body := w.Body.String()
	if !strings.Contains(body, "# HELP") {
		t.Error("Expected Prometheus metrics format on custom path")
	}

	t.Logf("Custom path metrics served successfully")
}

// TestMetricsHandlerContentType tests that content type is set correctly
func TestMetricsHandlerContentType(t *testing.T) {
	middleware := DefaultMetricsHandler()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	err := middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") && !strings.Contains(contentType, "prometheus") {
		t.Errorf("Expected Prometheus content type, got: %s", contentType)
	}

	t.Logf("Content-Type: %s", contentType)
}
