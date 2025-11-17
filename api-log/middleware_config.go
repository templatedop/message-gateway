package log

import "strings"

// MiddlewareConfig configures which paths should skip logging in the middleware.
// This is useful for health checks, metrics endpoints, and other high-frequency
// requests that don't need to be logged.
type MiddlewareConfig struct {
	// SkipPaths contains exact paths to skip logging (e.g., "/healthz", "/metrics")
	SkipPaths []string

	// SkipPathPrefixes contains path prefixes to skip logging (e.g., "/internal/", "/debug/")
	// Any path starting with these prefixes will be skipped
	SkipPathPrefixes []string

	// SkipMethodPaths contains method-specific paths to skip
	// Map keys are HTTP methods (e.g., "GET", "POST")
	// Map values are paths to skip for that method
	SkipMethodPaths map[string][]string
}

// DefaultMiddlewareConfig returns a MiddlewareConfig with sensible defaults.
// By default, it skips common health check paths: /healthz and /health
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		SkipPaths:        []string{"/healthz", "/health"},
		SkipPathPrefixes: []string{},
		SkipMethodPaths:  make(map[string][]string),
	}
}

// ShouldSkip determines if a request with the given method and path should skip logging.
// It checks in order:
// 1. Exact path matches in SkipPaths
// 2. Path prefix matches in SkipPathPrefixes
// 3. Method+path combinations in SkipMethodPaths
//
// Returns true if the request should skip logging, false otherwise.
func (c *MiddlewareConfig) ShouldSkip(method, path string) bool {
	if c == nil {
		return false
	}

	// Check exact paths
	for _, skipPath := range c.SkipPaths {
		if path == skipPath {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range c.SkipPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	// Check method+path combinations
	if paths, ok := c.SkipMethodPaths[method]; ok {
		for _, skipPath := range paths {
			if path == skipPath {
				return true
			}
		}
	}

	return false
}
