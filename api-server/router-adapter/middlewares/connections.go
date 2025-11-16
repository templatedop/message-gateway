package middlewares

import (
	"net/http"
	"sync/atomic"

	"MgApplication/api-server/router-adapter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ConnectionLimiterConfig configures the connection limiter middleware
type ConnectionLimiterConfig struct {
	// MaxConnections is the maximum number of concurrent connections allowed
	MaxConnections int64

	// Registry is the Prometheus registry for metrics (optional)
	Registry prometheus.Registerer

	// RejectMessage is the error message returned when connections are rejected
	RejectMessage string

	// RejectStatusCode is the HTTP status code for rejected connections (default: 503)
	RejectStatusCode int
}

// DefaultConnectionLimiterConfig returns default configuration
func DefaultConnectionLimiterConfig(maxConnections int64) ConnectionLimiterConfig {
	return ConnectionLimiterConfig{
		MaxConnections:   maxConnections,
		Registry:         prometheus.DefaultRegisterer,
		RejectMessage:    "Service temporarily unavailable - max connections reached",
		RejectStatusCode: http.StatusServiceUnavailable,
	}
}

// ConnectionLimiter returns a middleware that limits concurrent connections
// This middleware:
// - Tracks active connections using atomic counters
// - Rejects new connections when max limit is reached
// - Exposes Prometheus metrics for monitoring
// - Properly cleans up connection count when requests complete
func ConnectionLimiter(config ConnectionLimiterConfig) routeradapter.MiddlewareFunc {
	// Active connections counter
	var activeConnections int64

	// Initialize Prometheus metrics if registry provided
	var (
		activeConnectionsGauge     prometheus.Gauge
		rejectedConnectionsCounter prometheus.Counter
		maxConnectionsGauge        prometheus.Gauge
	)

	if config.Registry != nil {
		factory := promauto.With(config.Registry)

		activeConnectionsGauge = factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "http",
			Subsystem: "server",
			Name:      "active_connections",
			Help:      "Current number of active HTTP connections",
		})

		rejectedConnectionsCounter = factory.NewCounter(prometheus.CounterOpts{
			Namespace: "http",
			Subsystem: "server",
			Name:      "rejected_connections_total",
			Help:      "Total number of connections rejected due to max connections limit",
		})

		maxConnectionsGauge = factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "http",
			Subsystem: "server",
			Name:      "max_connections",
			Help:      "Configured maximum number of concurrent connections",
		})

		// Set max connections metric
		maxConnectionsGauge.Set(float64(config.MaxConnections))
	}

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Check if max connections reached
		current := atomic.LoadInt64(&activeConnections)
		if current >= config.MaxConnections {
			// Increment rejected connections metric
			if rejectedConnectionsCounter != nil {
				rejectedConnectionsCounter.Inc()
			}

			// Return 503 Service Unavailable
			statusCode := config.RejectStatusCode
			if statusCode == 0 {
				statusCode = http.StatusServiceUnavailable
			}

			return ctx.JSON(statusCode, map[string]interface{}{
				"error":              config.RejectMessage,
				"active_connections": current,
				"max_connections":    config.MaxConnections,
			})
		}

		// Increment active connections
		atomic.AddInt64(&activeConnections, 1)
		if activeConnectionsGauge != nil {
			activeConnectionsGauge.Inc()
		}

		// Ensure cleanup when request completes
		defer func() {
			atomic.AddInt64(&activeConnections, -1)
			if activeConnectionsGauge != nil {
				activeConnectionsGauge.Dec()
			}
		}()

		// Process request
		return next()
	}
}
