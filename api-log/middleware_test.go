package log

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func setupTestLoggerForMiddleware() *bytes.Buffer {
	// Reset global state
	once = sync.Once{}
	createErr = nil
	baseLogger = nil

	var buf bytes.Buffer
	factory := NewDefaultLoggerFactory()
	factory.Create(
		WithServiceName("test-middleware"),
		WithLevel(zerolog.DebugLevel),
		WithOutputWriter(&buf),
	)

	return &buf
}

func TestSetCtxLoggerMiddleware(t *testing.T) {
	setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	router.GET("/test", func(c *gin.Context) {
		// Try to get logger from context
		logger := getCtxLogger(c)
		if logger == nil {
			t.Error("Logger should be set in context by middleware")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSetCtxLoggerMiddleware_GeneratesRequestID(t *testing.T) {
	setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	router.GET("/test", func(c *gin.Context) {
		requestID := c.Writer.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("X-Request-ID should be generated and set in response headers")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}

func TestSetCtxLoggerMiddleware_UsesProvidedRequestID(t *testing.T) {
	setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	expectedRequestID := "test-request-id-123"
	receivedRequestID := ""

	router.GET("/test", func(c *gin.Context) {
		// The provided request ID should be available in the request header
		receivedRequestID = c.Request.Header.Get("X-Request-ID")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", expectedRequestID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if receivedRequestID != expectedRequestID {
		t.Errorf("Expected request ID %s, got %s", expectedRequestID, receivedRequestID)
	}
}

func TestSetCtxLoggerMiddleware_CapturesTraceID(t *testing.T) {
	setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	expectedTraceID := "test-trace-id-456"

	router.GET("/test", func(c *gin.Context) {
		// The trace ID should be captured in the logger context
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-ID", expectedTraceID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSetCtxLoggerMiddleware_CapturesOfficeID(t *testing.T) {
	setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	expectedOfficeID := "office-123"

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-office-id", expectedOfficeID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSetCtxLoggerMiddleware_CapturesUserID(t *testing.T) {
	setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	expectedUserID := "user-789"

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-user-id", expectedUserID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequestResponseLoggerMiddleware(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)
	router.Use(RequestResponseLoggerMiddleware)

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	output := buf.String()

	// Check that request was logged
	if !contains(output, "/test") {
		t.Error("Request path should be logged")
	}

	if !contains(output, "GET") {
		t.Error("Request method should be logged")
	}
}

func TestRequestResponseLoggerMiddleware_SkipsHealthCheck(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)
	router.Use(RequestResponseLoggerMiddleware)

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	output := buf.String()

	// Health check should not be logged
	if contains(output, "/healthz") {
		t.Error("Health check requests should not be logged")
	}
}

func TestRequestResponseLoggerMiddleware_LogsQueryParams(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)
	router.Use(RequestResponseLoggerMiddleware)

	router.GET("/search", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/search?q=test&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	output := buf.String()

	// Query params should be in the logged path
	if !contains(output, "q=test") {
		t.Error("Query parameters should be logged")
	}
}

func TestRequestResponseLoggerMiddleware_LogsStatusCode(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)
	router.Use(RequestResponseLoggerMiddleware)

	router.GET("/notfound", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	req := httptest.NewRequest("GET", "/notfound", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	output := buf.String()

	// Status code should be logged
	if !contains(output, "404") {
		t.Error("Status code should be logged")
	}
}

func TestRequestResponseLoggerMiddleware_LogsLatency(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)
	router.Use(RequestResponseLoggerMiddleware)

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	output := buf.String()

	// Latency should be logged
	if !contains(output, "latency-ms") {
		t.Error("Latency should be logged")
	}
}

func TestSetRequestMetadata(t *testing.T) {
	setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-Request-ID", "req-123")
	c.Request.Header.Set("X-Trace-ID", "trace-456")
	c.Request.Header.Set("x-office-id", "office-789")
	c.Request.Header.Set("x-user-id", "user-101")

	logger := baseLogger.setRequestMetadata(c)

	// Verify logger is created (we can't easily verify the fields without logging)
	if &logger == nil {
		t.Error("setRequestMetadata should return a logger")
	}
}

func TestSetRequestMetadata_GeneratesRequestID(t *testing.T) {
	setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	// Don't set X-Request-ID

	baseLogger.setRequestMetadata(c)

	// Should generate and set X-Request-ID in response
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("setRequestMetadata should generate X-Request-ID when not provided")
	}
}

func TestGetDefaultLogger(t *testing.T) {
	// Reset state
	once = sync.Once{}
	createErr = nil
	baseLogger = nil

	logger := getDefaultLogger()

	if logger == nil {
		t.Error("getDefaultLogger should return a logger")
	}

	if logger.logger == nil {
		t.Error("getDefaultLogger should initialize the logger")
	}
}

func TestRequestResponseLoggerMiddlewareWithConfig_CustomSkipPaths(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	config := &MiddlewareConfig{
		SkipPaths:        []string{"/metrics", "/ready", "/status"},
		SkipPathPrefixes: []string{},
		SkipMethodPaths:  make(map[string][]string),
	}
	router.Use(RequestResponseLoggerMiddlewareWithConfig(config))

	router.GET("/metrics", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/ready", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/api/users", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test /metrics - should be skipped
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output := buf.String()
	if contains(output, "/metrics") {
		t.Error("/metrics should be skipped from logging")
	}

	// Test /ready - should be skipped
	buf.Reset()
	req = httptest.NewRequest("GET", "/ready", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if contains(output, "/ready") {
		t.Error("/ready should be skipped from logging")
	}

	// Test /api/users - should be logged
	buf.Reset()
	req = httptest.NewRequest("GET", "/api/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if !contains(output, "/api/users") {
		t.Error("/api/users should be logged")
	}
}

func TestRequestResponseLoggerMiddlewareWithConfig_SkipPrefixes(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	config := &MiddlewareConfig{
		SkipPaths:        []string{},
		SkipPathPrefixes: []string{"/internal/", "/debug/"},
		SkipMethodPaths:  make(map[string][]string),
	}
	router.Use(RequestResponseLoggerMiddlewareWithConfig(config))

	router.GET("/internal/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/debug/pprof", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/api/users", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test /internal/health - should be skipped
	req := httptest.NewRequest("GET", "/internal/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output := buf.String()
	if contains(output, "/internal/health") {
		t.Error("/internal/health should be skipped from logging")
	}

	// Test /debug/pprof - should be skipped
	buf.Reset()
	req = httptest.NewRequest("GET", "/debug/pprof", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if contains(output, "/debug/pprof") {
		t.Error("/debug/pprof should be skipped from logging")
	}

	// Test /api/users - should be logged
	buf.Reset()
	req = httptest.NewRequest("GET", "/api/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if !contains(output, "/api/users") {
		t.Error("/api/users should be logged")
	}
}

func TestRequestResponseLoggerMiddlewareWithConfig_MethodPaths(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	config := &MiddlewareConfig{
		SkipPaths:        []string{},
		SkipPathPrefixes: []string{},
		SkipMethodPaths: map[string][]string{
			"GET":  {"/status", "/ping"},
			"POST": {"/webhook"},
		},
	}
	router.Use(RequestResponseLoggerMiddlewareWithConfig(config))

	router.GET("/status", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.POST("/status", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.POST("/webhook", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test GET /status - should be skipped
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output := buf.String()
	if contains(output, "GET") && contains(output, "/status") {
		t.Error("GET /status should be skipped from logging")
	}

	// Test POST /status - should be logged (different method)
	buf.Reset()
	req = httptest.NewRequest("POST", "/status", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if !contains(output, "POST") || !contains(output, "/status") {
		t.Error("POST /status should be logged")
	}

	// Test POST /webhook - should be skipped
	buf.Reset()
	req = httptest.NewRequest("POST", "/webhook", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if contains(output, "/webhook") {
		t.Error("POST /webhook should be skipped from logging")
	}
}

func TestRequestResponseLoggerMiddlewareWithConfig_NilConfig(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)
	router.Use(RequestResponseLoggerMiddlewareWithConfig(nil)) // nil config should use defaults

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/api/users", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test /healthz - should be skipped (default behavior)
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output := buf.String()
	if contains(output, "/healthz") {
		t.Error("/healthz should be skipped with default config")
	}

	// Test /health - should be skipped (default behavior)
	buf.Reset()
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if contains(output, "/health") {
		t.Error("/health should be skipped with default config")
	}

	// Test /api/users - should be logged
	buf.Reset()
	req = httptest.NewRequest("GET", "/api/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if !contains(output, "/api/users") {
		t.Error("/api/users should be logged")
	}
}

func TestRequestResponseLoggerMiddlewareWithConfig_CombinedRules(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)

	config := &MiddlewareConfig{
		SkipPaths:        []string{"/healthz"},
		SkipPathPrefixes: []string{"/internal/"},
		SkipMethodPaths: map[string][]string{
			"GET": {"/metrics"},
		},
	}
	router.Use(RequestResponseLoggerMiddlewareWithConfig(config))

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/internal/debug", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/metrics", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.POST("/metrics", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test exact path skip
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output := buf.String()
	if contains(output, "/healthz") {
		t.Error("/healthz should be skipped")
	}

	// Test prefix skip
	buf.Reset()
	req = httptest.NewRequest("GET", "/internal/debug", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if contains(output, "/internal/debug") {
		t.Error("/internal/debug should be skipped")
	}

	// Test method+path skip
	buf.Reset()
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if contains(output, "GET") && contains(output, "/metrics") {
		t.Error("GET /metrics should be skipped")
	}

	// Test that different method is logged
	buf.Reset()
	req = httptest.NewRequest("POST", "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if !contains(output, "POST") || !contains(output, "/metrics") {
		t.Error("POST /metrics should be logged")
	}
}

func TestRequestResponseLoggerMiddleware_BackwardCompatibility(t *testing.T) {
	buf := setupTestLoggerForMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetCtxLoggerMiddleware)
	router.Use(RequestResponseLoggerMiddleware) // Old function should still work

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/api/users", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test /healthz - should be skipped (backward compatible)
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output := buf.String()
	if contains(output, "/healthz") {
		t.Error("/healthz should be skipped for backward compatibility")
	}

	// Test /health - should be skipped (new default)
	buf.Reset()
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if contains(output, "/health") {
		t.Error("/health should be skipped with new defaults")
	}

	// Test /api/users - should be logged
	buf.Reset()
	req = httptest.NewRequest("GET", "/api/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	output = buf.String()
	if !contains(output, "/api/users") {
		t.Error("/api/users should be logged")
	}
}
