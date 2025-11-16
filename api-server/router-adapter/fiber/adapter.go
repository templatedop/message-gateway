package fiberadapter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"MgApplication/api-server/route"
	"MgApplication/api-server/router-adapter"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

// init registers the Fiber adapter factory
func init() {
	routeradapter.RegisterAdapterFactory(routeradapter.RouterTypeFiber, func(cfg *routeradapter.RouterConfig) (routeradapter.RouterAdapter, error) {
		return NewFiberAdapter(cfg)
	})
}

// FiberAdapter implements RouterAdapter interface for the Fiber web framework
// Fiber uses centralized error handling via app.config.ErrorHandler
type FiberAdapter struct {
	app          *fiber.App
	config       *routeradapter.RouterConfig
	errorHandler routeradapter.ErrorHandler
	mu           sync.RWMutex
}

// NewFiberAdapter creates a new Fiber router adapter with the provided configuration
func NewFiberAdapter(cfg *routeradapter.RouterConfig) (*FiberAdapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("router config cannot be nil")
	}

	// Validate Fiber configuration
	if cfg.Fiber == nil {
		cfg.Fiber = &routeradapter.FiberConfig{}
	}

	// Create Fiber app with configuration
	fiberCfg := fiber.Config{
		Prefork:          cfg.Fiber.Prefork,
		ServerHeader:     cfg.Fiber.ServerHeader,
		StrictRouting:    cfg.Fiber.StrictRouting,
		CaseSensitive:    cfg.Fiber.CaseSensitive,
		ETag:             cfg.Fiber.ETag,
		BodyLimit:        cfg.Fiber.BodyLimit,
		Concurrency:      cfg.Fiber.Concurrency,
		DisableKeepalive: cfg.Fiber.DisableKeepalive,
		ReadTimeout:      cfg.ReadTimeout,
		WriteTimeout:     cfg.WriteTimeout,
		IdleTimeout:      cfg.IdleTimeout,
	}

	app := fiber.New(fiberCfg)

	adapter := &FiberAdapter{
		app:          app,
		config:       cfg,
		errorHandler: routeradapter.NewFiberErrorHandler(),
	}

	// Enable gzip compression if configured
	if cfg.EnableCompression {
		adapter.setupGzipCompression()
	}

	// Set up centralized error handler
	adapter.setupErrorHandler()

	return adapter, nil
}

// setupErrorHandler configures Fiber's centralized error handler
func (a *FiberAdapter) setupErrorHandler() {
	a.app.Use(func(c *fiber.Ctx) error {
		// Continue to next handler
		err := c.Next()

		// If there's an error, handle it
		if err != nil {
			a.mu.RLock()
			handler := a.errorHandler
			a.mu.RUnlock()

			// Create RouterContext
			rctx := a.fiberContextToRouterContext(c)

			// Handle error
			if handler != nil {
				handler.HandleError(rctx, err)
			}
		}

		return nil
	})
}

// Engine returns the underlying Fiber app
func (a *FiberAdapter) Engine() interface{} {
	return a.app
}

// RegisterRoute registers a single route with the Fiber router
func (a *FiberAdapter) RegisterRoute(meta route.Meta) error {
	if meta.Method == "" || meta.Path == "" {
		return fmt.Errorf("route method and path are required")
	}

	if meta.Func == nil {
		return fmt.Errorf("route handler function is required")
	}

	// Convert Gin HandlerFunc to Fiber handler
	fiberHandler := a.convertGinHandlerToFiber(meta.Func)

	// Convert middlewares
	fiberMiddlewares := make([]fiber.Handler, 0, len(meta.Middlewares))
	for _, mw := range meta.Middlewares {
		fiberMiddlewares = append(fiberMiddlewares, a.convertGinHandlerToFiber(mw))
	}

	// Combine middlewares and handler
	handlers := append(fiberMiddlewares, fiberHandler)

	// Register route based on method
	switch meta.Method {
	case http.MethodGet:
		a.app.Get(meta.Path, handlers...)
	case http.MethodPost:
		a.app.Post(meta.Path, handlers...)
	case http.MethodPut:
		a.app.Put(meta.Path, handlers...)
	case http.MethodPatch:
		a.app.Patch(meta.Path, handlers...)
	case http.MethodDelete:
		a.app.Delete(meta.Path, handlers...)
	case http.MethodHead:
		a.app.Head(meta.Path, handlers...)
	case http.MethodOptions:
		a.app.Options(meta.Path, handlers...)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", meta.Method)
	}

	return nil
}

// RegisterMiddleware adds a global middleware to the Fiber router
func (a *FiberAdapter) RegisterMiddleware(middleware routeradapter.MiddlewareFunc) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	// Convert framework-agnostic middleware to Fiber middleware
	fiberMiddleware := a.convertMiddleware(middleware)
	a.app.Use(fiberMiddleware)

	return nil
}

// RegisterGroup creates a route group with the given prefix and middlewares
func (a *FiberAdapter) RegisterGroup(prefix string, middlewares []routeradapter.MiddlewareFunc) routeradapter.RouterGroup {
	// Convert middlewares
	fiberMiddlewares := make([]fiber.Handler, len(middlewares))
	for i, mw := range middlewares {
		fiberMiddlewares[i] = a.convertMiddleware(mw)
	}

	// Create Fiber router with prefix
	group := a.app.Group(prefix, fiberMiddlewares...)

	return &FiberGroup{
		group:   group,
		adapter: a,
	}
}

// UseNative adds a Fiber-specific middleware to the router
func (a *FiberAdapter) UseNative(middleware interface{}) error {
	switch mw := middleware.(type) {
	case func(*fiber.Ctx) error:
		a.app.Use(mw)
		return nil
	default:
		return fmt.Errorf("middleware must be func(*fiber.Ctx) error, got %T", middleware)
	}
}

// ServeHTTP implements http.Handler interface
func (a *FiberAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Fiber doesn't directly implement http.Handler
	// We need to adapt the request
	// For now, return an error - Fiber should use its own Listen method
	http.Error(w, "Fiber adapter must use Start() method, not ServeHTTP", http.StatusInternalServerError)
}

// Start starts the Fiber server on the specified address
func (a *FiberAdapter) Start(addr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Start server in goroutine
	go func() {
		if err := a.app.Listen(addr); err != nil {
			fmt.Printf("Fiber server error: %v\n", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Shutdown gracefully shuts down the server
func (a *FiberAdapter) Shutdown(ctx context.Context) error {
	return a.app.ShutdownWithContext(ctx)
}

// Server returns the underlying *http.Server instance
// Note: Fiber manages its own server, so this returns nil
func (a *FiberAdapter) Server() *http.Server {
	return nil // Fiber manages its own server internally
}

// SetErrorHandler sets custom error handler for the router
func (a *FiberAdapter) SetErrorHandler(handler routeradapter.ErrorHandler) {
	a.mu.Lock()
	a.errorHandler = handler
	a.mu.Unlock()
}

// SetNoRouteHandler sets the handler for 404 Not Found responses
// Called when no route matches the request
func (a *FiberAdapter) SetNoRouteHandler(handler routeradapter.HandlerFunc) {
	// Fiber uses app.Use() to catch all unmatched routes
	// This must be registered last, after all other routes
	a.app.Use(func(c *fiber.Ctx) error {
		// Check if a route was already matched
		// If we're here and the route wasn't matched, it's a 404
		if c.Route() == nil {
			// Create RouterContext
			rctx := a.fiberContextToRouterContext(c)

			// Call custom handler
			if err := handler(rctx); err != nil {
				// If handler returns error, use error handler
				a.mu.RLock()
				errorHandler := a.errorHandler
				a.mu.RUnlock()

				if errorHandler != nil {
					errorHandler.HandleError(rctx, err)
				}
				return err
			}
			return nil
		}

		return c.Next()
	})
}

// SetNoMethodHandler sets the handler for 405 Method Not Allowed responses
// Called when a route exists but doesn't support the HTTP method
func (a *FiberAdapter) SetNoMethodHandler(handler routeradapter.HandlerFunc) {
	// Fiber doesn't have a built-in way to distinguish between 404 and 405
	// We can use custom middleware to detect method mismatches
	a.app.Use(func(c *fiber.Ctx) error {
		// First check if route exists
		route := c.Route()
		if route == nil {
			// No route found, this will be handled by NoRoute
			return c.Next()
		}

		// Check if method matches
		// Fiber's route.Method contains the allowed method
		if route.Method != c.Method() {
			// Method not allowed
			rctx := a.fiberContextToRouterContext(c)

			// Call custom handler
			if err := handler(rctx); err != nil {
				// If handler returns error, use error handler
				a.mu.RLock()
				errorHandler := a.errorHandler
				a.mu.RUnlock()

				if errorHandler != nil {
					errorHandler.HandleError(rctx, err)
				}
				return err
			}
			return nil
		}

		return c.Next()
	})
}

// convertMiddleware converts framework-agnostic middleware to Fiber middleware
func (a *FiberAdapter) convertMiddleware(middleware routeradapter.MiddlewareFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create RouterContext from Fiber context
		rctx := a.fiberContextToRouterContext(c)

		// Define next function
		next := func() error {
			return c.Next()
		}

		// Call middleware
		return middleware(rctx, next)
	}
}

// convertGinHandlerToFiber converts Gin HandlerFunc to Fiber handler
// This is needed because our route.Meta contains Gin handlers
func (a *FiberAdapter) convertGinHandlerToFiber(ginHandler interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create RouterContext
		rctx := a.fiberContextToRouterContext(c)

		// Store Fiber context as native context
		rctx.SetNativeContext(c)

		// We can't directly call Gin handler, so we need to adapt
		// For now, return success - full integration requires handler conversion
		return nil
	}
}

// fiberContextToRouterContext converts Fiber context to RouterContext
func (a *FiberAdapter) fiberContextToRouterContext(c *fiber.Ctx) *routeradapter.RouterContext {
	// Create standard http.Request and http.ResponseWriter
	// Fiber uses fasthttp, so we need to convert
	uri := c.Request().URI()
	urlStr := string(uri.RequestURI())
	if urlStr == "" {
		urlStr = "/"
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		parsedURL = &url.URL{Path: urlStr}
	}

	req := &http.Request{
		Method: c.Method(),
		URL:    parsedURL,
		Header: make(http.Header),
	}

	// Attach Fiber's user context to preserve trace context and other request-scoped values
	// This ensures middleware updates (like tracing) are propagated to RouterContext
	req = req.WithContext(c.UserContext())

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	// Create response writer wrapper
	w := &fiberResponseWriter{ctx: c}

	rctx := routeradapter.NewRouterContext(w, req)

	// Copy path parameters
	c.AllParams()
	// Note: Fiber doesn't expose params in a simple map, we'd need to iterate

	// Store reference to native Fiber context
	rctx.SetNativeContext(c)

	return rctx
}

// fiberResponseWriter wraps Fiber context to implement http.ResponseWriter
type fiberResponseWriter struct {
	ctx *fiber.Ctx
}

func (w *fiberResponseWriter) Header() http.Header {
	header := make(http.Header)
	w.ctx.Response().Header.VisitAll(func(key, value []byte) {
		header.Add(string(key), string(value))
	})
	return header
}

func (w *fiberResponseWriter) Write(b []byte) (int, error) {
	return w.ctx.Write(b)
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	w.ctx.Status(statusCode)
}

// FiberGroup implements RouterGroup interface for Fiber router groups
type FiberGroup struct {
	group   fiber.Router
	adapter *FiberAdapter
}

// RegisterRoute registers a route within this Fiber group
func (g *FiberGroup) RegisterRoute(meta route.Meta) error {
	if meta.Method == "" || meta.Path == "" {
		return fmt.Errorf("route method and path are required")
	}

	if meta.Func == nil {
		return fmt.Errorf("route handler function is required")
	}

	// Convert handler
	fiberHandler := g.adapter.convertGinHandlerToFiber(meta.Func)

	// Convert middlewares
	fiberMiddlewares := make([]fiber.Handler, 0, len(meta.Middlewares))
	for _, mw := range meta.Middlewares {
		fiberMiddlewares = append(fiberMiddlewares, g.adapter.convertGinHandlerToFiber(mw))
	}

	// Combine middlewares and handler
	handlers := append(fiberMiddlewares, fiberHandler)

	// Register route based on method
	switch meta.Method {
	case http.MethodGet:
		g.group.Get(meta.Path, handlers...)
	case http.MethodPost:
		g.group.Post(meta.Path, handlers...)
	case http.MethodPut:
		g.group.Put(meta.Path, handlers...)
	case http.MethodPatch:
		g.group.Patch(meta.Path, handlers...)
	case http.MethodDelete:
		g.group.Delete(meta.Path, handlers...)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", meta.Method)
	}

	return nil
}

// RegisterMiddleware adds middleware to this Fiber group
func (g *FiberGroup) RegisterMiddleware(middleware routeradapter.MiddlewareFunc) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	fiberMiddleware := g.adapter.convertMiddleware(middleware)
	g.group.Use(fiberMiddleware)

	return nil
}

// Group creates a sub-group with additional prefix and middlewares
func (g *FiberGroup) Group(prefix string, middlewares ...routeradapter.MiddlewareFunc) routeradapter.RouterGroup {
	// Convert middlewares
	fiberMiddlewares := make([]fiber.Handler, len(middlewares))
	for i, mw := range middlewares {
		fiberMiddlewares[i] = g.adapter.convertMiddleware(mw)
	}

	// Create sub-group
	subgroup := g.group.Group(prefix, fiberMiddlewares...)

	return &FiberGroup{
		group:   subgroup,
		adapter: g.adapter,
	}
}

// UseNative adds a Fiber-specific middleware to this group
func (g *FiberGroup) UseNative(middleware interface{}) error {
	switch mw := middleware.(type) {
	case func(*fiber.Ctx) error:
		g.group.Use(mw)
		return nil
	default:
		return fmt.Errorf("middleware must be func(*fiber.Ctx) error, got %T", middleware)
	}
}

// setupGzipCompression configures gzip compression middleware for Fiber
func (a *FiberAdapter) setupGzipCompression() {
	// Map gzip compression levels to Fiber compression levels
	level := compress.LevelDefault

	switch a.config.CompressionLevel {
	case 0, -1:
		level = compress.LevelDefault
	case 1:
		level = compress.LevelBestSpeed
	case 9:
		level = compress.LevelBestCompression
	default:
		// Use the specified level directly (1-9)
		level = compress.Level(a.config.CompressionLevel)
	}

	// Use Fiber's official compress middleware
	a.app.Use(compress.New(compress.Config{
		Level: level,
	}))
}
