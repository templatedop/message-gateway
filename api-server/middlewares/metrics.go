package middlewares

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	HttpServerMetricsRequestsCount    = "http_server_requests_total"
	HttpServerMetricsRequestsDuration = "http_server_requests_duration_seconds"
	HttpServerMetricsNotFoundPath     = "/not-found"
)

type RequestMetricsMiddlewareConfig struct {
	Skipper                 func(*gin.Context) bool
	Registry                prometheus.Registerer
	Namespace               string
	Buckets                 []float64
	Subsystem               string
	NormalizeRequestPath    bool
	NormalizeResponseStatus bool
}

var DefaultRequestMetricsMiddlewareConfig = RequestMetricsMiddlewareConfig{
	Registry:                prometheus.DefaultRegisterer,
	Namespace:               "",
	Subsystem:               "",
	Buckets:                 prometheus.DefBuckets,
	NormalizeRequestPath:    true,
	NormalizeResponseStatus: true,
}

var (
	metricsOnce          sync.Once
	httpRequestsCounter  *prometheus.CounterVec
	httpRequestsDuration *prometheus.HistogramVec
)

func RequestMetricsMiddleware() gin.HandlerFunc {
	return RequestMetricsMiddlewareWithConfig(DefaultRequestMetricsMiddlewareConfig)
}

func RequestMetricsMiddlewareWithConfig(config RequestMetricsMiddlewareConfig) gin.HandlerFunc {
	if config.Skipper == nil {
		config.Skipper = func(*gin.Context) bool { return false }
	}

	if config.Registry == nil {
		config.Registry = DefaultRequestMetricsMiddlewareConfig.Registry
	}

	if config.Namespace == "" {
		config.Namespace = DefaultRequestMetricsMiddlewareConfig.Namespace
	}

	if config.Subsystem == "" {
		config.Subsystem = DefaultRequestMetricsMiddlewareConfig.Subsystem
	}

	if len(config.Buckets) == 0 {
		config.Buckets = DefaultRequestMetricsMiddlewareConfig.Buckets
	}

	// Register metrics only once using sync.Once to avoid panic on multiple middleware usage
	metricsOnce.Do(func() {
		httpRequestsCounter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      HttpServerMetricsRequestsCount,
				Help:      "Number of processed HTTP requests",
			},
			[]string{
				"status",
				"method",
				"path",
			},
		)

		httpRequestsDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      HttpServerMetricsRequestsDuration,
				Help:      "Time spent processing HTTP requests",
				Buckets:   config.Buckets,
			},
			[]string{
				"method",
				"path",
			},
		)

		config.Registry.MustRegister(httpRequestsCounter, httpRequestsDuration)
	})

	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}

		req := c.Request

		var path string
		if config.NormalizeRequestPath {
			path = c.FullPath()
			if path == "" {
				path = HttpServerMetricsNotFoundPath
			}
		} else {
			path = req.URL.Path
			if req.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, req.URL.RawQuery)
			}
		}

		timer := prometheus.NewTimer(httpRequestsDuration.WithLabelValues(req.Method, path))
		c.Next()
		timer.ObserveDuration()

		status := ""
		if config.NormalizeResponseStatus {
			status = StatusNormalisation(c.Writer.Status())
		} else {
			status = strconv.Itoa(c.Writer.Status())
		}

		httpRequestsCounter.WithLabelValues(status, req.Method, path).Inc()
	}
}
