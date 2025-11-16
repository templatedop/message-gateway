package middlewares

import (
	"strconv"
	"time"

	"MgApplication/api-server/router-adapter"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	HTTPServerMetricsRequestsCount    = "http_server_requests_total"
	HTTPServerMetricsRequestsDuration = "http_server_requests_duration_seconds"
	HTTPServerMetricsNotFoundPath     = "/not-found"
)

// MetricsConfig configures the metrics middleware
type MetricsConfig struct {
	// Skipper allows skipping metrics for specific requests
	Skipper func(ctx *routeradapter.RouterContext) bool

	// Registry is the Prometheus registry to use
	// If nil, uses prometheus.DefaultRegisterer
	Registry prometheus.Registerer

	// Namespace is the Prometheus namespace
	Namespace string

	// Subsystem is the Prometheus subsystem
	Subsystem string

	// Buckets are the histogram buckets for request duration
	// If nil, uses prometheus.DefBuckets
	Buckets []float64

	// NormalizeRequestPath normalizes dynamic paths (e.g., /users/123 -> /users/:id)
	NormalizeRequestPath bool

	// NormalizeResponseStatus normalizes HTTP status codes (e.g., 201 -> 2xx)
	NormalizeResponseStatus bool
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Skipper:                 nil,
		Registry:                prometheus.DefaultRegisterer,
		Namespace:               "",
		Subsystem:               "",
		Buckets:                 prometheus.DefBuckets,
		NormalizeRequestPath:    true,
		NormalizeResponseStatus: true,
	}
}

// Metrics returns a middleware that collects Prometheus metrics for HTTP requests
func Metrics(config ...MetricsConfig) routeradapter.MiddlewareFunc {
	cfg := DefaultMetricsConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set defaults
	if cfg.Skipper == nil {
		cfg.Skipper = func(*routeradapter.RouterContext) bool { return false }
	}
	if cfg.Registry == nil {
		cfg.Registry = prometheus.DefaultRegisterer
	}
	if len(cfg.Buckets) == 0 {
		cfg.Buckets = prometheus.DefBuckets
	}

	// Create counter for total requests
	httpRequestsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      HTTPServerMetricsRequestsCount,
			Help:      "Number of processed HTTP requests",
		},
		[]string{
			"status",
			"method",
			"path",
		},
	)

	// Create histogram for request duration
	httpRequestsDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      HTTPServerMetricsRequestsDuration,
			Help:      "Time spent processing HTTP requests",
			Buckets:   cfg.Buckets,
		},
		[]string{
			"status",
			"method",
			"path",
		},
	)

	// Register metrics
	cfg.Registry.MustRegister(httpRequestsCounter)
	cfg.Registry.MustRegister(httpRequestsDuration)

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Skip if configured
		if cfg.Skipper(ctx) {
			return next()
		}

		// Record start time
		start := time.Now()

		// Execute next middleware/handler
		err := next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get request details
		method := ctx.Request.Method
		path := ctx.Request.URL.Path
		status := ctx.StatusCode()

		// Normalize path if configured
		if cfg.NormalizeRequestPath {
			path = normalizePath(path)
		}

		// Normalize status if configured
		statusStr := strconv.Itoa(status)
		if cfg.NormalizeResponseStatus {
			statusStr = normalizeStatus(status)
		}

		// Record metrics
		httpRequestsCounter.WithLabelValues(statusStr, method, path).Inc()
		httpRequestsDuration.WithLabelValues(statusStr, method, path).Observe(duration)

		return err
	}
}

// normalizePath normalizes dynamic paths
// For now, returns the path as-is. Can be enhanced to detect patterns like /users/:id
func normalizePath(path string) string {
	if path == "" {
		return HTTPServerMetricsNotFoundPath
	}
	return path
}

// normalizeStatus normalizes HTTP status codes to groups (2xx, 3xx, 4xx, 5xx)
func normalizeStatus(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
