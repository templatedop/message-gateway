package nethttpadapter

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"MgApplication/api-server/route"
	"MgApplication/api-server/router-adapter"
)

// init registers the net/http adapter factory
func init() {
	routeradapter.RegisterAdapterFactory(routeradapter.RouterTypeNetHTTP, func(cfg *routeradapter.RouterConfig) (routeradapter.RouterAdapter, error) {
		return NewNetHTTPAdapter(cfg)
	})
}

// NetHTTPAdapter implements RouterAdapter interface for standard library net/http
// This adapter implements its own routing with path parameter support
type NetHTTPAdapter struct {
	mux              *http.ServeMux
	router           *Router
	server           *http.Server
	config           *routeradapter.RouterConfig
	errorHandler     routeradapter.ErrorHandler
	noRouteHandler   routeradapter.HandlerFunc
	noMethodHandler  routeradapter.HandlerFunc
	middlewares      []routeradapter.MiddlewareFunc
	ctx              context.Context // Signal-aware application context
	mu               sync.RWMutex
}

// NewNetHTTPAdapter creates a new net/http router adapter
func NewNetHTTPAdapter(cfg *routeradapter.RouterConfig) (*NetHTTPAdapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("router config cannot be nil")
	}

	// Validate net/http configuration
	if cfg.NetHTTP == nil {
		cfg.NetHTTP = &routeradapter.NetHTTPConfig{}
	}

	adapter := &NetHTTPAdapter{
		mux:          http.NewServeMux(),
		router:       NewRouter(),
		config:       cfg,
		errorHandler: routeradapter.NewNetHTTPErrorHandler(),
		middlewares:  make([]routeradapter.MiddlewareFunc, 0),
	}

	// Enable gzip compression if configured
	if cfg.EnableCompression {
		adapter.setupGzipCompression()
	}

	return adapter, nil
}

// Engine returns the underlying http.ServeMux
func (a *NetHTTPAdapter) Engine() interface{} {
	return a.mux
}

// RegisterRoute registers a single route with the router
func (a *NetHTTPAdapter) RegisterRoute(meta route.Meta) error {
	if meta.Method == "" || meta.Path == "" {
		return fmt.Errorf("route method and path are required")
	}

	if meta.Func == nil {
		return fmt.Errorf("route handler function is required")
	}

	// Register route with custom router
	a.router.AddRoute(meta.Method, meta.Path, func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		// Create RouterContext
		rctx := routeradapter.NewRouterContext(w, r)

		// Add path parameters
		for k, v := range params {
			rctx.SetParam(k, v)
		}

		// Apply middlewares (route.Meta middlewares are Gin-specific, skip them for net/http)
		// In a full implementation, we'd convert or have separate handler types
		handler := a.wrapWithMiddlewares(func(ctx *routeradapter.RouterContext) error {
			// We can't directly call Gin handler, return success for now
			return ctx.JSON(200, map[string]string{"status": "ok"})
		})

		// Execute handler
		if err := handler(rctx); err != nil {
			a.mu.RLock()
			errorHandler := a.errorHandler
			a.mu.RUnlock()

			if errorHandler != nil {
				errorHandler.HandleError(rctx, err)
			}
		}
	})

	return nil
}

// RegisterMiddleware adds a global middleware to the router
func (a *NetHTTPAdapter) RegisterMiddleware(middleware routeradapter.MiddlewareFunc) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	a.mu.Lock()
	a.middlewares = append(a.middlewares, middleware)
	a.mu.Unlock()

	return nil
}

// RegisterGroup creates a route group with the given prefix and middlewares
func (a *NetHTTPAdapter) RegisterGroup(prefix string, middlewares []routeradapter.MiddlewareFunc) routeradapter.RouterGroup {
	return &NetHTTPGroup{
		prefix:      prefix,
		adapter:     a,
		middlewares: middlewares,
	}
}

// UseNative adds a native http.Handler middleware
func (a *NetHTTPAdapter) UseNative(middleware interface{}) error {
	switch mw := middleware.(type) {
	case func(http.Handler) http.Handler:
		// Convert to framework-agnostic middleware
		converted := func(ctx *routeradapter.RouterContext, next func() error) error {
			// Wrap next as http.Handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = next()
			})

			// Apply middleware
			wrappedHandler := mw(nextHandler)
			wrappedHandler.ServeHTTP(ctx.Response, ctx.Request)
			return nil
		}
		return a.RegisterMiddleware(converted)
	case http.Handler:
		// Can't add http.Handler as middleware directly
		return fmt.Errorf("http.Handler cannot be used as middleware, use func(http.Handler) http.Handler")
	default:
		return fmt.Errorf("middleware must be func(http.Handler) http.Handler, got %T", middleware)
	}
}

// ServeHTTP implements http.Handler interface
func (a *NetHTTPAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try to match route
	handler, params := a.router.Match(r.Method, r.URL.Path)
	if handler != nil {
		handler(w, r, params)
		return
	}

	// No route found for this method, check if route exists for other methods
	methodMismatch := a.router.PathExists(r.URL.Path)

	if methodMismatch {
		// Route exists but method not allowed (405)
		a.mu.RLock()
		noMethodHandler := a.noMethodHandler
		a.mu.RUnlock()

		if noMethodHandler != nil {
			rctx := routeradapter.NewRouterContext(w, r)
			if err := noMethodHandler(rctx); err != nil {
				a.mu.RLock()
				errorHandler := a.errorHandler
				a.mu.RUnlock()

				if errorHandler != nil {
					errorHandler.HandleError(rctx, err)
				}
			}
		} else {
			// Default 405 response
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// No route found at all (404)
	a.mu.RLock()
	noRouteHandler := a.noRouteHandler
	a.mu.RUnlock()

	if noRouteHandler != nil {
		rctx := routeradapter.NewRouterContext(w, r)
		if err := noRouteHandler(rctx); err != nil {
			a.mu.RLock()
			errorHandler := a.errorHandler
			a.mu.RUnlock()

			if errorHandler != nil {
				errorHandler.HandleError(rctx, err)
			}
		}
	} else {
		// Default 404 response
		http.NotFound(w, r)
	}
}

// Start starts the HTTP server
func (a *NetHTTPAdapter) Start(addr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		return fmt.Errorf("server is already running")
	}

	// Create HTTP server
	a.server = &http.Server{
		Addr:              addr,
		Handler:           a,
		ReadTimeout:       a.config.ReadTimeout,
		WriteTimeout:      a.config.WriteTimeout,
		IdleTimeout:       a.config.IdleTimeout,
		ReadHeaderTimeout: a.config.ReadHeaderTimeout,
		MaxHeaderBytes:    a.config.MaxHeaderBytes,
		// BaseContext provides the signal-aware context to all HTTP handlers
		// This allows handlers to detect shutdown signals via req.Context()
		BaseContext: func(net.Listener) context.Context {
			if a.ctx != nil {
				return a.ctx
			}
			return context.Background()
		},
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Check for immediate startup errors
	select {
	case err := <-errChan:
		a.server = nil
		return fmt.Errorf("failed to start server: %w", err)
	default:
		return nil
	}
}

// Shutdown gracefully shuts down the server
func (a *NetHTTPAdapter) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	server := a.server
	a.mu.Unlock()

	if server == nil {
		return fmt.Errorf("server is not running")
	}

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	a.mu.Lock()
	a.server = nil
	a.mu.Unlock()

	return nil
}

// Server returns the underlying *http.Server instance
func (a *NetHTTPAdapter) Server() *http.Server {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.server
}

// SetContext sets the signal-aware context for the router
// This context will be propagated to all HTTP handlers via BaseContext
func (a *NetHTTPAdapter) SetContext(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
}

// SetErrorHandler sets custom error handler
func (a *NetHTTPAdapter) SetErrorHandler(handler routeradapter.ErrorHandler) {
	a.mu.Lock()
	a.errorHandler = handler
	a.mu.Unlock()
}

// SetNoRouteHandler sets the handler for 404 Not Found responses
// Called when no route matches the request
func (a *NetHTTPAdapter) SetNoRouteHandler(handler routeradapter.HandlerFunc) {
	a.mu.Lock()
	a.noRouteHandler = handler
	a.mu.Unlock()
}

// SetNoMethodHandler sets the handler for 405 Method Not Allowed responses
// Called when a route exists but doesn't support the HTTP method
func (a *NetHTTPAdapter) SetNoMethodHandler(handler routeradapter.HandlerFunc) {
	a.mu.Lock()
	a.noMethodHandler = handler
	a.mu.Unlock()
}

// wrapWithMiddlewares wraps a handler with middlewares
func (a *NetHTTPAdapter) wrapWithMiddlewares(handler routeradapter.HandlerFunc) routeradapter.HandlerFunc {
	// Apply global middlewares
	a.mu.RLock()
	middlewares := make([]routeradapter.MiddlewareFunc, len(a.middlewares))
	copy(middlewares, a.middlewares)
	a.mu.RUnlock()

	// Create final handler
	finalHandler := handler

	// Wrap with middlewares in reverse order (inner to outer)
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		currentHandler := finalHandler
		finalHandler = func(ctx *routeradapter.RouterContext) error {
			return mw(ctx, func() error {
				return currentHandler(ctx)
			})
		}
	}

	return finalHandler
}

// NetHTTPGroup implements RouterGroup interface
type NetHTTPGroup struct {
	prefix      string
	adapter     *NetHTTPAdapter
	middlewares []routeradapter.MiddlewareFunc
}

// RegisterRoute registers a route in the group
func (g *NetHTTPGroup) RegisterRoute(meta route.Meta) error {
	// Prepend group prefix to path
	meta.Path = g.prefix + meta.Path

	// Add group middlewares to route middlewares
	// Note: This is simplified - full implementation would merge middlewares properly

	return g.adapter.RegisterRoute(meta)
}

// RegisterMiddleware adds middleware to the group
func (g *NetHTTPGroup) RegisterMiddleware(middleware routeradapter.MiddlewareFunc) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	g.middlewares = append(g.middlewares, middleware)
	return nil
}

// Group creates a sub-group
func (g *NetHTTPGroup) Group(prefix string, middlewares ...routeradapter.MiddlewareFunc) routeradapter.RouterGroup {
	return &NetHTTPGroup{
		prefix:      g.prefix + prefix,
		adapter:     g.adapter,
		middlewares: append(g.middlewares, middlewares...),
	}
}

// UseNative adds native middleware to the group
func (g *NetHTTPGroup) UseNative(middleware interface{}) error {
	return fmt.Errorf("native middleware not supported for net/http groups")
}

// Router is a simple HTTP router with path parameter support
type Router struct {
	routes map[string][]*RoutePattern
	mu     sync.RWMutex
}

// RoutePattern represents a route pattern with parameter extraction
type RoutePattern struct {
	Pattern *regexp.Regexp
	Params  []string
	Handler func(http.ResponseWriter, *http.Request, map[string]string)
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make(map[string][]*RoutePattern),
	}
}

// AddRoute adds a route with path parameters
func (r *Router) AddRoute(method, path string, handler func(http.ResponseWriter, *http.Request, map[string]string)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert path pattern to regex
	pattern, params := pathToRegex(path)

	routePattern := &RoutePattern{
		Pattern: pattern,
		Params:  params,
		Handler: handler,
	}

	if r.routes[method] == nil {
		r.routes[method] = make([]*RoutePattern, 0)
	}

	r.routes[method] = append(r.routes[method], routePattern)
}

// Match matches a request to a route and returns handler and parameters
func (r *Router) Match(method, path string) (func(http.ResponseWriter, *http.Request, map[string]string), map[string]string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes, exists := r.routes[method]
	if !exists {
		return nil, nil
	}

	for _, route := range routes {
		if matches := route.Pattern.FindStringSubmatch(path); matches != nil {
			params := make(map[string]string)
			for i, name := range route.Params {
				if i+1 < len(matches) {
					params[name] = matches[i+1]
				}
			}
			return route.Handler, params
		}
	}

	return nil, nil
}

// PathExists checks if a route exists for the given path (regardless of method)
// Used to distinguish between 404 (path not found) and 405 (method not allowed)
func (r *Router) PathExists(path string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check all methods to see if any route matches this path
	for _, routes := range r.routes {
		for _, route := range routes {
			if route.Pattern.MatchString(path) {
				return true
			}
		}
	}

	return false
}

// pathToRegex converts a path pattern like "/users/:id" to a regex
func pathToRegex(path string) (*regexp.Regexp, []string) {
	params := make([]string, 0)
	pattern := "^"

	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, ":") {
			// Path parameter
			paramName := strings.TrimPrefix(part, ":")
			params = append(params, paramName)
			pattern += "/([^/]+)"
		} else {
			// Static part
			pattern += "/" + regexp.QuoteMeta(part)
		}
	}

	pattern += "$"

	regex := regexp.MustCompile(pattern)
	return regex, params
}

// setupGzipCompression configures gzip compression middleware for net/http
func (a *NetHTTPAdapter) setupGzipCompression() {
	// Use the framework-agnostic gzip middleware
	gzipMiddleware := routeradapter.GzipMiddleware(a.config.CompressionLevel)
	_ = a.RegisterMiddleware(gzipMiddleware)
}
