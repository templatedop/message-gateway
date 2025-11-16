package middlewares

import (
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Benchmark payloads
var (
	smallPayload  = "Hello, World!"                                            // ~13 bytes
	mediumPayload = strings.Repeat("Hello, World! This is test data. ", 50)   // ~1.8 KB
	largePayload  = strings.Repeat("Hello, World! This is test data. ", 500)  // ~18 KB
	hugePayload   = strings.Repeat("Hello, World! This is test data. ", 2000) // ~72 KB

	// JSON-like payload
	jsonSmall = `{"id":1,"name":"John Doe","email":"john@example.com"}`
	jsonMedium = strings.Repeat(`{"id":1,"name":"John Doe","email":"john@example.com","address":"123 Main St","city":"New York","country":"USA","phone":"+1-555-1234"},`, 50)
	jsonLarge = strings.Repeat(`{"id":1,"name":"John Doe","email":"john@example.com","address":"123 Main St","city":"New York","country":"USA","phone":"+1-555-1234"},`, 500)
)

// setupRouter creates a test router with or without gzip
func setupRouter(useGzip bool, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	if useGzip {
		router.Use(Gzip())
	}

	router.GET("/test", handler)
	return router
}

// Benchmarks WITHOUT gzip compression

func BenchmarkWithoutGzip_SmallPayload(b *testing.B) {
	router := setupRouter(false, func(c *gin.Context) {
		c.String(http.StatusOK, smallPayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutGzip_MediumPayload(b *testing.B) {
	router := setupRouter(false, func(c *gin.Context) {
		c.String(http.StatusOK, mediumPayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutGzip_LargePayload(b *testing.B) {
	router := setupRouter(false, func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutGzip_HugePayload(b *testing.B) {
	router := setupRouter(false, func(c *gin.Context) {
		c.String(http.StatusOK, hugePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutGzip_JSONSmall(b *testing.B) {
	router := setupRouter(false, func(c *gin.Context) {
		c.String(http.StatusOK, jsonSmall)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutGzip_JSONMedium(b *testing.B) {
	router := setupRouter(false, func(c *gin.Context) {
		c.String(http.StatusOK, jsonMedium)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutGzip_JSONLarge(b *testing.B) {
	router := setupRouter(false, func(c *gin.Context) {
		c.String(http.StatusOK, jsonLarge)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Benchmarks WITH gzip compression

func BenchmarkWithGzip_SmallPayload(b *testing.B) {
	router := setupRouter(true, func(c *gin.Context) {
		c.String(http.StatusOK, smallPayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithGzip_MediumPayload(b *testing.B) {
	router := setupRouter(true, func(c *gin.Context) {
		c.String(http.StatusOK, mediumPayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithGzip_LargePayload(b *testing.B) {
	router := setupRouter(true, func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithGzip_HugePayload(b *testing.B) {
	router := setupRouter(true, func(c *gin.Context) {
		c.String(http.StatusOK, hugePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithGzip_JSONSmall(b *testing.B) {
	router := setupRouter(true, func(c *gin.Context) {
		c.String(http.StatusOK, jsonSmall)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithGzip_JSONMedium(b *testing.B) {
	router := setupRouter(true, func(c *gin.Context) {
		c.String(http.StatusOK, jsonMedium)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithGzip_JSONLarge(b *testing.B) {
	router := setupRouter(true, func(c *gin.Context) {
		c.String(http.StatusOK, jsonLarge)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Benchmarks with different compression levels

func BenchmarkWithGzip_BestSpeed_LargePayload(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	config := DefaultGzipConfig()
	config.CompressionLevel = gzip.BestSpeed
	router.Use(GzipWithConfig(config))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithGzip_DefaultCompression_LargePayload(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	config := DefaultGzipConfig()
	config.CompressionLevel = gzip.DefaultCompression
	router.Use(GzipWithConfig(config))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithGzip_BestCompression_LargePayload(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	config := DefaultGzipConfig()
	config.CompressionLevel = gzip.BestCompression
	router.Use(GzipWithConfig(config))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Benchmark concurrent requests

func BenchmarkWithGzip_Concurrent(b *testing.B) {
	router := setupRouter(true, func(c *gin.Context) {
		c.String(http.StatusOK, mediumPayload)
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkWithoutGzip_Concurrent(b *testing.B) {
	router := setupRouter(false, func(c *gin.Context) {
		c.String(http.StatusOK, mediumPayload)
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}
