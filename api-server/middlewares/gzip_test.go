package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestGzip_SmallResponse tests that small responses are not compressed
func TestGzip_SmallResponse(t *testing.T) {
	router := gin.New()
	router.Use(Gzip())

	smallPayload := "Hello, World!"
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, smallPayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Small responses should not be compressed
	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, smallPayload, w.Body.String())
}

// TestGzip_LargeResponse tests that large responses are compressed
func TestGzip_LargeResponse(t *testing.T) {
	router := gin.New()
	router.Use(Gzip())

	// Create a payload larger than MinCompressionSize (1KB)
	largePayload := strings.Repeat("Hello, World! This is a test payload. ", 100) // ~3.8KB

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", w.Header().Get("Vary"))

	// Decompress and verify content
	reader, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.Equal(t, largePayload, string(decompressed))
}

// TestGzip_NoAcceptEncoding tests that responses are not compressed when client doesn't accept gzip
func TestGzip_NoAcceptEncoding(t *testing.T) {
	router := gin.New()
	router.Use(Gzip())

	largePayload := strings.Repeat("Hello, World! ", 200)

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No Accept-Encoding header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, largePayload, w.Body.String())
}

// TestGzip_JSONResponse tests compression of JSON responses
func TestGzip_JSONResponse(t *testing.T) {
	router := gin.New()
	router.Use(Gzip())

	// Create a large JSON response
	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Address  string `json:"address"`
		City     string `json:"city"`
		Country  string `json:"country"`
		Phone    string `json:"phone"`
		Company  string `json:"company"`
	}

	users := make([]User, 100)
	for i := 0; i < 100; i++ {
		users[i] = User{
			ID:      i,
			Name:    "John Doe",
			Email:   "john.doe@example.com",
			Address: "123 Main St, Apt 4B",
			City:    "New York",
			Country: "USA",
			Phone:   "+1-555-123-4567",
			Company: "Acme Corporation",
		}
	}

	router.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, users)
	})

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Verify we can decompress the response
	reader, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)

	// Verify it's valid JSON
	assert.Contains(t, string(decompressed), `"name":"John Doe"`)
}

// TestGzip_ExcludedPaths tests that excluded paths are not compressed
func TestGzip_ExcludedPaths(t *testing.T) {
	config := DefaultGzipConfig()
	config.ExcludedPaths = []string{"/debug/", "/metrics"}

	router := gin.New()
	router.Use(GzipWithConfig(config))

	largePayload := strings.Repeat("Debug info ", 200)

	router.GET("/debug/pprof", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	router.GET("/api/data", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	tests := []struct {
		path             string
		shouldCompress   bool
	}{
		{"/debug/pprof", false},
		{"/metrics", false},
		{"/api/data", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.shouldCompress {
				assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
			} else {
				assert.Empty(t, w.Header().Get("Content-Encoding"))
			}
		})
	}
}

// TestGzip_ExcludedExtensions tests that excluded file extensions are not compressed
func TestGzip_ExcludedExtensions(t *testing.T) {
	config := DefaultGzipConfig()

	router := gin.New()
	router.Use(GzipWithConfig(config))

	// Create a large "image" payload
	largePayload := strings.Repeat("binary data ", 200)

	router.GET("/image.png", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	router.GET("/data.json", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	tests := []struct {
		path           string
		shouldCompress bool
	}{
		{"/image.png", false},
		{"/data.json", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.shouldCompress {
				assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
			} else {
				assert.Empty(t, w.Header().Get("Content-Encoding"))
			}
		})
	}
}

// TestGzip_CustomCompressionLevel tests custom compression levels
func TestGzip_CustomCompressionLevel(t *testing.T) {
	config := DefaultGzipConfig()
	config.CompressionLevel = gzip.BestCompression

	router := gin.New()
	router.Use(GzipWithConfig(config))

	largePayload := strings.Repeat("Hello, World! ", 200)

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Verify we can decompress the response
	reader, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.Equal(t, largePayload, string(decompressed))
}

// TestGzip_CustomMinSize tests custom minimum size threshold
func TestGzip_CustomMinSize(t *testing.T) {
	config := DefaultGzipConfig()
	config.MinSize = 100 // Very small threshold

	router := gin.New()
	router.Use(GzipWithConfig(config))

	// Payload between default MinSize (1KB) and custom MinSize (100B)
	payload := strings.Repeat("Hello! ", 20) // ~140 bytes

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, payload)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Verify we can decompress the response
	reader, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.Equal(t, payload, string(decompressed))
}

// TestGzip_MultipleRequests tests that the middleware handles multiple requests correctly
func TestGzip_MultipleRequests(t *testing.T) {
	router := gin.New()
	router.Use(Gzip())

	largePayload := strings.Repeat("Test data ", 200)

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	// Make multiple requests to ensure writer pool works correctly
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		// Verify we can decompress the response
		reader, err := gzip.NewReader(w.Body)
		require.NoError(t, err)

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)
		reader.Close()

		assert.Equal(t, largePayload, string(decompressed))
	}
}

// TestGzip_StatusCodes tests that gzip works with different status codes
func TestGzip_StatusCodes(t *testing.T) {
	router := gin.New()
	router.Use(Gzip())

	largePayload := strings.Repeat("Error message ", 200)

	router.GET("/success", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	router.GET("/error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, largePayload)
	})

	router.GET("/notfound", func(c *gin.Context) {
		c.String(http.StatusNotFound, largePayload)
	})

	statusCodes := []int{http.StatusOK, http.StatusInternalServerError, http.StatusNotFound}
	paths := []string{"/success", "/error", "/notfound"}

	for i, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, statusCodes[i], w.Code)
			assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

			// Verify we can decompress the response
			reader, err := gzip.NewReader(w.Body)
			require.NoError(t, err)
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.Equal(t, largePayload, string(decompressed))
		})
	}
}

// TestGzip_ChunkedResponse tests that gzip handles streaming/chunked responses
func TestGzip_ChunkedResponse(t *testing.T) {
	router := gin.New()
	router.Use(Gzip())

	router.GET("/stream", func(c *gin.Context) {
		// Simulate streaming response by writing in chunks
		for i := 0; i < 10; i++ {
			c.Writer.WriteString(strings.Repeat("Chunk data ", 20))
		}
	})

	req := httptest.NewRequest("GET", "/stream", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Verify we can decompress the response
	reader, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)

	// Verify all chunks are present
	expectedContent := strings.Repeat(strings.Repeat("Chunk data ", 20), 10)
	assert.Equal(t, expectedContent, string(decompressed))
}

// TestGzip_CompressionRatio tests that gzip actually reduces response size
func TestGzip_CompressionRatio(t *testing.T) {
	router := gin.New()
	router.Use(Gzip())

	// Highly compressible data
	largePayload := strings.Repeat("AAAAAAAAAABBBBBBBBBBCCCCCCCCCC", 200) // ~6KB

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	// Request with gzip
	reqGzip := httptest.NewRequest("GET", "/test", nil)
	reqGzip.Header.Set("Accept-Encoding", "gzip")
	wGzip := httptest.NewRecorder()
	router.ServeHTTP(wGzip, reqGzip)

	// Request without gzip
	router2 := gin.New()
	router2.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, largePayload)
	})

	reqNoGzip := httptest.NewRequest("GET", "/test", nil)
	wNoGzip := httptest.NewRecorder()
	router2.ServeHTTP(wNoGzip, reqNoGzip)

	compressedSize := wGzip.Body.Len()
	uncompressedSize := wNoGzip.Body.Len()

	// Verify compression actually reduces size
	assert.Less(t, compressedSize, uncompressedSize)

	// Calculate compression ratio
	ratio := float64(compressedSize) / float64(uncompressedSize)
	t.Logf("Compression ratio: %.2f%% (compressed: %d bytes, original: %d bytes)",
		ratio*100, compressedSize, uncompressedSize)

	// For highly repetitive data, we should get good compression
	assert.Less(t, ratio, 0.5) // At least 50% compression
}

// TestDefaultGzipConfig tests the default configuration
func TestDefaultGzipConfig(t *testing.T) {
	config := DefaultGzipConfig()

	assert.Equal(t, DefaultCompressionLevel, config.CompressionLevel)
	assert.Equal(t, MinCompressionSize, config.MinSize)
	assert.NotEmpty(t, config.ExcludedExtensions)
	assert.NotEmpty(t, config.ExcludedPaths)

	// Verify some expected excluded extensions
	assert.Contains(t, config.ExcludedExtensions, ".png")
	assert.Contains(t, config.ExcludedExtensions, ".jpg")
	assert.Contains(t, config.ExcludedExtensions, ".zip")

	// Verify some expected excluded paths
	assert.Contains(t, config.ExcludedPaths, "/debug/")
	assert.Contains(t, config.ExcludedPaths, "/metrics")
}
