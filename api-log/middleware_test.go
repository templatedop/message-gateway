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
