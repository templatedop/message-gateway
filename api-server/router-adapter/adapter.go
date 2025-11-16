package routeradapter

import (
	"context"
	"net/http"

	"MgApplication/api-server/route"
)

// RouterAdapter is the main interface that all framework adapters must implement.
// This abstraction allows switching between different web frameworks (Gin, Fiber, Echo, net/http)
// while maintaining the same route registration and middleware system.
type RouterAdapter interface {
	// Engine returns the underlying framework instance (e.g., *gin.Engine, *fiber.App, *echo.Echo)
	// Returns interface{} because each framework has different types
	Engine() interface{}

	// RegisterRoute registers a single route with the router
	// The route.Meta contains all information needed: method, path, handler, etc.
	RegisterRoute(meta route.Meta) error

	// RegisterMiddleware adds a global middleware to the router
	// This middleware will be applied to all routes
	RegisterMiddleware(middleware MiddlewareFunc) error

	// RegisterGroup creates a route group with the given prefix and middlewares
	// Returns a RouterGroup that can register routes with the prefix
	RegisterGroup(prefix string, middlewares []MiddlewareFunc) RouterGroup

	// Use adds middleware using framework-specific middleware type
	// This allows using existing framework middlewares without conversion
	UseNative(middleware interface{}) error

	// ServeHTTP implements http.Handler interface
	// Allows the router to be used with standard net/http servers
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	// Start starts the HTTP server on the specified address
	// Returns error if server fails to start
	Start(addr string) error

	// Shutdown gracefully shuts down the server
	// Waits for existing connections to finish within the context deadline
	Shutdown(ctx context.Context) error

	// Server returns the underlying *http.Server instance
	// Useful for accessing server configuration
	Server() *http.Server

	// SetErrorHandler sets custom error handler for the router
	// Different frameworks handle errors differently (middleware vs centralized)
	SetErrorHandler(handler ErrorHandler)

	// SetNoRouteHandler sets the handler for 404 Not Found responses
	// Called when no route matches the request
	SetNoRouteHandler(handler HandlerFunc)

	// SetNoMethodHandler sets the handler for 405 Method Not Allowed responses
	// Called when a route exists but doesn't support the HTTP method
	SetNoMethodHandler(handler HandlerFunc)
}

// RouterGroup represents a group of routes with a common prefix and middlewares
// This allows organizing routes hierarchically (e.g., /api/v1/users, /api/v1/posts)
type RouterGroup interface {
	// RegisterRoute registers a route within this group
	// The route path will be prefixed with the group's prefix
	RegisterRoute(meta route.Meta) error

	// RegisterMiddleware adds middleware specific to this group
	// These middlewares are applied to all routes in the group
	RegisterMiddleware(middleware MiddlewareFunc) error

	// Group creates a sub-group with additional prefix and middlewares
	// Allows nested route groups (e.g., /api/v1/admin/users)
	Group(prefix string, middlewares ...MiddlewareFunc) RouterGroup

	// UseNative adds framework-specific middleware to this group
	UseNative(middleware interface{}) error
}

// MiddlewareFunc is a framework-agnostic middleware function
// It receives a RouterContext and a next function to call the next middleware/handler
// Returns error that will be handled by the framework's error handler
type MiddlewareFunc func(ctx *RouterContext, next func() error) error

// HandlerFunc is a framework-agnostic HTTP handler function
// Similar to http.HandlerFunc but uses RouterContext for framework independence
type HandlerFunc func(ctx *RouterContext) error

// ErrorHandler handles errors in a framework-agnostic way
// Different frameworks have different error handling strategies:
// - Gin: Middleware-based (errors collected in context, handled by middleware)
// - Fiber/Echo: Centralized error handler
// - net/http: Manual error handling in each handler
type ErrorHandler interface {
	// HandleError processes an error and sends appropriate response
	// The implementation varies by framework but the interface is the same
	HandleError(ctx *RouterContext, err error)
}
