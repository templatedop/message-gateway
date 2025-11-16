package pprof

import (
	"net/http/pprof"

	"MgApplication/api-server/router-adapter"
)

// RegisterPprofRoutes registers all pprof endpoints on the router
// Registers the following endpoints:
//   - /debug/pprof/ (index)
//   - /debug/pprof/cmdline
//   - /debug/pprof/profile
//   - /debug/pprof/symbol
//   - /debug/pprof/trace
//   - /debug/pprof/allocs
//   - /debug/pprof/block
//   - /debug/pprof/goroutine
//   - /debug/pprof/heap
//   - /debug/pprof/mutex
//   - /debug/pprof/threadcreate
func RegisterPprofRoutes(group routeradapter.RouterGroup) error {
	// Index handler
	if err := group.RegisterMiddleware(PprofIndexHandler("/debug/pprof/")); err != nil {
		return err
	}

	// Cmdline handler
	if err := group.RegisterMiddleware(PprofCmdlineHandler("/debug/pprof/cmdline")); err != nil {
		return err
	}

	// Profile handler
	if err := group.RegisterMiddleware(PprofProfileHandler("/debug/pprof/profile")); err != nil {
		return err
	}

	// Symbol handler
	if err := group.RegisterMiddleware(PprofSymbolHandler("/debug/pprof/symbol")); err != nil {
		return err
	}

	// Trace handler
	if err := group.RegisterMiddleware(PprofTraceHandler("/debug/pprof/trace")); err != nil {
		return err
	}

	// Allocs handler
	if err := group.RegisterMiddleware(PprofAllocsHandler("/debug/pprof/allocs")); err != nil {
		return err
	}

	// Block handler
	if err := group.RegisterMiddleware(PprofBlockHandler("/debug/pprof/block")); err != nil {
		return err
	}

	// Goroutine handler
	if err := group.RegisterMiddleware(PprofGoroutineHandler("/debug/pprof/goroutine")); err != nil {
		return err
	}

	// Heap handler
	if err := group.RegisterMiddleware(PprofHeapHandler("/debug/pprof/heap")); err != nil {
		return err
	}

	// Mutex handler
	if err := group.RegisterMiddleware(PprofMutexHandler("/debug/pprof/mutex")); err != nil {
		return err
	}

	// ThreadCreate handler
	if err := group.RegisterMiddleware(PprofThreadCreateHandler("/debug/pprof/threadcreate")); err != nil {
		return err
	}

	return nil
}

// PprofIndexHandler returns a middleware that serves the pprof index page
func PprofIndexHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Index(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofCmdlineHandler returns a middleware that serves the pprof cmdline endpoint
func PprofCmdlineHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Cmdline(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofProfileHandler returns a middleware that serves the pprof profile endpoint
func PprofProfileHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Profile(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofSymbolHandler returns a middleware that serves the pprof symbol endpoint
func PprofSymbolHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Symbol(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofTraceHandler returns a middleware that serves the pprof trace endpoint
func PprofTraceHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Trace(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofAllocsHandler returns a middleware that serves the pprof allocs endpoint
func PprofAllocsHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Handler("allocs").ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofBlockHandler returns a middleware that serves the pprof block endpoint
func PprofBlockHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Handler("block").ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofGoroutineHandler returns a middleware that serves the pprof goroutine endpoint
func PprofGoroutineHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Handler("goroutine").ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofHeapHandler returns a middleware that serves the pprof heap endpoint
func PprofHeapHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Handler("heap").ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofMutexHandler returns a middleware that serves the pprof mutex endpoint
func PprofMutexHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Handler("mutex").ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}
}

// PprofThreadCreateHandler returns a middleware that serves the pprof threadcreate endpoint
func PprofThreadCreateHandler(path string) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		if ctx.Request.URL.Path != path {
			return next()
		}
		pprof.Handler("threadcreate").ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}
}
