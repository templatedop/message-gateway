package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MgApplication/api-server/router-adapter"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TestBodyLimiter tests the body size limiting middleware
func TestBodyLimiter(t *testing.T) {
	tests := []struct {
		name           string
		limit          int64
		bodySize       int
		expectStatus   int
		expectRejected bool
	}{
		{
			name:           "body within limit",
			limit:          1024,
			bodySize:       512,
			expectStatus:   200,
			expectRejected: false,
		},
		{
			name:           "body exceeds limit",
			limit:          100,
			bodySize:       200,
			expectStatus:   413,
			expectRejected: true,
		},
		{
			name:           "body exactly at limit",
			limit:          100,
			bodySize:       100,
			expectStatus:   200,
			expectRejected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := BodyLimiter(tt.limit)

			// Create test request with body
			body := bytes.Repeat([]byte("a"), tt.bodySize)
			req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Create RouterContext
			ctx := routeradapter.NewRouterContext(w, req)

			// Track if next was called
			nextCalled := false
			next := func() error {
				nextCalled = true
				return ctx.JSON(200, map[string]string{"status": "ok"})
			}

			// Execute middleware
			err := middleware(ctx, next)

			// Verify results
			if tt.expectRejected {
				if err == nil {
					t.Error("Expected error for oversized body")
				}
				if nextCalled {
					t.Error("Next handler should not be called for oversized body")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !nextCalled {
					t.Error("Next handler should be called for acceptable body size")
				}
			}
		})
	}
}

// TestRecovery tests panic recovery middleware
func TestRecovery(t *testing.T) {
	tests := []struct {
		name        string
		handler     func(*routeradapter.RouterContext) error
		expectPanic bool
	}{
		{
			name: "no panic",
			handler: func(ctx *routeradapter.RouterContext) error {
				return ctx.JSON(200, map[string]string{"status": "ok"})
			},
			expectPanic: false,
		},
		{
			name: "panic with string",
			handler: func(ctx *routeradapter.RouterContext) error {
				panic("test panic")
			},
			expectPanic: true,
		},
		{
			name: "panic with error",
			handler: func(ctx *routeradapter.RouterContext) error {
				panic(fmt.Errorf("test error"))
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := Recovery()

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			ctx := routeradapter.NewRouterContext(w, req)

			// Execute middleware
			err := middleware(ctx, func() error {
				return tt.handler(ctx)
			})

			// Recovery middleware should never return error, it handles panics
			if err != nil {
				t.Errorf("Recovery middleware should not return error, got: %v", err)
			}

			// Check status code
			if tt.expectPanic && w.Code != 500 {
				t.Errorf("Expected status 500 after panic, got %d", w.Code)
			}
		})
	}
}

// TestCORS tests CORS middleware
func TestCORS(t *testing.T) {
	tests := []struct {
		name          string
		config        CORSConfig
		origin        string
		method        string
		expectAllowed bool
		expectHeaders map[string]string
	}{
		{
			name: "allowed origin",
			config: CORSConfig{
				AllowOrigins: []string{"http://example.com"},
				AllowMethods: []string{"GET", "POST"},
			},
			origin:        "http://example.com",
			method:        "GET",
			expectAllowed: true,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin": "http://example.com",
			},
		},
		{
			name: "wildcard origin",
			config: CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET"},
			},
			origin:        "http://anywhere.com",
			method:        "GET",
			expectAllowed: true,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name: "preflight request",
			config: CORSConfig{
				AllowOrigins: []string{"http://example.com"},
				AllowMethods: []string{"POST"},
			},
			origin:        "http://example.com",
			method:        "OPTIONS",
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := CORS(tt.config)

			// Create test request
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			if tt.method == "OPTIONS" {
				req.Header.Set("Access-Control-Request-Method", "POST")
			}

			w := httptest.NewRecorder()
			ctx := routeradapter.NewRouterContext(w, req)

			// Execute middleware
			nextCalled := false
			err := middleware(ctx, func() error {
				nextCalled = true
				return nil
			})

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify CORS headers
			for header, expectedValue := range tt.expectHeaders {
				actualValue := w.Header().Get(header)
				if actualValue != expectedValue {
					t.Errorf("Expected %s header to be %q, got %q", header, expectedValue, actualValue)
				}
			}

			// OPTIONS requests should not call next handler
			if tt.method == "OPTIONS" && nextCalled {
				t.Error("OPTIONS request should not call next handler")
			}
		})
	}
}

// TestRateLimiter tests rate limiting middleware
func TestRateLimiter(t *testing.T) {
	// This test would require the LeakyBucket implementation
	// Skipping for now as it depends on external package
	t.Skip("Rate limiter requires LeakyBucket implementation")
}

// TestTimeout tests request timeout middleware
func TestTimeout(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		handlerDelay  time.Duration
		expectTimeout bool
	}{
		{
			name:          "completes within timeout",
			timeout:       100 * time.Millisecond,
			handlerDelay:  10 * time.Millisecond,
			expectTimeout: false,
		},
		{
			name:          "exceeds timeout",
			timeout:       50 * time.Millisecond,
			handlerDelay:  200 * time.Millisecond,
			expectTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := Timeout(tt.timeout)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			ctx := routeradapter.NewRouterContext(w, req)

			// Execute middleware
			err := middleware(ctx, func() error {
				time.Sleep(tt.handlerDelay)
				return ctx.JSON(200, map[string]string{"status": "ok"})
			})

			if tt.expectTimeout {
				if err == nil {
					t.Error("Expected timeout error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestTracing tests OpenTelemetry tracing middleware
func TestTracing(t *testing.T) {
	// Create a test tracer provider
	tp := sdktrace.NewTracerProvider()
	defer tp.Shutdown(context.Background())

	otel.SetTracerProvider(tp)

	// Create tracing config
	config := TracingConfig{
		TracerProvider:    tp,
		TextMapPropagator: propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
		ServiceName:       "test-service",
	}

	// Create middleware
	middleware := Tracing(config)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	// Execute middleware
	err := middleware(ctx, func() error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify trace ID header was added
	traceID := w.Header().Get("X-Trace-ID")
	if traceID == "" {
		t.Error("Expected X-Trace-ID header to be set")
	}

	t.Logf("Trace ID: %s", traceID)
}

// TestMetrics tests Prometheus metrics middleware
func TestMetrics(t *testing.T) {
	// Create a test registry
	registry := prometheus.NewRegistry()

	// Create metrics config
	config := MetricsConfig{
		Registry:  registry,
		Namespace: "test",
		Subsystem: "service",
	}

	// Create middleware
	middleware := Metrics(config)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	// Execute middleware
	err := middleware(ctx, func() error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify metrics were recorded
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Check for HTTP request metrics
	found := false
	for _, mf := range metrics {
		if strings.Contains(mf.GetName(), "http_server_requests") {
			found = true
			t.Logf("Found metric: %s", mf.GetName())
		}
	}

	if !found {
		t.Error("Expected HTTP request metrics to be recorded")
	}
}

// TestRequestResponseLogger tests logging middleware
func TestRequestResponseLogger(t *testing.T) {
	// Create middleware
	middleware := RequestResponseLogger()

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	// Execute middleware
	err := middleware(ctx, func() error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Log("Logging middleware executed successfully")
}

// TestSignatureVerification tests request signature verification
func TestSignatureVerification(t *testing.T) {
	// Test public/private keys (base64 encoded)
	// These are test keys only
	pubKey := "N3Iu59pOFhQ4EcH0jRx3vkhOxUHb6Gh5USHyApvWccM="
	privKey := "/wnL6WJmFKG4X14zzvNrZog8+dHtaNoD30rEdwxIbf3p+U4BH83F5SrRvIC8M/0Qi6zorAydLGk+/bE8KsBtFA=="

	t.Run("valid signature", func(t *testing.T) {
		// Create test JSON body
		testBody := map[string]string{"test": "data"}
		bodyBytes, _ := json.Marshal(testBody)

		// Sign the body
		sigHeader, err := signJSON(bodyBytes, privKey, 3600)
		if err != nil {
			t.Fatalf("Failed to sign JSON: %v", err)
		}

		// Create request signature verifier
		middleware := RequestSignatureVerifier(SignatureConfig{
			PublicKey:       pubKey,
			ValiditySeconds: 3600,
			SignatureHeader: "sig",
			SkipMethods:     []string{"GET"},
		})

		// Create test request
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
		req.Header.Set("sig", sigHeader)
		w := httptest.NewRecorder()
		ctx := routeradapter.NewRouterContext(w, req)

		// Execute middleware
		nextCalled := false
		err = middleware(ctx, func() error {
			nextCalled = true
			return ctx.JSON(200, map[string]string{"status": "ok"})
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !nextCalled {
			t.Error("Next handler should be called for valid signature")
		}

		t.Log("Signature verification passed ✓")
	})

	t.Run("missing signature header", func(t *testing.T) {
		middleware := RequestSignatureVerifier(SignatureConfig{
			PublicKey: pubKey,
		})

		req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
		w := httptest.NewRecorder()
		ctx := routeradapter.NewRouterContext(w, req)

		err := middleware(ctx, func() error {
			return nil
		})

		if err == nil {
			t.Error("Expected error for missing signature")
		}
	})
}

// TestResponseSigner tests response signing middleware
func TestResponseSigner(t *testing.T) {
	privKey := "/wnL6WJmFKG4X14zzvNrZog8+dHtaNoD30rEdwxIbf3p+U4BH83F5SrRvIC8M/0Qi6zorAydLGk+/bE8KsBtFA=="

	middleware := ResponseSigner(SignatureConfig{
		PrivateKey:      privKey,
		ValiditySeconds: 3600,
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	// Execute middleware
	err := middleware(ctx, func() error {
		ctx.Response.Header().Set("Content-Type", "application/json")
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify signature header was added
	sigHeader := w.Header().Get("sig")
	if sigHeader == "" {
		t.Error("Expected sig header to be set")
	}

	t.Logf("Response signature added ✓")
}

// TestConnectionLimiter tests connection limiting middleware
func TestConnectionLimiter(t *testing.T) {
	config := ConnectionLimiterConfig{
		MaxConnections:   2,
		RejectStatusCode: 503,
	}

	middleware := ConnectionLimiter(config)

	// Test within limit
	t.Run("within limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		ctx := routeradapter.NewRouterContext(w, req)

		err := middleware(ctx, func() error {
			return ctx.JSON(200, map[string]string{"status": "ok"})
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Test concurrent connections
	t.Run("concurrent connections", func(t *testing.T) {
		// Create a blocker to hold connections
		blocker := make(chan struct{})
		defer close(blocker)

		// Start 2 concurrent requests (at max limit)
		for i := 0; i < 2; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				ctx := routeradapter.NewRouterContext(w, req)

				middleware(ctx, func() error {
					<-blocker // Wait
					return ctx.JSON(200, map[string]string{"status": "ok"})
				})
			}()
		}

		// Give goroutines time to start
		time.Sleep(50 * time.Millisecond)

		// Try a 3rd connection (should be rejected)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		ctx := routeradapter.NewRouterContext(w, req)

		err := middleware(ctx, func() error {
			return ctx.JSON(200, map[string]string{"status": "ok"})
		})

		if err == nil {
			t.Error("Expected error for connection limit exceeded")
		}
	})
}

// Helper function to read response body
func readBody(r io.Reader) string {
	body, _ := io.ReadAll(r)
	return string(body)
}
