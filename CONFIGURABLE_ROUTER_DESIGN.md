# Configurable Router Design

## Overview

This design document outlines the implementation of a configurable router system that allows developers to switch between different web frameworks (Gin, Fiber, Echo, net/http) via configuration.

## Problem Statement

Currently, the API server is tightly coupled to the Gin web framework. Different frameworks have different characteristics:

- **Gin**: Middleware-based error handling, good for traditional REST APIs
- **Fiber**: Centralized error handling, high performance, Express-like API
- **Echo**: Centralized error handling, minimal, flexible
- **net/http**: Standard library, maximum control, no framework overhead

Developers should be able to choose the framework that best suits their needs via configuration.

## Architecture

### 1. Router Abstraction Layer

Create a framework-agnostic interface that all router adapters implement:

```go
// RouterAdapter is the main interface that all framework adapters implement
type RouterAdapter interface {
    // Engine returns the underlying framework instance
    Engine() interface{}

    // RegisterRoute registers a route with the router
    RegisterRoute(meta route.Meta) error

    // RegisterMiddleware adds middleware to the router
    RegisterMiddleware(middleware Middleware) error

    // RegisterGroup creates a route group with optional middlewares
    RegisterGroup(prefix string, middlewares []Middleware) RouterGroup

    // ServeHTTP implements http.Handler
    ServeHTTP(w http.ResponseWriter, r *http.Request)

    // Start starts the HTTP server
    Start(addr string) error

    // Shutdown gracefully shuts down the server
    Shutdown(ctx context.Context) error
}

// RouterGroup represents a group of routes with common prefix/middlewares
type RouterGroup interface {
    RegisterRoute(meta route.Meta) error
    RegisterMiddleware(middleware Middleware) error
}

// Middleware is a framework-agnostic middleware function
type Middleware interface {
    Handle(ctx *RouterContext, next func() error) error
}

// RouterContext is a framework-agnostic request/response context
type RouterContext struct {
    Request  *http.Request
    Response http.ResponseWriter
    Params   map[string]string
    Data     map[string]interface{} // For storing request-scoped data
}
```

### 2. Framework Adapters

Implement adapters for each supported framework:

#### Gin Adapter
```go
type GinAdapter struct {
    engine *gin.Engine
    server *http.Server
}

// Handles error via middleware
// Existing implementation mostly compatible
```

#### Fiber Adapter
```go
type FiberAdapter struct {
    app *fiber.App
}

// Centralized error handling via app.config.ErrorHandler
// Convert route.Meta to fiber routes
// Convert middlewares to fiber middleware
```

#### Echo Adapter
```go
type EchoAdapter struct {
    echo *echo.Echo
}

// Centralized error handling via echo.HTTPErrorHandler
// Convert route.Meta to echo routes
// Convert middlewares to echo middleware
```

#### Net/HTTP Adapter
```go
type NetHTTPAdapter struct {
    mux    *http.ServeMux
    server *http.Server
}

// Manual error handling
// Basic routing with net/http
// Middleware chain implementation
```

### 3. Error Handling Strategy

Different frameworks handle errors differently:

**Gin** (Middleware-based):
```go
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        if len(c.Errors) > 0 {
            // Handle errors
        }
    }
}
```

**Fiber/Echo** (Centralized):
```go
// Fiber
app.config.ErrorHandler = func(c *fiber.Ctx, err error) error {
    // Handle error centrally
    return c.Status(500).JSON(errorResponse)
}

// Echo
e.HTTPErrorHandler = func(err error, c echo.Context) {
    // Handle error centrally
    c.JSON(500, errorResponse)
}
```

**Strategy**: Create error handler adapters that normalize error handling:

```go
type ErrorHandler interface {
    HandleError(ctx *RouterContext, err error)
}

type GinErrorHandler struct{}
type FiberErrorHandler struct{}
type EchoErrorHandler struct{}
type NetHTTPErrorHandler struct{}
```

### 4. Configuration

Add configuration support to select router:

```yaml
# config.yaml
server:
  router: "gin"  # Options: gin, fiber, echo, nethttp
  port: 8080

  # Router-specific configs
  gin:
    mode: "release"

  fiber:
    prefork: false
    caseSensitive: false

  echo:
    debug: false

  nethttp:
    readTimeout: 30s
    writeTimeout: 30s
```

Configuration structure:
```go
type RouterConfig struct {
    Type      string          // gin, fiber, echo, nethttp
    Port      int
    GinConfig *GinConfig
    FiberConfig *FiberConfig
    EchoConfig *EchoConfig
    NetHTTPConfig *NetHTTPConfig
}
```

### 5. Factory Pattern

Use factory pattern to create appropriate router adapter:

```go
func NewRouterAdapter(cfg *RouterConfig) (RouterAdapter, error) {
    switch strings.ToLower(cfg.Type) {
    case "gin", "":
        return NewGinAdapter(cfg.GinConfig)
    case "fiber":
        return NewFiberAdapter(cfg.FiberConfig)
    case "echo":
        return NewEchoAdapter(cfg.EchoConfig)
    case "nethttp", "http":
        return NewNetHTTPAdapter(cfg.NetHTTPConfig)
    default:
        return nil, fmt.Errorf("unsupported router type: %s", cfg.Type)
    }
}
```

### 6. Middleware Conversion

Convert existing Gin middlewares to framework-agnostic:

```go
// Wrapper that converts framework-specific middleware to our interface
type GinMiddlewareWrapper struct {
    handler gin.HandlerFunc
}

type FiberMiddlewareWrapper struct {
    handler fiber.Handler
}

// ... etc
```

### 7. Route Registration

Modify existing route registration to use adapter:

```go
// Before (Gin-specific)
func (r *Router) registerRoutes(app *gin.Engine, registries []*registry) {
    for _, reg := range registries {
        group := app.Group(reg.base, reg.mws...)
        for _, route := range reg.routes {
            meta := reg.toMeta(route)
            group.Handle(meta.Method, meta.Path, meta.Func)
        }
    }
}

// After (Framework-agnostic)
func (r *Router) registerRoutes(adapter RouterAdapter, registries []*registry) error {
    for _, reg := range registries {
        group := adapter.RegisterGroup(reg.base, convertMiddlewares(reg.mws))
        for _, route := range reg.routes {
            meta := reg.toMeta(route)
            if err := group.RegisterRoute(meta); err != nil {
                return err
            }
        }
    }
    return nil
}
```

## Implementation Plan

### Phase 1: Abstraction Layer
1. Create `api-server/router-adapter/` package
2. Define `RouterAdapter` interface
3. Define `RouterContext` struct
4. Define error handler interfaces

### Phase 2: Gin Adapter (Reference Implementation)
1. Implement `GinAdapter` wrapping existing Gin code
2. This serves as reference and maintains backward compatibility
3. Ensure all existing tests pass

### Phase 3: Fiber Adapter
1. Add Fiber dependency
2. Implement `FiberAdapter`
3. Handle centralized error handling
4. Convert route.Meta to Fiber routes
5. Test basic functionality

### Phase 4: Echo Adapter
1. Add Echo dependency
2. Implement `EchoAdapter`
3. Handle centralized error handling
4. Convert route.Meta to Echo routes
5. Test basic functionality

### Phase 5: Net/HTTP Adapter
1. Implement `NetHTTPAdapter`
2. Build middleware chain system
3. Implement basic routing with path parameters
4. Test basic functionality

### Phase 6: Configuration & Factory
1. Add configuration schema
2. Implement factory pattern
3. Update server initialization
4. Add configuration validation

### Phase 7: Testing & Documentation
1. Create comprehensive tests for each adapter
2. Create benchmarks comparing frameworks
3. Document usage and migration guide
4. Create examples for each framework

## Directory Structure

```
api-server/
├── router-adapter/
│   ├── adapter.go              # RouterAdapter interface
│   ├── context.go              # RouterContext implementation
│   ├── error_handler.go        # Error handler interfaces
│   ├── middleware.go           # Middleware interfaces
│   ├── gin/
│   │   ├── adapter.go          # GinAdapter implementation
│   │   ├── middleware.go       # Gin middleware wrappers
│   │   └── error_handler.go    # Gin error handler
│   ├── fiber/
│   │   ├── adapter.go          # FiberAdapter implementation
│   │   ├── middleware.go       # Fiber middleware wrappers
│   │   └── error_handler.go    # Fiber error handler
│   ├── echo/
│   │   ├── adapter.go          # EchoAdapter implementation
│   │   ├── middleware.go       # Echo middleware wrappers
│   │   └── error_handler.go    # Echo error handler
│   ├── nethttp/
│   │   ├── adapter.go          # NetHTTPAdapter implementation
│   │   ├── middleware.go       # Net/HTTP middleware chain
│   │   ├── router.go           # Path parameter routing
│   │   └── error_handler.go    # Net/HTTP error handler
│   ├── factory.go              # RouterAdapter factory
│   └── config.go               # Router configuration
├── server.go                    # Updated to use RouterAdapter
└── ...
```

## Migration Strategy

### Backward Compatibility

- Default router remains Gin
- No breaking changes to existing handler code
- Existing tests continue to work
- Configuration is optional (defaults to Gin)

### Migration Path

1. **No changes needed** - If not configured, uses Gin by default
2. **Optional migration** - Add `server.router: fiber` to config
3. **Test with new router** - Verify application works
4. **Switch back if needed** - Change config back to Gin

## Performance Considerations

### Abstraction Overhead

- Minimal overhead from interface calls (virtual dispatch)
- No reflection in hot path
- Framework-specific optimizations preserved within adapters

### Benchmark Plan

Create benchmarks comparing:
- Request handling latency
- Throughput (requests/second)
- Memory allocations
- JSON serialization (already using goccy/go-json)

## Security Considerations

1. **Input validation** - Preserve existing validation regardless of framework
2. **CORS** - Adapt CORS middleware to each framework
3. **Rate limiting** - Ensure rate limiting works across frameworks
4. **Error messages** - Consistent error responses across frameworks

## Benefits

1. **Framework flexibility** - Choose best framework for use case
2. **Performance tuning** - Switch to faster framework if needed
3. **Future-proof** - Easy to add new frameworks
4. **Learning** - Developers can experiment with different frameworks
5. **Migration path** - Easier to migrate between frameworks

## Risks & Mitigations

### Risk 1: Abstraction Complexity
**Mitigation**: Keep abstraction minimal, document clearly

### Risk 2: Framework-specific Features
**Mitigation**: Document which features are framework-specific, provide escape hatches

### Risk 3: Maintenance Burden
**Mitigation**: Focus on core features, don't support every framework feature

### Risk 4: Performance Regression
**Mitigation**: Comprehensive benchmarking, optimize hot paths

## Success Criteria

1. ✅ All existing tests pass with Gin adapter
2. ✅ Can switch between frameworks via configuration
3. ✅ Performance within 5% of native framework usage
4. ✅ Comprehensive documentation and examples
5. ✅ All four frameworks (Gin, Fiber, Echo, net/http) working

## Timeline Estimate

- Phase 1: 4 hours (Abstraction layer)
- Phase 2: 2 hours (Gin adapter)
- Phase 3: 4 hours (Fiber adapter)
- Phase 4: 4 hours (Echo adapter)
- Phase 5: 6 hours (Net/HTTP adapter - most complex)
- Phase 6: 3 hours (Configuration & factory)
- Phase 7: 5 hours (Testing & documentation)

**Total**: ~28 hours of development time

## Next Steps

1. Get approval on design
2. Start with Phase 1 (abstraction layer)
3. Implement Gin adapter (Phase 2) to validate design
4. Proceed with other adapters
5. Add configuration and factory
6. Comprehensive testing and documentation
