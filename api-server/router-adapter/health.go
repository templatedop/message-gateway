package routeradapter

import (
	"net/http"
	"sync/atomic"
)

// HealthCheck manages health check state for the router
type HealthCheck struct {
	shuttingDown atomic.Bool
}

// NewHealthCheck creates a new health check manager
func NewHealthCheck() *HealthCheck {
	return &HealthCheck{}
}

// MarkShuttingDown marks the service as shutting down
// Health checks will return 503 after this is called
func (h *HealthCheck) MarkShuttingDown() {
	h.shuttingDown.Store(true)
}

// IsShuttingDown returns true if the service is shutting down
func (h *HealthCheck) IsShuttingDown() bool {
	return h.shuttingDown.Load()
}

// HealthzHandler returns a health check middleware that handles /healthz endpoint
// Returns 200 OK when healthy, 503 Service Unavailable when shutting down
// This middleware should be registered globally on the router
func HealthzHandler(healthCheck *HealthCheck) MiddlewareFunc {
	if healthCheck == nil {
		healthCheck = NewHealthCheck()
	}

	return func(ctx *RouterContext, next func() error) error {
		// Only handle /healthz path
		if ctx.Request.URL.Path != "/healthz" || ctx.Request.Method != "GET" {
			return next()
		}

		// Check shutdown state
		if healthCheck.IsShuttingDown() {
			return ctx.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "shutting down",
			})
		}

		// Return healthy status
		return ctx.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	}
}
