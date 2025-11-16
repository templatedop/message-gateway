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

// gzipWriterPool pools gzip writers for reuse
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// GzipMiddleware returns a framework-agnostic gzip compression middleware
func GzipMiddleware(level int) MiddlewareFunc {
	return func(ctx *RouterContext, next func() error) error {
		// Check if client accepts gzip
		if !strings.Contains(ctx.Request.Header.Get("Accept-Encoding"), "gzip") {
			return next()
		}

		// Get gzip writer from pool
		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)

		// Reset writer with new level
		if level >= gzip.NoCompression && level <= gzip.BestCompression {
			gz.Reset(ctx.Response)
			// Note: We can't change the level of an existing writer easily
			// So we create a new one with the desired level
			var err error
			gz, err = gzip.NewWriterLevel(ctx.Response, level)
			if err != nil {
				return next() // Fall back to no compression
			}
		} else {
			gz.Reset(ctx.Response)
		}
		defer gz.Close()

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
