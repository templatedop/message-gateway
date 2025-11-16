package statsviz

import (
	"MgApplication/api-server/router-adapter"

	"github.com/arl/statsviz"
)

// StatsvizHandler returns a middleware that serves statsviz at the specified path
// Statsviz is a real-time visualization tool for Go application statistics
// The handler responds to two paths:
//   - {path}/*      - Serves the statsviz HTML interface
//   - {path}/ws     - Serves the websocket endpoint for real-time updates
func StatsvizHandler(path string) (routeradapter.MiddlewareFunc, error) {
	// Create statsviz server
	srv, err := statsviz.NewServer()
	if err != nil {
		return nil, err
	}

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Check if request path matches statsviz path
		if len(ctx.Request.URL.Path) < len(path) {
			return next()
		}

		// Extract the path relative to the base path
		if ctx.Request.URL.Path[:len(path)] != path {
			return next()
		}

		// Get the sub-path after the base path
		subPath := ctx.Request.URL.Path[len(path):]

		// Handle websocket endpoint
		if subPath == "/ws" {
			srv.Ws()(ctx.Response, ctx.Request)
			return nil
		}

		// Handle index (HTML interface)
		srv.Index()(ctx.Response, ctx.Request)
		return nil
	}, nil
}

// DefaultStatsvizHandler returns a middleware that serves statsviz at /debug/statsviz
func DefaultStatsvizHandler() (routeradapter.MiddlewareFunc, error) {
	return StatsvizHandler("/debug/statsviz")
}
