package echoadapter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MgApplication/api-server/route"
	"MgApplication/api-server/router-adapter"

	"github.com/labstack/echo/v4"
)

func TestNewEchoAdapter(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type: routeradapter.RouterTypeEcho,
		Port: 8080,
		Echo: &routeradapter.EchoConfig{
			Debug: false,
		},
	}

	adapter, err := NewEchoAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create Echo adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Adapter is nil")
	}

	// Verify engine type
	engine := adapter.Engine()
	if _, ok := engine.(*echo.Echo); !ok {
		t.Fatalf("Engine is not *echo.Echo, got %T", engine)
	}
}

func TestEchoAdapter_RegisterRoute(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type: routeradapter.RouterTypeEcho,
		Port: 8080,
		Echo: &routeradapter.EchoConfig{
			HideBanner: true,
		},
	}

	adapter, err := NewEchoAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Create a simple handler
	handler := func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "hello"})
	}

	// Register route
	meta := route.Meta{
		Method: "GET",
		Path:   "/test",
		Func:   handler,
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	adapter.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestEchoAdapter_RegisterGroup(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type: routeradapter.RouterTypeEcho,
		Port: 8080,
		Echo: &routeradapter.EchoConfig{
			HideBanner: true,
		},
	}

	adapter, err := NewEchoAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Create group
	group := adapter.RegisterGroup("/api/v1", nil)

	// Register route in group
	handler := func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "group route"})
	}

	meta := route.Meta{
		Method: "GET",
		Path:   "/users",
		Func:   handler,
	}

	if err := group.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route in group: %v", err)
	}

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)

	adapter.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestEchoAdapter_Middleware(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type: routeradapter.RouterTypeEcho,
		Port: 8080,
		Echo: &routeradapter.EchoConfig{
			HideBanner: true,
		},
	}

	adapter, err := NewEchoAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Create middleware that adds a header
	middleware := func(ctx *routeradapter.RouterContext, next func() error) error {
		ctx.SetHeader("X-Test", "middleware-works")
		return next()
	}

	// Register middleware
	if err := adapter.RegisterMiddleware(middleware); err != nil {
		t.Fatalf("Failed to register middleware: %v", err)
	}

	// Register route
	handler := func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "test"})
	}

	meta := route.Meta{
		Method: "GET",
		Path:   "/test",
		Func:   handler,
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	adapter.ServeHTTP(w, req)

	// Check if middleware header was added
	if header := w.Header().Get("X-Test"); header != "middleware-works" {
		t.Errorf("Expected header 'X-Test: middleware-works', got '%s'", header)
	}
}

func TestEchoAdapter_NativeMiddleware(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type: routeradapter.RouterTypeEcho,
		Port: 8080,
		Echo: &routeradapter.EchoConfig{
			HideBanner: true,
		},
	}

	adapter, err := NewEchoAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Create Echo middleware
	echoMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Echo-Middleware", "native")
			return next(c)
		}
	}

	// Register native middleware
	if err := adapter.UseNative(echoMiddleware); err != nil {
		t.Fatalf("Failed to register native middleware: %v", err)
	}

	// Register route
	handler := func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "test"})
	}

	meta := route.Meta{
		Method: "GET",
		Path:   "/test",
		Func:   handler,
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	adapter.ServeHTTP(w, req)

	// Check if middleware header was added
	if header := w.Header().Get("X-Echo-Middleware"); header != "native" {
		t.Errorf("Expected header 'X-Echo-Middleware: native', got '%s'", header)
	}
}

func TestEchoAdapter_StartShutdown(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type:         routeradapter.RouterTypeEcho,
		Port:         38888, // Use non-standard port
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Echo: &routeradapter.EchoConfig{
			HideBanner: true,
			HidePort:   true,
		},
	}

	adapter, err := NewEchoAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Register a test route
	handler := func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "test"})
	}

	meta := route.Meta{
		Method: "GET",
		Path:   "/health",
		Func:   handler,
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Start server
	if err := adapter.Start(":38888"); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is running
	resp, err := http.Get("http://localhost:38888/health")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("Response body is empty")
	}

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := adapter.Shutdown(ctx); err != nil {
		t.Fatalf("Failed to shutdown server: %v", err)
	}

	// Verify server is shutdown
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get("http://localhost:38888/health")
	if err == nil {
		t.Error("Server should be shutdown but is still responding")
	}
}

func TestEchoAdapter_ErrorHandling(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type: routeradapter.RouterTypeEcho,
		Port: 8080,
		Echo: &routeradapter.EchoConfig{
			HideBanner: true,
		},
	}

	adapter, err := NewEchoAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Set error handler
	adapter.SetErrorHandler(routeradapter.NewEchoErrorHandler())

	// Register route that returns error
	handler := func(c echo.Context) error {
		return echo.NewHTTPError(500, "test error")
	}

	meta := route.Meta{
		Method: "GET",
		Path:   "/error",
		Func:   handler,
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)

	adapter.ServeHTTP(w, req)

	// Error handler should have processed the error
	// Status should be 500 (internal server error)
	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestEchoAdapter_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		config *routeradapter.EchoConfig
	}{
		{
			name:   "default config",
			config: &routeradapter.EchoConfig{},
		},
		{
			name: "debug mode",
			config: &routeradapter.EchoConfig{
				Debug: true,
			},
		},
		{
			name: "hide banner",
			config: &routeradapter.EchoConfig{
				HideBanner: true,
			},
		},
		{
			name: "hide port",
			config: &routeradapter.EchoConfig{
				HidePort: true,
			},
		},
		{
			name: "disable HTTP/2",
			config: &routeradapter.EchoConfig{
				DisableHTTP2: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &routeradapter.RouterConfig{
				Type: routeradapter.RouterTypeEcho,
				Port: 8080,
				Echo: tt.config,
			}

			adapter, err := NewEchoAdapter(cfg)
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			if adapter == nil {
				t.Fatal("Adapter is nil")
			}
		})
	}
}
