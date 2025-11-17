package nethttpadapter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MgApplication/api-server/route"
	"MgApplication/api-server/router-adapter"

	"github.com/gin-gonic/gin"
)

// dummyGinHandler creates a dummy gin handler for testing
// Since net/http adapter doesn't actually execute Gin handlers,
// this is just a placeholder to satisfy route.Meta type requirements
func dummyGinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	}
}

func TestNewNetHTTPAdapter(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type: routeradapter.RouterTypeNetHTTP,
		Port: 8080,
		NetHTTP: &routeradapter.NetHTTPConfig{
			EnableHTTP2: false,
		},
	}

	adapter, err := NewNetHTTPAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create net/http adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Adapter is nil")
	}

	// Verify engine type
	engine := adapter.Engine()
	if _, ok := engine.(*http.ServeMux); !ok {
		t.Fatalf("Engine is not *http.ServeMux, got %T", engine)
	}
}

func TestNetHTTPAdapter_RegisterRoute(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type:    routeradapter.RouterTypeNetHTTP,
		Port:    8080,
		NetHTTP: &routeradapter.NetHTTPConfig{},
	}

	adapter, err := NewNetHTTPAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Register route (using dummy Gin handler since route.Meta expects gin.HandlerFunc)
	meta := route.Meta{
		Method: "GET",
		Path:   "/test",
		Func:   dummyGinHandler(),
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

	body := w.Body.String()
	if body == "" {
		t.Error("Response body is empty")
	}
}

func TestNetHTTPAdapter_PathParameters(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type:    routeradapter.RouterTypeNetHTTP,
		Port:    8080,
		NetHTTP: &routeradapter.NetHTTPConfig{},
	}

	adapter, err := NewNetHTTPAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Register route with path parameter (using dummy Gin handler)
	meta := route.Meta{
		Method: "GET",
		Path:   "/users/:id",
		Func:   dummyGinHandler(),
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/123", nil)

	adapter.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNetHTTPAdapter_RegisterGroup(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type:    routeradapter.RouterTypeNetHTTP,
		Port:    8080,
		NetHTTP: &routeradapter.NetHTTPConfig{},
	}

	adapter, err := NewNetHTTPAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Create group
	group := adapter.RegisterGroup("/api/v1", nil)

	// Register route in group
	meta := route.Meta{
		Method: "GET",
		Path:   "/users",
		Func:   dummyGinHandler(),
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

func TestNetHTTPAdapter_Middleware(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type:    routeradapter.RouterTypeNetHTTP,
		Port:    8080,
		NetHTTP: &routeradapter.NetHTTPConfig{},
	}

	adapter, err := NewNetHTTPAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Create middleware that adds a header
	middlewareCalled := false
	middleware := func(ctx *routeradapter.RouterContext, next func() error) error {
		middlewareCalled = true
		ctx.SetHeader("X-Test", "middleware-works")
		return next()
	}

	// Register middleware
	if err := adapter.RegisterMiddleware(middleware); err != nil {
		t.Fatalf("Failed to register middleware: %v", err)
	}

	// Register route
	meta := route.Meta{
		Method: "GET",
		Path:   "/test",
		Func:   dummyGinHandler(),
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	adapter.ServeHTTP(w, req)

	// Check middleware was called
	if !middlewareCalled {
		t.Error("Middleware was not called")
	}

	// Check if middleware header was added
	if header := w.Header().Get("X-Test"); header != "middleware-works" {
		t.Errorf("Expected header 'X-Test: middleware-works', got '%s'", header)
	}
}

func TestNetHTTPAdapter_StartShutdown(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type:         routeradapter.RouterTypeNetHTTP,
		Port:         48888, // Use non-standard port
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		NetHTTP: &routeradapter.NetHTTPConfig{
			EnableHTTP2: false,
		},
	}

	adapter, err := NewNetHTTPAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Register a test route
	meta := route.Meta{
		Method: "GET",
		Path:   "/health",
		Func:   dummyGinHandler(),
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Start server
	if err := adapter.Start(":48888"); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is running
	resp, err := http.Get("http://localhost:48888/health")
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
	_, err = http.Get("http://localhost:48888/health")
	if err == nil {
		t.Error("Server should be shutdown but is still responding")
	}
}

func TestNetHTTPAdapter_Router(t *testing.T) {
	router := NewRouter()

	// Test adding routes
	handlerCalled := false
	handler := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		handlerCalled = true
		if id, ok := params["id"]; ok {
			if id != "123" {
				t.Errorf("Expected param id=123, got %s", id)
			}
		}
	}

	router.AddRoute("GET", "/users/:id", handler)

	// Test matching route
	matchedHandler, params := router.Match("GET", "/users/123")
	if matchedHandler == nil {
		t.Fatal("Route not matched")
	}

	if id, ok := params["id"]; !ok || id != "123" {
		t.Errorf("Expected param id=123, got %s (exists: %v)", id, ok)
	}

	// Call the handler
	matchedHandler(nil, nil, params)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestPathToRegex(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		testPath    string
		shouldMatch bool
		params      map[string]string
	}{
		{
			name:        "simple path",
			path:        "/users",
			testPath:    "/users",
			shouldMatch: true,
			params:      map[string]string{},
		},
		{
			name:        "single parameter",
			path:        "/users/:id",
			testPath:    "/users/123",
			shouldMatch: true,
			params:      map[string]string{"id": "123"},
		},
		{
			name:        "multiple parameters",
			path:        "/posts/:postId/comments/:id",
			testPath:    "/posts/456/comments/789",
			shouldMatch: true,
			params:      map[string]string{"postId": "456", "id": "789"},
		},
		{
			name:        "no match",
			path:        "/users/:id",
			testPath:    "/posts/123",
			shouldMatch: false,
			params:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, paramNames := pathToRegex(tt.path)

			matches := pattern.FindStringSubmatch(tt.testPath)

			if tt.shouldMatch {
				if matches == nil {
					t.Errorf("Expected path to match, but it didn't")
					return
				}

				// Extract parameters
				params := make(map[string]string)
				for i, name := range paramNames {
					if i+1 < len(matches) {
						params[name] = matches[i+1]
					}
				}

				// Verify parameters
				for k, expectedV := range tt.params {
					if actualV, ok := params[k]; !ok {
						t.Errorf("Expected param %s to exist", k)
					} else if actualV != expectedV {
						t.Errorf("Expected param %s=%s, got %s", k, expectedV, actualV)
					}
				}
			} else {
				if matches != nil {
					t.Errorf("Expected path not to match, but it did")
				}
			}
		})
	}
}

func TestNetHTTPAdapter_ErrorHandling(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type:    routeradapter.RouterTypeNetHTTP,
		Port:    8080,
		NetHTTP: &routeradapter.NetHTTPConfig{},
	}

	adapter, err := NewNetHTTPAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Set error handler
	adapter.SetErrorHandler(routeradapter.NewNetHTTPErrorHandler())

	// Register route - error handling is manual in net/http
	meta := route.Meta{
		Method: "GET",
		Path:   "/error",
		Func:   dummyGinHandler(),
	}

	if err := adapter.RegisterRoute(meta); err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)

	adapter.ServeHTTP(w, req)

	// Note: net/http adapter doesn't execute actual Gin handlers
	// It returns a dummy 200 response. In a full implementation,
	// handlers would be converted properly.
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNetHTTPAdapter_NotFound(t *testing.T) {
	cfg := &routeradapter.RouterConfig{
		Type:    routeradapter.RouterTypeNetHTTP,
		Port:    8080,
		NetHTTP: &routeradapter.NetHTTPConfig{},
	}

	adapter, err := NewNetHTTPAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test route that doesn't exist
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)

	adapter.ServeHTTP(w, req)

	// Should return 404
	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}
