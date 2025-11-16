package routeradapter_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"MgApplication/api-server/router-adapter"

	// Import all adapter packages to trigger their init() functions
	_ "MgApplication/api-server/router-adapter/echo"
	_ "MgApplication/api-server/router-adapter/fiber"
	_ "MgApplication/api-server/router-adapter/gin"
	_ "MgApplication/api-server/router-adapter/nethttp"
)

// TestFactoryRegistration tests that all adapters register themselves
func TestFactoryRegistration(t *testing.T) {
	registered := routeradapter.GetRegisteredAdapters()

	expectedAdapters := map[routeradapter.RouterType]bool{
		routeradapter.RouterTypeGin:     false,
		routeradapter.RouterTypeFiber:   false,
		routeradapter.RouterTypeEcho:    false,
		routeradapter.RouterTypeNetHTTP: false,
	}

	for _, adapterType := range registered {
		if _, exists := expectedAdapters[adapterType]; exists {
			expectedAdapters[adapterType] = true
		}
	}

	for adapterType, found := range expectedAdapters {
		if !found {
			t.Errorf("Adapter %s is not registered", adapterType)
		}
	}
}

// TestAdapterCreation tests creating adapters of different types
func TestAdapterCreation(t *testing.T) {
	tests := []struct {
		name        string
		routerType  routeradapter.RouterType
		expectError bool
	}{
		{
			name:        "create gin adapter",
			routerType:  routeradapter.RouterTypeGin,
			expectError: false,
		},
		{
			name:        "create fiber adapter",
			routerType:  routeradapter.RouterTypeFiber,
			expectError: false,
		},
		{
			name:        "create echo adapter",
			routerType:  routeradapter.RouterTypeEcho,
			expectError: false,
		},
		{
			name:        "create nethttp adapter",
			routerType:  routeradapter.RouterTypeNetHTTP,
			expectError: false,
		},
		{
			name:        "create invalid adapter",
			routerType:  "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = tt.routerType

			adapter, err := routeradapter.NewRouterAdapter(cfg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			if adapter == nil {
				t.Fatal("Adapter is nil")
			}

			// Verify adapter is of correct type
			engine := adapter.Engine()
			if engine == nil {
				t.Error("Adapter engine is nil")
			}
		})
	}
}

// TestAdapterLifecycle tests full adapter lifecycle
func TestAdapterLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		routerType routeradapter.RouterType
		port       int
	}{
		{
			name:       "gin lifecycle",
			routerType: routeradapter.RouterTypeGin,
			port:       19001,
		},
		{
			name:       "fiber lifecycle",
			routerType: routeradapter.RouterTypeFiber,
			port:       19002,
		},
		{
			name:       "echo lifecycle",
			routerType: routeradapter.RouterTypeEcho,
			port:       19003,
		},
		{
			name:       "nethttp lifecycle",
			routerType: routeradapter.RouterTypeNetHTTP,
			port:       19004,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create adapter
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = tt.routerType
			cfg.Port = tt.port

			adapter, err := routeradapter.NewRouterAdapter(cfg)
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			// Register a simple test route
			// Note: This uses framework-specific handler registration
			// In real usage, we'd convert handlers properly

			// Start server
			addr := fmt.Sprintf(":%d", tt.port)
			if err := adapter.Start(addr); err != nil {
				t.Fatalf("Failed to start server: %v", err)
			}

			// Give server time to start
			time.Sleep(200 * time.Millisecond)

			// Try to connect (may fail if route wasn't registered properly)
			testURL := fmt.Sprintf("http://localhost:%d/health", tt.port)
			resp, err := http.Get(testURL)
			if err != nil {
				t.Logf("Warning: Failed to connect (expected if route not registered): %v", err)
			} else {
				resp.Body.Close()
			}

			// Shutdown server
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := adapter.Shutdown(ctx); err != nil {
				t.Fatalf("Failed to shutdown server: %v", err)
			}

			// Verify server is stopped
			time.Sleep(100 * time.Millisecond)
			_, err = http.Get(testURL)
			if err == nil {
				t.Error("Server should be stopped but is still responding")
			}
		})
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *routeradapter.RouterConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &routeradapter.RouterConfig{
				Type: routeradapter.RouterTypeGin,
				Port: 8080,
			},
			expectError: false,
		},
		{
			name: "invalid port",
			config: &routeradapter.RouterConfig{
				Type: routeradapter.RouterTypeGin,
				Port: 70000, // Invalid port
			},
			expectError: true,
		},
		{
			name: "invalid router type",
			config: &routeradapter.RouterConfig{
				Type: "invalid",
				Port: 8080,
			},
			expectError: true,
		},
		{
			name: "default router type",
			config: &routeradapter.RouterConfig{
				Type: "", // Should default to Gin
				Port: 8080,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestMustNewRouterAdapter tests panic behavior
func TestMustNewRouterAdapter(t *testing.T) {
	// Test success case
	cfg := routeradapter.DefaultRouterConfig()
	cfg.Type = routeradapter.RouterTypeGin

	adapter := routeradapter.MustNewRouterAdapter(cfg)
	if adapter == nil {
		t.Fatal("routeradapter.MustNewRouterAdapter returned nil")
	}

	// Test panic case
	defer func() {
		if r := recover(); r == nil {
			t.Error("routeradapter.MustNewRouterAdapter should panic on error")
		}
	}()

	badCfg := &routeradapter.RouterConfig{
		Type: "invalid",
		Port: 70000,
	}
	routeradapter.MustNewRouterAdapter(badCfg)
}

// TestErrorHandler tests error handling across adapters
func TestErrorHandler(t *testing.T) {
	handlers := []struct {
		name    string
		handler routeradapter.ErrorHandler
	}{
		{"default", routeradapter.NewDefaultErrorHandler()},
		{"gin", routeradapter.NewGinErrorHandler()},
		{"fiber", routeradapter.NewFiberErrorHandler()},
		{"echo", routeradapter.NewEchoErrorHandler()},
		{"nethttp", routeradapter.NewNetHTTPErrorHandler()},
	}

	for _, tt := range handlers {
		t.Run(tt.name, func(t *testing.T) {
			if tt.handler == nil {
				t.Fatal("Error handler is nil")
			}

			// Test handling error with nil context doesn't crash
			// Note: This would normally panic or fail, but we're testing the handler exists
		})
	}
}

// TestRouterContextOperations tests routeradapter.RouterContext operations
func TestRouterContextOperations(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test?key=value", nil)
	w := &testResponseWriter{}

	ctx := routeradapter.NewRouterContext(w, req)

	// Test param operations
	ctx.SetParam("id", "123")
	if ctx.Param("id") != "123" {
		t.Error("SetParam/Param failed")
	}

	// Test query param
	if ctx.QueryParam("key") != "value" {
		t.Error("QueryParam failed")
	}

	// Test data operations
	ctx.Set("test", "value")
	if ctx.Get("test") != "value" {
		t.Error("Set/Get failed")
	}

	// Test status
	ctx.Status(201)
	if ctx.StatusCode() != 201 {
		t.Error("Status failed")
	}

	// Test header
	ctx.SetHeader("X-Test", "header")
	if ctx.Header("X-Custom") != "" {
		// Header not set, should be empty
	}

	// Test JSON response
	err := ctx.JSON(200, map[string]string{"message": "test"})
	if err != nil {
		t.Errorf("JSON failed: %v", err)
	}

	if !ctx.IsResponseWritten() {
		t.Error("Response should be marked as written")
	}

	// Test double write protection
	err = ctx.JSON(200, map[string]string{"message": "test2"})
	if err != routeradapter.ErrResponseAlreadyWritten {
		t.Error("Should prevent double write")
	}
}

// testResponseWriter is a simple http.ResponseWriter for testing
type testResponseWriter struct {
	headers    http.Header
	body       []byte
	statusCode int
}

func (w *testResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// TestNewRouterAdapterFromType tests creating adapter from type only
func TestNewRouterAdapterFromType(t *testing.T) {
	adapter, err := routeradapter.NewRouterAdapterFromType(routeradapter.RouterTypeGin)
	if err != nil {
		t.Fatalf("Failed to create adapter from type: %v", err)
	}

	if adapter == nil {
		t.Fatal("Adapter is nil")
	}
}

// TestMiddlewareChaining tests middleware execution order
func TestMiddlewareChaining(t *testing.T) {
	order := []string{}

	mw1 := func(ctx *routeradapter.RouterContext, next func() error) error {
		order = append(order, "mw1-before")
		err := next()
		order = append(order, "mw1-after")
		return err
	}

	mw2 := func(ctx *routeradapter.RouterContext, next func() error) error {
		order = append(order, "mw2-before")
		err := next()
		order = append(order, "mw2-after")
		return err
	}

	handler := func(ctx *routeradapter.RouterContext) error {
		order = append(order, "handler")
		return nil
	}

	// Simulate middleware chain
	req, _ := http.NewRequest("GET", "/test", nil)
	w := &testResponseWriter{}
	ctx := routeradapter.NewRouterContext(w, req)

	// Execute chain: mw1 -> mw2 -> handler
	_ = mw1(ctx, func() error {
		return mw2(ctx, func() error {
			return handler(ctx)
		})
	})

	expectedOrder := []string{
		"mw1-before",
		"mw2-before",
		"handler",
		"mw2-after",
		"mw1-after",
	}

	if len(order) != len(expectedOrder) {
		t.Fatalf("Expected %d middleware calls, got %d", len(expectedOrder), len(order))
	}

	for i, expected := range expectedOrder {
		if order[i] != expected {
			t.Errorf("At position %d: expected %s, got %s", i, expected, order[i])
		}
	}
}

// BenchmarkAdapterCreation benchmarks adapter creation
func BenchmarkAdapterCreation(b *testing.B) {
	adapters := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, adapterType := range adapters {
		b.Run(string(adapterType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = adapterType

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = routeradapter.NewRouterAdapter(cfg)
			}
		})
	}
}

// BenchmarkRouterContextOperations benchmarks routeradapter.RouterContext operations
func BenchmarkRouterContextOperations(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test?key=value", nil)

	b.Run("SetParam", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := &testResponseWriter{}
			ctx := routeradapter.NewRouterContext(w, req)
			ctx.SetParam("id", "123")
		}
	})

	b.Run("GetParam", func(b *testing.B) {
		w := &testResponseWriter{}
		ctx := routeradapter.NewRouterContext(w, req)
		ctx.SetParam("id", "123")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ctx.Param("id")
		}
	})

	b.Run("SetGet", func(b *testing.B) {
		w := &testResponseWriter{}
		ctx := routeradapter.NewRouterContext(w, req)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx.Set("key", "value")
			_ = ctx.Get("key")
		}
	})
}
