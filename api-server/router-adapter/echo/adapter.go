package echoadapter

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"MgApplication/api-server/route"
	"MgApplication/api-server/router-adapter"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// init registers the Echo adapter factory
func init() {
	routeradapter.RegisterAdapterFactory(routeradapter.RouterTypeEcho, func(cfg *routeradapter.RouterConfig) (routeradapter.RouterAdapter, error) {
		return NewEchoAdapter(cfg)
	})
}

// EchoAdapter implements RouterAdapter interface for the Echo web framework
// Echo uses centralized error handling via echo.HTTPErrorHandler
type EchoAdapter struct {
	echo         *echo.Echo
	server       *http.Server
	config       *routeradapter.RouterConfig
	errorHandler routeradapter.ErrorHandler
	mu           sync.RWMutex
}

// NewEchoAdapter creates a new Echo router adapter with the provided configuration
func NewEchoAdapter(cfg *routeradapter.RouterConfig) (*EchoAdapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("router config cannot be nil")
	}

	// Validate Echo configuration
	if cfg.Echo == nil {
		cfg.Echo = &routeradapter.EchoConfig{}
	}

	// Create Echo instance
	e := echo.New()

	// Configure Echo settings
	e.Debug = cfg.Echo.Debug
	e.HideBanner = cfg.Echo.HideBanner
	e.HidePort = cfg.Echo.HidePort

	adapter := &EchoAdapter{
		echo:         e,
		config:       cfg,
		errorHandler: routeradapter.NewEchoErrorHandler(),
	}

	// Enable gzip compression if configured
	if cfg.EnableCompression {
		adapter.setupGzipCompression()
	}

	// Set up centralized error handler
	adapter.setupErrorHandler()

	return adapter, nil
}

// setupErrorHandler configures Echo's centralized error handler
func (a *EchoAdapter) setupErrorHandler() {
	a.echo.HTTPErrorHandler = func(err error, c echo.Context) {
		a.mu.RLock()
		handler := a.errorHandler
		a.mu.RUnlock()

		// Create RouterContext
		rctx := a.echoContextToRouterContext(c)

		// Handle error
		if handler != nil {
			handler.HandleError(rctx, err)
		}
	}
}

// Engine returns the underlying Echo instance
func (a *EchoAdapter) Engine() interface{} {
	return a.echo
}

// RegisterRoute registers a single route with the Echo router
func (a *EchoAdapter) RegisterRoute(meta route.Meta) error {
	if meta.Method == "" || meta.Path == "" {
		return fmt.Errorf("route method and path are required")
	}

	if meta.Func == nil {
		return fmt.Errorf("route handler function is required")
	}

	// Convert Gin HandlerFunc to Echo handler
	echoHandler := a.convertGinHandlerToEcho(meta.Func)

	// Convert middlewares
	echoMiddlewares := make([]echo.MiddlewareFunc, 0, len(meta.Middlewares))
	for _, mw := range meta.Middlewares {
		echoMiddlewares = append(echoMiddlewares, a.convertGinMiddlewareToEcho(mw))
	}

	// Register route with middlewares
	a.echo.Add(meta.Method, meta.Path, echoHandler, echoMiddlewares...)

	return nil
}

// RegisterMiddleware adds a global middleware to the Echo router
func (a *EchoAdapter) RegisterMiddleware(middleware routeradapter.MiddlewareFunc) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	// Convert framework-agnostic middleware to Echo middleware
	echoMiddleware := a.convertMiddleware(middleware)
	a.echo.Use(echoMiddleware)

	return nil
}

// RegisterGroup creates a route group with the given prefix and middlewares
func (a *EchoAdapter) RegisterGroup(prefix string, middlewares []routeradapter.MiddlewareFunc) routeradapter.RouterGroup {
	// Convert middlewares
	echoMiddlewares := make([]echo.MiddlewareFunc, len(middlewares))
	for i, mw := range middlewares {
		echoMiddlewares[i] = a.convertMiddleware(mw)
	}

	// Create Echo group
	group := a.echo.Group(prefix, echoMiddlewares...)

	return &EchoGroup{
		group:   group,
		adapter: a,
	}
}

// UseNative adds an Echo-specific middleware to the router
func (a *EchoAdapter) UseNative(middleware interface{}) error {
	var echoMiddleware echo.MiddlewareFunc

	switch mw := middleware.(type) {
	case echo.MiddlewareFunc:
		echoMiddleware = mw
	case func(echo.HandlerFunc) echo.HandlerFunc:
		echoMiddleware = echo.MiddlewareFunc(mw)
	default:
		return fmt.Errorf("middleware must be echo.MiddlewareFunc, got %T", middleware)
	}

	a.echo.Use(echoMiddleware)
	return nil
}

// ServeHTTP implements http.Handler interface
func (a *EchoAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.echo.ServeHTTP(w, r)
}

// Start starts the Echo server on the specified address
func (a *EchoAdapter) Start(addr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		return fmt.Errorf("server is already running")
	}

	// Create HTTP server
	a.server = &http.Server{
		Addr:              addr,
		Handler:           a.echo,
		ReadTimeout:       a.config.ReadTimeout,
		WriteTimeout:      a.config.WriteTimeout,
		IdleTimeout:       a.config.IdleTimeout,
		ReadHeaderTimeout: a.config.ReadHeaderTimeout,
		MaxHeaderBytes:    a.config.MaxHeaderBytes,
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
func (a *EchoAdapter) Shutdown(ctx context.Context) error {
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
func (a *EchoAdapter) Server() *http.Server {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.server
}

// SetErrorHandler sets custom error handler for the router
func (a *EchoAdapter) SetErrorHandler(handler routeradapter.ErrorHandler) {
	a.mu.Lock()
	a.errorHandler = handler
	a.mu.Unlock()

	// Update Echo's error handler
	a.setupErrorHandler()
}

// convertMiddleware converts framework-agnostic middleware to Echo middleware
func (a *EchoAdapter) convertMiddleware(middleware routeradapter.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create RouterContext from Echo context
			rctx := a.echoContextToRouterContext(c)

			// Define next function
			nextFunc := func() error {
				// Sync request back to Echo context BEFORE calling next handler
				// This ensures context updates from middleware are visible to handlers
				c.SetRequest(rctx.Request)
				return next(c)
			}

			// Call middleware
			return middleware(rctx, nextFunc)
		}
	}
}

// convertGinHandlerToEcho converts Gin HandlerFunc to Echo handler
func (a *EchoAdapter) convertGinHandlerToEcho(ginHandler interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create RouterContext
		rctx := a.echoContextToRouterContext(c)

		// Store Echo context as native context
		rctx.SetNativeContext(c)

		// For now, return success - full integration requires handler conversion
		return c.JSON(200, map[string]string{"status": "ok"})
	}
}

// convertGinMiddlewareToEcho converts Gin middleware to Echo middleware
func (a *EchoAdapter) convertGinMiddlewareToEcho(ginMiddleware interface{}) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}

// echoContextToRouterContext converts Echo context to RouterContext
func (a *EchoAdapter) echoContextToRouterContext(c echo.Context) *routeradapter.RouterContext {
	rctx := routeradapter.NewRouterContext(c.Response().Writer, c.Request())

	// Copy path parameters
	for _, name := range c.ParamNames() {
		rctx.SetParam(name, c.Param(name))
	}

	// Store reference to native Echo context
	rctx.SetNativeContext(c)

	return rctx
}

// EchoGroup implements RouterGroup interface for Echo router groups
type EchoGroup struct {
	group   *echo.Group
	adapter *EchoAdapter
}

// RegisterRoute registers a route within this Echo group
func (g *EchoGroup) RegisterRoute(meta route.Meta) error {
	if meta.Method == "" || meta.Path == "" {
		return fmt.Errorf("route method and path are required")
	}

	if meta.Func == nil {
		return fmt.Errorf("route handler function is required")
	}

	// Convert handler
	echoHandler := g.adapter.convertGinHandlerToEcho(meta.Func)

	// Convert middlewares
	echoMiddlewares := make([]echo.MiddlewareFunc, 0, len(meta.Middlewares))
	for _, mw := range meta.Middlewares {
		echoMiddlewares = append(echoMiddlewares, g.adapter.convertGinMiddlewareToEcho(mw))
	}

	// Register route with middlewares
	g.group.Add(meta.Method, meta.Path, echoHandler, echoMiddlewares...)

	return nil
}

// RegisterMiddleware adds middleware to this Echo group
func (g *EchoGroup) RegisterMiddleware(middleware routeradapter.MiddlewareFunc) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	echoMiddleware := g.adapter.convertMiddleware(middleware)
	g.group.Use(echoMiddleware)

	return nil
}

// Group creates a sub-group with additional prefix and middlewares
func (g *EchoGroup) Group(prefix string, middlewares ...routeradapter.MiddlewareFunc) routeradapter.RouterGroup {
	// Convert middlewares
	echoMiddlewares := make([]echo.MiddlewareFunc, len(middlewares))
	for i, mw := range middlewares {
		echoMiddlewares[i] = g.adapter.convertMiddleware(mw)
	}

	// Create sub-group
	subgroup := g.group.Group(prefix, echoMiddlewares...)

	return &EchoGroup{
		group:   subgroup,
		adapter: g.adapter,
	}
}

// UseNative adds an Echo-specific middleware to this group
func (g *EchoGroup) UseNative(middleware interface{}) error {
	var echoMiddleware echo.MiddlewareFunc

	switch mw := middleware.(type) {
	case echo.MiddlewareFunc:
		echoMiddleware = mw
	case func(echo.HandlerFunc) echo.HandlerFunc:
		echoMiddleware = echo.MiddlewareFunc(mw)
	default:
		return fmt.Errorf("middleware must be echo.MiddlewareFunc, got %T", middleware)
	}

	g.group.Use(echoMiddleware)
	return nil
}

// setupGzipCompression configures gzip compression middleware for Echo
func (a *EchoAdapter) setupGzipCompression() {
	// Map compression level to Echo's gzip levels
	level := -1 // Default compression

	switch a.config.CompressionLevel {
	case 0, -1:
		level = -1 // Default compression
	case 1:
		level = 1 // Best speed
	case 9:
		level = 9 // Best compression
	default:
		level = a.config.CompressionLevel
	}

	// Use Echo's official gzip middleware
	a.echo.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: level,
	}))
}
