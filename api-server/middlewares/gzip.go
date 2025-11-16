package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultCompressionLevel is the default gzip compression level
	DefaultCompressionLevel = gzip.DefaultCompression

	// MinCompressionSize is the minimum response size to enable compression
	MinCompressionSize = 1024 // 1KB
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		writer, _ := gzip.NewWriterLevel(io.Discard, DefaultCompressionLevel)
		return writer
	},
}

// GzipConfig holds the configuration for gzip compression
type GzipConfig struct {
	// CompressionLevel is the gzip compression level (0-9, where 9 is maximum compression)
	CompressionLevel int

	// MinSize is the minimum response size (in bytes) to enable compression
	MinSize int

	// ExcludedExtensions contains file extensions that should not be compressed
	ExcludedExtensions []string

	// ExcludedPaths contains URL paths that should not be compressed
	ExcludedPaths []string
}

// DefaultGzipConfig returns the default gzip configuration
func DefaultGzipConfig() GzipConfig {
	return GzipConfig{
		CompressionLevel: DefaultCompressionLevel,
		MinSize:          MinCompressionSize,
		ExcludedExtensions: []string{
			".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".ico",
			".mp4", ".avi", ".mov", ".mp3", ".wav",
			".zip", ".tar", ".gz", ".bz2", ".7z",
			".pdf", ".woff", ".woff2", ".ttf", ".eot",
		},
		ExcludedPaths: []string{
			"/debug/",
			"/metrics",
		},
	}
}

// gzipWriter wraps gin.ResponseWriter to intercept and compress response data
type gzipWriter struct {
	gin.ResponseWriter
	writer       *gzip.Writer
	wroteHeader  bool
	minSize      int
	buffer       []byte
	statusCode   int
	shouldCompress bool
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	// If we haven't written the header yet, buffer the data
	if !g.wroteHeader {
		g.buffer = append(g.buffer, data...)

		// If buffer is large enough, decide whether to compress
		if len(g.buffer) >= g.minSize {
			g.flushBuffer()
		}
		return len(data), nil
	}

	// Header already written, write directly
	if g.shouldCompress && g.writer != nil {
		return g.writer.Write(data)
	}
	return g.ResponseWriter.Write(data)
}

func (g *gzipWriter) WriteHeader(code int) {
	g.statusCode = code
	// Don't write header yet, wait for first write to determine compression
}

func (g *gzipWriter) flushBuffer() {
	if g.wroteHeader {
		return
	}

	g.wroteHeader = true

	// Determine if we should compress
	g.shouldCompress = len(g.buffer) >= g.minSize

	if g.shouldCompress && g.writer != nil {
		g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		g.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
		g.ResponseWriter.Header().Del("Content-Length")
		g.ResponseWriter.WriteHeader(g.statusCode)
		g.writer.Write(g.buffer)
	} else {
		// Don't compress, write original response
		g.ResponseWriter.WriteHeader(g.statusCode)
		g.ResponseWriter.Write(g.buffer)
	}

	g.buffer = nil
}

// Gzip returns a middleware that compresses HTTP responses using gzip
func Gzip() gin.HandlerFunc {
	return GzipWithConfig(DefaultGzipConfig())
}

// GzipWithConfig returns a middleware that compresses HTTP responses using gzip with custom configuration
func GzipWithConfig(config GzipConfig) gin.HandlerFunc {
	// Ensure defaults are set
	if config.CompressionLevel < gzip.DefaultCompression || config.CompressionLevel > gzip.BestCompression {
		config.CompressionLevel = DefaultCompressionLevel
	}
	if config.MinSize <= 0 {
		config.MinSize = MinCompressionSize
	}

	return func(c *gin.Context) {
		// Check if client accepts gzip encoding
		if !shouldCompress(c.Request, config) {
			c.Next()
			return
		}

		// Get gzip writer from pool
		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer func() {
			gz.Reset(io.Discard)
			gzipWriterPool.Put(gz)
		}()

		// Create custom response writer
		gzWriter := &gzipWriter{
			ResponseWriter: c.Writer,
			writer:        gz,
			minSize:       config.MinSize,
			buffer:        make([]byte, 0, config.MinSize),
			statusCode:    http.StatusOK,
		}

		// Reset gzip writer to use our custom writer
		gz.Reset(gzWriter.ResponseWriter)

		// Replace the response writer
		c.Writer = gzWriter

		// Process request
		c.Next()

		// Flush any remaining buffered data
		gzWriter.flushBuffer()

		// Close gzip writer if compression was used
		if gzWriter.shouldCompress && gzWriter.writer != nil {
			gz.Close()
		}
	}
}

// shouldCompress determines if the request should be compressed
func shouldCompress(req *http.Request, config GzipConfig) bool {
	// Check if client accepts gzip encoding
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}

	// Check if path should be excluded
	for _, path := range config.ExcludedPaths {
		if strings.HasPrefix(req.URL.Path, path) {
			return false
		}
	}

	// Check if extension should be excluded
	for _, ext := range config.ExcludedExtensions {
		if strings.HasSuffix(req.URL.Path, ext) {
			return false
		}
	}

	return true
}
