package middlewares

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/gin-gonic/gin"
)

const (
	HttpServerMetricsRequestsCount    = "http_server_requests_total"
	HttpServerMetricsRequestsDuration = "http_server_requests_duration_seconds"
	HttpServerMetricsNotFoundPath     = "/not-found"
)

type RequestMetricsMiddlewareConfig struct {
	Skipper                 func(*gin.Context) bool
	MetricsSet              *metrics.Set
	Namespace               string
	Buckets                 []float64
	Subsystem               string
	NormalizeRequestPath    bool
	NormalizeResponseStatus bool
}

var DefaultRequestMetricsMiddlewareConfig = RequestMetricsMiddlewareConfig{
	MetricsSet:              nil,
	Namespace:               "",
	Subsystem:               "",
	Buckets:                 nil,
	NormalizeRequestPath:    true,
	NormalizeResponseStatus: true,
}

var (
	metricsOnce sync.Once
	metricsSet  *metrics.Set
)

func RequestMetricsMiddleware() gin.HandlerFunc {
	return RequestMetricsMiddlewareWithConfig(DefaultRequestMetricsMiddlewareConfig)
}

func RequestMetricsMiddlewareWithConfig(config RequestMetricsMiddlewareConfig) gin.HandlerFunc {
	if config.Skipper == nil {
		config.Skipper = func(*gin.Context) bool { return false }
	}

	if config.MetricsSet == nil {
		config.MetricsSet = metrics.NewSet()
	}

	if config.Namespace == "" {
		config.Namespace = DefaultRequestMetricsMiddlewareConfig.Namespace
	}

	if config.Subsystem == "" {
		config.Subsystem = DefaultRequestMetricsMiddlewareConfig.Subsystem
	}

	// Store the metrics set for use in handler
	metricsOnce.Do(func() {
		metricsSet = config.MetricsSet
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

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := ""
		if config.NormalizeResponseStatus {
			status = StatusNormalisation(c.Writer.Status())
		} else {
			status = strconv.Itoa(c.Writer.Status())
		}

		// Create metric names with labels
		// VictoriaMetrics uses metric names with labels embedded in the name
		counterName := fmt.Sprintf(`%s{status="%s",method="%s",path="%s"}`,
			HttpServerMetricsRequestsCount, status, req.Method, path)
		summaryName := fmt.Sprintf(`%s{method="%s",path="%s"}`,
			HttpServerMetricsRequestsDuration, req.Method, path)

		// Get or create metrics
		metricsSet.GetOrCreateCounter(counterName).Inc()
		metricsSet.GetOrCreateSummary(summaryName).Update(duration)
	}
}
