package prometheus

import (
	"MgApplication/api-server/router-adapter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler returns a middleware that serves Prometheus metrics at the specified path
// This middleware exposes the default Prometheus registry at /metrics (or custom path)
func MetricsHandler(path string, registry ...*prometheus.Registry) routeradapter.MiddlewareFunc {
	// Use default registry if none provided
	var reg prometheus.Gatherer = prometheus.DefaultGatherer
	if len(registry) > 0 && registry[0] != nil {
		reg = registry[0]
	}

	// Create promhttp handler
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Only handle the metrics path
		if ctx.Request.URL.Path != path {
			return next()
		}

		// Serve metrics
		handler.ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}
}

// DefaultMetricsHandler returns a middleware that serves Prometheus metrics at /metrics
// Uses the default Prometheus registry
func DefaultMetricsHandler() routeradapter.MiddlewareFunc {
	return MetricsHandler("/metrics")
}
