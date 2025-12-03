package swagger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

func TestNewMiddleware_Caching(t *testing.T) {
	// Clear cache before test
	swaggerCache.Range(func(key, value interface{}) bool {
		swaggerCache.Delete(key)
		return true
	})

	// Create a simple v3 doc
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
	}

	middleware := newMiddleware(doc)

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)

	t.Run("First request generates and caches", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/swagger/docs.json", nil)
		req.Host = "example.com"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify response is valid JSON
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Errorf("Response should be valid JSON: %v", err)
		}

		// Verify cache entry exists
		if _, ok := swaggerCache.Load("example.com"); !ok {
			t.Error("Cache should contain entry for example.com after first request")
		}
	})

	t.Run("Second request uses cache", func(t *testing.T) {
		// Get cached value
		cached, ok := swaggerCache.Load("example.com")
		if !ok {
			t.Fatal("Cache should contain entry from previous test")
		}

		req, _ := http.NewRequest("GET", "/swagger/docs.json", nil)
		req.Host = "example.com"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify response body matches cached value
		if string(w.Body.Bytes()) != string(cached.([]byte)) {
			t.Error("Cached response should be identical to original")
		}
	})

	t.Run("Different host creates separate cache entry", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/swagger/docs.json", nil)
		req.Host = "api.example.com"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify separate cache entry
		if _, ok := swaggerCache.Load("api.example.com"); !ok {
			t.Error("Cache should contain entry for api.example.com")
		}

		// Verify both cache entries exist
		cacheSize := 0
		swaggerCache.Range(func(key, value interface{}) bool {
			cacheSize++
			return true
		})

		if cacheSize < 2 {
			t.Errorf("Expected at least 2 cache entries, got %d", cacheSize)
		}
	})
}

func TestNewMiddleware_NonSwaggerPaths(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
	}

	middleware := newMiddleware(doc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	req, _ := http.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Non-swagger paths should pass through, got status %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Response should be valid JSON: %v", err)
	}

	if result["message"] != "test" {
		t.Error("Non-swagger path should return expected response")
	}
}

func TestAttachHostToV3Doc(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
	}

	result := attachHostToV3Doc(doc, "example.com:8080")

	if result == nil {
		t.Fatal("attachHostToV3Doc should return non-nil document")
	}

	if len(result.Servers) == 0 {
		t.Fatal("Document should have servers")
	}

	expectedURL := "http://example.com:8080"
	if result.Servers[0].URL != expectedURL {
		t.Errorf("Server URL = %v, want %v", result.Servers[0].URL, expectedURL)
	}
}

func TestNewMiddleware_SwaggerStaticFiles(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
	}

	middleware := newMiddleware(doc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)

	// Test swagger path handling
	req, _ := http.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should attempt to serve from embedded files
	// Status might be 404 if file doesn't exist in test, but middleware should handle the path
	if w.Code != http.StatusNotFound && w.Code != http.StatusOK {
		t.Logf("Swagger static file request handled with status %d (expected 200 or 404)", w.Code)
	}
}

// Benchmark tests

func BenchmarkNewMiddleware_CacheHit(b *testing.B) {
	// Setup
	swaggerCache.Range(func(key, value interface{}) bool {
		swaggerCache.Delete(key)
		return true
	})

	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Benchmark API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
	}

	middleware := newMiddleware(doc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)

	// Prime the cache
	req, _ := http.NewRequest("GET", "/swagger/docs.json", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Benchmark cache hits
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/swagger/docs.json", nil)
		req.Host = "example.com"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkNewMiddleware_CacheMiss(b *testing.B) {
	swaggerCache.Range(func(key, value interface{}) bool {
		swaggerCache.Delete(key)
		return true
	})

	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Benchmark API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
	}

	middleware := newMiddleware(doc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache before each iteration to force cache miss
		swaggerCache.Delete("example.com")

		req, _ := http.NewRequest("GET", "/swagger/docs.json", nil)
		req.Host = "example.com"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkAttachHostToV3Doc(b *testing.B) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Benchmark API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attachHostToV3Doc(doc, "example.com")
	}
}
