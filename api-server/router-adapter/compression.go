package routeradapter

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// gzipResponseWriter wraps http.ResponseWriter to support gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// gzipWriterPools - one pool per compression level for optimal performance
var gzipWriterPools [10]sync.Pool

func init() {
	// Initialize pools for each compression level (0-9)
	for i := 0; i < 10; i++ {
		level := i
		gzipWriterPools[i] = sync.Pool{
			New: func() interface{} {
				w, _ := gzip.NewWriterLevel(nil, level)
				return w
			},
		}
	}
}

// GzipMiddleware returns a framework-agnostic gzip compression middleware
func GzipMiddleware(level int) MiddlewareFunc {
	return func(ctx *RouterContext, next func() error) error {
		// Check if client accepts gzip
		if !strings.Contains(ctx.Request.Header.Get("Accept-Encoding"), "gzip") {
			return next()
		}

		// Validate and normalize compression level
		if level < gzip.NoCompression || level > gzip.BestCompression {
			level = gzip.DefaultCompression
		}

		// Get writer from correct pool for this compression level
		gz := gzipWriterPools[level].Get().(*gzip.Writer)
		gz.Reset(ctx.Response)
		defer func() {
			gz.Close()
			gzipWriterPools[level].Put(gz)
		}()

		// Set response headers
		ctx.SetHeader("Content-Encoding", "gzip")
		ctx.SetHeader("Vary", "Accept-Encoding")

		// Wrap response writer
		gzipWriter := &gzipResponseWriter{
			ResponseWriter: ctx.Response,
			Writer:         gz,
		}

		// Replace the response writer
		originalWriter := ctx.Response
		ctx.Response = gzipWriter

		// Call next middleware/handler
		err := next()

		// Restore original writer
		ctx.Response = originalWriter

		return err
	}
}

// ShouldCompress determines if the response should be compressed based on content type
func ShouldCompress(contentType string) bool {
	// List of compressible content types
	compressibleTypes := []string{
		"text/",
		"application/json",
		"application/javascript",
		"application/xml",
		"application/x-javascript",
		"image/svg+xml",
	}

	contentType = strings.ToLower(contentType)
	for _, typ := range compressibleTypes {
		if strings.HasPrefix(contentType, typ) {
			return true
		}
	}

	return false
}
