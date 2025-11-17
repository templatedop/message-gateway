package ginadapter

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"MgApplication/api-server/route"
	"MgApplication/api-server/router-adapter"

	"github.com/gin-gonic/gin"
)

// init registers the Gin adapter factory
func init() {
	routeradapter.RegisterAdapterFactory(routeradapter.RouterTypeGin, func(cfg *routeradapter.RouterConfig) (routeradapter.RouterAdapter, error) {
		return NewGinAdapter(cfg)
	})
}

// GinAdapter implements RouterAdapter interface for the Gin web framework
// This adapter wraps Gin's engine and provides a framework-agnostic interface
// for route registration, middleware, and request handling.
//
// Gin uses middleware-based error handling, where errors are collected in the
// context and handled by an error handler middleware at the end of the chain.
type GinAdapter struct {
	engine       *gin.Engine
	server       *http.Server
	config       *routeradapter.RouterConfig
	errorHandler routeradapter.ErrorHandler
	ctx          context.Context // Signal-aware application context
	mu           sync.RWMutex
}

// NewGinAdapter creates a new Gin router adapter with the provided configuration
func NewGinAdapter(cfg *routeradapter.RouterConfig) (*GinAdapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("router config cannot be nil")
	}

	// Validate Gin configuration
	if cfg.Gin == nil {
		cfg.Gin = &routeradapter.GinConfig{
			Mode: "release",
		}
	}

	// Set Gin mode
	switch cfg.Gin.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "release", "":
		gin.SetMode(gin.ReleaseMode)
	default:
		return nil, fmt.Errorf("invalid gin mode: %s", cfg.Gin.Mode)
	}

	// Create Gin engine
	engine := gin.New()

	// Configure Gin settings
	if cfg.Gin.RemoveExtraSlash {
		engine.RemoveExtraSlash = true
	}

	if cfg.Gin.ForwardedByClientIP {
		engine.ForwardedByClientIP = true
	}

	// Set trusted proxies if specified
	if len(cfg.Gin.TrustedProxies) > 0 {
		if err := engine.SetTrustedProxies(cfg.Gin.TrustedProxies); err != nil {
			return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
		}
	}

	adapter := &GinAdapter{
		engine:       engine,
		config:       cfg,
		errorHandler: routeradapter.NewGinErrorHandler(),
	}

	// Enable gzip compression if configured
	if cfg.EnableCompression {
		adapter.setupGzipCompression()
	}

	return adapter, nil
}

// Engine returns the underlying Gin engine
// Cast the returned interface{} to *gin.Engine to access Gin-specific features
func (a *GinAdapter) Engine() interface{} {
	return a.engine
}

// RegisterRoute registers a single route with the Gin router
// Converts route.Meta to Gin route registration
func (a *GinAdapter) RegisterRoute(meta route.Meta) error {
	if meta.Method == "" || meta.Path == "" {
		return fmt.Errorf("route method and path are required")
	}

	if meta.Func == nil {
		return fmt.Errorf("route handler function is required")
	}

	// Gin's route.Meta already contains gin.HandlerFunc, so we can use it directly
	// Combine route middlewares with handler
	handlers := make([]gin.HandlerFunc, 0, len(meta.Middlewares)+1)
	handlers = append(handlers, meta.Middlewares...)
	handlers = append(handlers, meta.Func)

	// Register route with Gin
	a.engine.Handle(meta.Method, meta.Path, handlers...)

	return nil
}

// RegisterMiddleware adds a global middleware to the Gin router
// The middleware will be applied to all routes registered after this call
func (a *GinAdapter) RegisterMiddleware(middleware routeradapter.MiddlewareFunc) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	// Convert framework-agnostic middleware to Gin middleware
	ginMiddleware := a.convertMiddleware(middleware)
	a.engine.Use(ginMiddleware)

	return nil
}

// RegisterGroup creates a route group with the given prefix and middlewares
// Returns a GinGroup that implements RouterGroup interface
func (a *GinAdapter) RegisterGroup(prefix string, middlewares []routeradapter.MiddlewareFunc) routeradapter.RouterGroup {
	// Convert framework-agnostic middlewares to Gin middlewares
	ginMiddlewares := make([]gin.HandlerFunc, len(middlewares))
	for i, mw := range middlewares {
		ginMiddlewares[i] = a.convertMiddleware(mw)
	}

	// Create Gin router group
	group := a.engine.Group(prefix, ginMiddlewares...)

	return &GinGroup{
		group:   group,
		adapter: a,
	}
}

// UseNative adds a Gin-specific middleware to the router
// This allows using existing Gin middlewares without conversion
func (a *GinAdapter) UseNative(middleware interface{}) error {
	// Try to convert to gin.HandlerFunc
	var ginMiddleware gin.HandlerFunc

	switch mw := middleware.(type) {
	case gin.HandlerFunc:
		ginMiddleware = mw
	case func(*gin.Context):
		ginMiddleware = gin.HandlerFunc(mw)
	default:
		return fmt.Errorf("middleware must be gin.HandlerFunc or func(*gin.Context), got %T", middleware)
	}

	a.engine.Use(ginMiddleware)
	return nil
}

// ServeHTTP implements http.Handler interface
// Delegates to Gin's ServeHTTP implementation
func (a *GinAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.engine.ServeHTTP(w, r)
}

// Start starts the HTTP server on the specified address
func (a *GinAdapter) Start(addr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		return fmt.Errorf("server is already running")
	}

	// Create HTTP server
	a.server = &http.Server{
		Addr:              addr,
		Handler:           a.engine,
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

	// Start server in a goroutine
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
// Waits for existing connections to finish within the context deadline
func (a *GinAdapter) Shutdown(ctx context.Context) error {
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
func (a *GinAdapter) Server() *http.Server {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.server
}

// SetContext sets the signal-aware context for the router
// This context will be propagated to all HTTP handlers via BaseContext
func (a *GinAdapter) SetContext(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
}

// SetErrorHandler sets custom error handler for the router
// For Gin, this adds an error handler middleware
func (a *GinAdapter) SetErrorHandler(handler routeradapter.ErrorHandler) {
	a.mu.Lock()
	a.errorHandler = handler
	a.mu.Unlock()

	// Add error handler middleware to Gin
	// This middleware runs after all other middlewares and handlers
	// It checks for errors in the context and handles them
	a.engine.Use(func(c *gin.Context) {
		c.Next() // Execute all other middlewares and handler

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error (most recent)
			err := c.Errors.Last().Err

			// Create RouterContext for error handler
			rctx := routeradapter.NewRouterContext(c.Writer, c.Request)
			rctx.SetNativeContext(c)

			// Handle error
			a.mu.RLock()
			errorHandler := a.errorHandler
			a.mu.RUnlock()

			if errorHandler != nil {
				errorHandler.HandleError(rctx, err)
			}
		}
	})
}

// SetNoRouteHandler sets the handler for 404 Not Found responses
// Called when no route matches the request
func (a *GinAdapter) SetNoRouteHandler(handler routeradapter.HandlerFunc) {
	a.engine.NoRoute(func(c *gin.Context) {
		// Create RouterContext for handler
		rctx := a.ginContextToRouterContext(c)

		// Call the custom handler
		if err := handler(rctx); err != nil {
			// If handler returns error, use error handler
			a.mu.RLock()
			errorHandler := a.errorHandler
			a.mu.RUnlock()

			if errorHandler != nil {
				errorHandler.HandleError(rctx, err)
			}
		}
	})
}

// SetNoMethodHandler sets the handler for 405 Method Not Allowed responses
// Called when a route exists but doesn't support the HTTP method
func (a *GinAdapter) SetNoMethodHandler(handler routeradapter.HandlerFunc) {
	a.engine.NoMethod(func(c *gin.Context) {
		// Create RouterContext for handler
		rctx := a.ginContextToRouterContext(c)

		// Call the custom handler
		if err := handler(rctx); err != nil {
			// If handler returns error, use error handler
			a.mu.RLock()
			errorHandler := a.errorHandler
			a.mu.RUnlock()

			if errorHandler != nil {
				errorHandler.HandleError(rctx, err)
			}
		}
	})
}

// convertMiddleware converts framework-agnostic middleware to Gin middleware
func (a *GinAdapter) convertMiddleware(middleware routeradapter.MiddlewareFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create RouterContext from Gin context
		rctx := a.ginContextToRouterContext(c)

		// Define next function
		nextCalled := false
		next := func() error {
			nextCalled = true
			// Sync request back to Gin context BEFORE calling next handler
			// This ensures context updates from middleware are visible to handlers
			c.Request = rctx.Request
			c.Next()
			return nil
		}

		// Call middleware
		if err := middleware(rctx, next); err != nil {
			// Add error to Gin context
			c.Error(err)
			c.Abort()
			return
		}

		// If next wasn't called, call it now (and sync request)
		if !nextCalled {
			c.Request = rctx.Request
			c.Next()
		}
	}
}

// ginContextToRouterContext converts Gin context to RouterContext
func (a *GinAdapter) ginContextToRouterContext(c *gin.Context) *routeradapter.RouterContext {
	rctx := routeradapter.NewRouterContext(c.Writer, c.Request)

	// Copy path parameters
	for _, param := range c.Params {
		rctx.SetParam(param.Key, param.Value)
	}

	// Store reference to native Gin context
	rctx.SetNativeContext(c)

	return rctx
}

// GinGroup implements RouterGroup interface for Gin router groups
type GinGroup struct {
	group   *gin.RouterGroup
	adapter *GinAdapter
}

// RegisterRoute registers a route within this Gin group
func (g *GinGroup) RegisterRoute(meta route.Meta) error {
	if meta.Method == "" || meta.Path == "" {
		return fmt.Errorf("route method and path are required")
	}

	if meta.Func == nil {
		return fmt.Errorf("route handler function is required")
	}

	// Combine route middlewares with handler
	handlers := make([]gin.HandlerFunc, 0, len(meta.Middlewares)+1)
	handlers = append(handlers, meta.Middlewares...)
	handlers = append(handlers, meta.Func)

	// Register route with Gin group
	g.group.Handle(meta.Method, meta.Path, handlers...)

	return nil
}

// RegisterMiddleware adds middleware to this Gin group
func (g *GinGroup) RegisterMiddleware(middleware routeradapter.MiddlewareFunc) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	// Convert and add middleware to group
	ginMiddleware := g.adapter.convertMiddleware(middleware)
	g.group.Use(ginMiddleware)

	return nil
}

// Group creates a sub-group with additional prefix and middlewares
func (g *GinGroup) Group(prefix string, middlewares ...routeradapter.MiddlewareFunc) routeradapter.RouterGroup {
	// Convert middlewares
	ginMiddlewares := make([]gin.HandlerFunc, len(middlewares))
	for i, mw := range middlewares {
		ginMiddlewares[i] = g.adapter.convertMiddleware(mw)
	}

	// Create sub-group
	subgroup := g.group.Group(prefix, ginMiddlewares...)

	return &GinGroup{
		group:   subgroup,
		adapter: g.adapter,
	}
}

// UseNative adds a Gin-specific middleware to this group
func (g *GinGroup) UseNative(middleware interface{}) error {
	// Try to convert to gin.HandlerFunc
	var ginMiddleware gin.HandlerFunc

	switch mw := middleware.(type) {
	case gin.HandlerFunc:
		ginMiddleware = mw
	case func(*gin.Context):
		ginMiddleware = gin.HandlerFunc(mw)
	default:
		return fmt.Errorf("middleware must be gin.HandlerFunc or func(*gin.Context), got %T", middleware)
	}

	g.group.Use(ginMiddleware)
	return nil
}

// setupGzipCompression configures gzip compression middleware for Gin
func (a *GinAdapter) setupGzipCompression() {
	level := a.config.CompressionLevel
	if level == 0 {
		level = -1 // gzip.DefaultCompression
	}

	// Use Gin's native gzip middleware if available, otherwise use framework-agnostic
	gzipMiddleware := func(c *gin.Context) {
		// Check if client accepts gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Create gzip writer
		gz, err := gzip.NewWriterLevel(c.Writer, level)
		if err != nil {
			c.Next()
			return
		}
		defer gz.Close()

		// Set response headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Wrap response writer
		c.Writer = &gzipWriter{ResponseWriter: c.Writer, Writer: gz}

		c.Next()
	}

	a.engine.Use(gzipMiddleware)
}

// gzipWriter wraps gin.ResponseWriter for gzip compression
type gzipWriter struct {
	gin.ResponseWriter
	Writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.Writer.Write([]byte(s))
}
