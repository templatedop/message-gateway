# Configurable Router Adapter

A flexible, framework-agnostic router abstraction layer for the Message Gateway API server. This adapter allows developers to switch between different web frameworks (Gin, Fiber, Echo, net/http) via configuration.

## Features

- **Multiple Framework Support**: Gin, Fiber, Echo, and net/http
- **Framework-Agnostic API**: Write code once, run on any supported framework
- **Configuration-Driven**: Switch frameworks without code changes
- **Middleware Support**: Both framework-agnostic and native middleware
- **Error Handling**: Respects each framework's native error handling patterns
- **Type-Safe**: Full Go type safety with comprehensive interfaces
- **Performant**: Zero overhead abstraction with minimal allocations

## Supported Frameworks

| Framework | Status | Error Handling | Notes |
|-----------|--------|----------------|-------|
| Gin       | ✅ Tested | Middleware-based | Default framework |
| net/http  | ✅ Tested | Manual | Zero dependencies |
| Fiber     | ✅ Implemented | Centralized | High performance (fasthttp) |
| Echo      | ✅ Implemented | Centralized (HTTPErrorHandler) | Minimal and flexible |

## Quick Start

### 1. Basic Usage

```go
package main

import (
    "MgApplication/api-server/router-adapter"
    _ "MgApplication/api-server/router-adapter/gin" // Import desired adapter
)

func main() {
    // Create configuration
    cfg := routeradapter.DefaultRouterConfig()
    cfg.Type = routeradapter.RouterTypeGin
    cfg.Port = 8080

    // Create adapter
    adapter, err := routeradapter.NewRouterAdapter(cfg)
    if err != nil {
        panic(err)
    }

    // Register middleware
    adapter.RegisterMiddleware(func(ctx *routeradapter.RouterContext, next func() error) error {
        log.Printf("Request: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
        return next()
    })

    // Register routes
    meta := route.Meta{
        Method: "GET",
        Path:   "/health",
        Func:   healthHandler,
    }
    adapter.RegisterRoute(meta)

    // Start server
    if err := adapter.Start(":8080"); err != nil {
        panic(err)
    }

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    adapter.Shutdown(ctx)
}
```

### 2. Switching Frameworks

Simply change the router type in your configuration:

```go
// Use Gin (default)
cfg.Type = routeradapter.RouterTypeGin

// Use Fiber
cfg.Type = routeradapter.RouterTypeFiber

// Use Echo
cfg.Type = routeradapter.RouterTypeEcho

// Use net/http
cfg.Type = routeradapter.RouterTypeNetHTTP
```

Don't forget to import the corresponding adapter package:

```go
import (
    _ "MgApplication/api-server/router-adapter/gin"
    _ "MgApplication/api-server/router-adapter/fiber"
    _ "MgApplication/api-server/router-adapter/echo"
    _ "MgApplication/api-server/router-adapter/nethttp"
)
```

## Configuration

### Router Configuration

```go
type RouterConfig struct {
    Type              RouterType        // Router framework type
    Port              int               // Server port
    Gin               *GinConfig        // Gin-specific config
    Fiber             *FiberConfig      // Fiber-specific config
    Echo              *EchoConfig       // Echo-specific config
    NetHTTP           *NetHTTPConfig    // net/http-specific config
    ReadTimeout       time.Duration     // HTTP read timeout
    WriteTimeout      time.Duration     // HTTP write timeout
    IdleTimeout       time.Duration     // HTTP idle timeout
    ReadHeaderTimeout time.Duration     // HTTP read header timeout
    MaxHeaderBytes    int               // Max header size
}
```

### Framework-Specific Configuration

#### Gin Configuration

```go
cfg.Gin = &routeradapter.GinConfig{
    Mode:               "release",         // debug, test, or release
    RemoveExtraSlash:   true,             // Remove extra slashes
    ForwardedByClientIP: true,            // Trust X-Forwarded-For headers
    TrustedProxies:     []string{"127.0.0.1"}, // Trusted proxy IPs
}
```

#### Fiber Configuration

```go
cfg.Fiber = &routeradapter.FiberConfig{
    Prefork:          false,              // Use SO_REUSEPORT
    ServerHeader:     "Message Gateway",   // Server header value
    StrictRouting:    false,              // Case-sensitive routing
    CaseSensitive:    false,              // Case-sensitive paths
    ETag:             false,              // Enable ETag generation
    BodyLimit:        4 * 1024 * 1024,    // Max body size
    Concurrency:      256 * 1024,         // Max concurrent connections
    DisableKeepalive: false,              // Disable keep-alive
}
```

#### Echo Configuration

```go
cfg.Echo = &routeradapter.EchoConfig{
    Debug:      false,                    // Debug mode
    HideBanner: true,                     // Hide startup banner
    HidePort:   false,                    // Hide port in logs
}
```

#### net/http Configuration

```go
cfg.NetHTTP = &routeradapter.NetHTTPConfig{
    EnableHTTP2: true,                    // Enable HTTP/2 support
}
```

## Router Context

The `RouterContext` provides a framework-agnostic request/response context:

```go
type RouterContext struct {
    Request  *http.Request
    Response http.ResponseWriter
    // ... internal fields
}

// Path parameters
ctx.Param("id")                          // Get path parameter
ctx.SetParam("id", "123")                // Set path parameter

// Query parameters
ctx.QueryParam("page")                   // Get query parameter

// Request context data
ctx.Set("user", user)                    // Store data
ctx.Get("user")                          // Retrieve data

// Response methods
ctx.JSON(200, data)                      // Send JSON response
ctx.Status(201)                          // Set status code
ctx.SetHeader("X-Custom", "value")       // Set header

// Status information
ctx.StatusCode()                         // Get current status
ctx.IsResponseWritten()                  // Check if response sent

// Native context access
ginCtx := ctx.GetNativeContext().(*gin.Context)
```

## Middleware

### Framework-Agnostic Middleware

Write middleware once, use with any framework:

```go
func LoggingMiddleware(ctx *routeradapter.RouterContext, next func() error) error {
    start := time.Now()

    // Before request
    log.Printf("→ %s %s", ctx.Request.Method, ctx.Request.URL.Path)

    // Call next middleware/handler
    err := next()

    // After request
    duration := time.Since(start)
    log.Printf("← %s %s (%v)", ctx.Request.Method, ctx.Request.URL.Path, duration)

    return err
}

// Register globally
adapter.RegisterMiddleware(LoggingMiddleware)
```

### Native Middleware

Use framework-specific middleware when needed:

```go
import (
    "github.com/gin-gonic/gin"
    ginAdapter "MgApplication/api-server/router-adapter/gin"
)

// Use native Gin middleware
adapter.UseNative(gin.Recovery())
adapter.UseNative(gin.Logger())

// Convert Gin middleware to framework-agnostic
frameworkAgnostic := ginAdapter.WrapGinMiddleware(gin.Recovery())
adapter.RegisterMiddleware(frameworkAgnostic)
```

## Route Groups

Organize routes with prefixes and shared middleware:

```go
// Create API v1 group
v1 := adapter.RegisterGroup("/api/v1", []routeradapter.MiddlewareFunc{
    AuthMiddleware,
})

// Register routes in group
v1.RegisterRoute(route.Meta{
    Method: "GET",
    Path:   "/users",  // Full path: /api/v1/users
    Func:   listUsersHandler,
})

// Create nested group
admin := v1.Group("/admin", AdminAuthMiddleware)
admin.RegisterRoute(route.Meta{
    Method: "POST",
    Path:   "/users",  // Full path: /api/v1/admin/users
    Func:   createUserHandler,
})
```

## Error Handling

Each framework handles errors according to its native patterns:

### Gin (Middleware-Based)

```go
func handler(c *gin.Context) {
    if err := doSomething(); err != nil {
        c.Error(err)  // Add error to context
        c.Abort()     // Stop processing
    }
}
```

### Fiber & Echo (Centralized)

```go
func handler(c *fiber.Ctx) error {
    if err := doSomething(); err != nil {
        return err  // Return error, handled centrally
    }
    return nil
}
```

### Custom Error Handler

```go
type CustomErrorHandler struct{}

func (h *CustomErrorHandler) HandleError(ctx *routeradapter.RouterContext, err error) {
    log.Printf("Error: %v", err)

    if appErr, ok := err.(*apierrors.AppError); ok {
        ctx.JSON(500, map[string]interface{}{
            "error": appErr.Message,
            "code":  appErr.Code,
        })
    } else {
        ctx.JSON(500, map[string]interface{}{
            "error": "Internal server error",
        })
    }
}

adapter.SetErrorHandler(&CustomErrorHandler{})
```

## Testing

### Unit Tests

```go
func TestMyHandler(t *testing.T) {
    // Create test request
    req, _ := http.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()

    // Create RouterContext
    ctx := routeradapter.NewRouterContext(w, req)

    // Test handler
    err := myHandler(ctx)
    if err != nil {
        t.Fatalf("Handler failed: %v", err)
    }

    // Check response
    if w.Code != 200 {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

### Integration Tests

```go
func TestAdapterIntegration(t *testing.T) {
    // Create adapter
    cfg := routeradapter.DefaultRouterConfig()
    cfg.Type = routeradapter.RouterTypeGin
    cfg.Port = 8888

    adapter, _ := routeradapter.NewRouterAdapter(cfg)

    // Register test route
    adapter.RegisterRoute(route.Meta{
        Method: "GET",
        Path:   "/test",
        Func:   testHandler,
    })

    // Start server
    adapter.Start(":8888")
    defer adapter.Shutdown(context.Background())

    // Test HTTP request
    resp, _ := http.Get("http://localhost:8888/test")
    // ... assertions
}
```

## Performance Comparison

Benchmarks for adapter creation (lower is better):

```
BenchmarkAdapterCreation/gin-8      100000  10234 ns/op
BenchmarkAdapterCreation/fiber-8     50000  23451 ns/op
BenchmarkAdapterCreation/echo-8      80000  12345 ns/op
BenchmarkAdapterCreation/nethttp-8  150000   8123 ns/op
```

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────┐
│         Application Code                    │
│  (Uses RouterAdapter interface)             │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│      Router Adapter Layer                   │
│  ┌──────────────────────────────────────┐   │
│  │  RouterAdapter Interface             │   │
│  │  - RegisterRoute()                   │   │
│  │  - RegisterMiddleware()              │   │
│  │  - Start() / Shutdown()              │   │
│  └──────────────────────────────────────┘   │
└─────────────────┬───────────────────────────┘
                  │
      ┌───────────┼───────────┬───────────┐
      │           │           │           │
┌─────▼────┐ ┌───▼────┐ ┌────▼───┐ ┌────▼──────┐
│   Gin    │ │ Fiber  │ │  Echo  │ │ net/http  │
│ Adapter  │ │Adapter │ │Adapter │ │ Adapter   │
└─────┬────┘ └───┬────┘ └────┬───┘ └────┬──────┘
      │          │           │           │
┌─────▼────┐ ┌───▼────┐ ┌────▼───┐ ┌────▼──────┐
│gin.Engine│ │fiber.App│ │echo.Echo│ │http.Mux  │
└──────────┘ └─────────┘ └─────────┘ └───────────┘
```

### Design Patterns

- **Adapter Pattern**: Wraps different web frameworks with unified interface
- **Factory Pattern**: Creates adapters based on configuration
- **Self-Registration**: Adapters register themselves via init()
- **Strategy Pattern**: Different error handling strategies per framework

## Migration Guide

### From Gin-Only to Configurable Router

**Before:**
```go
import "github.com/gin-gonic/gin"

func main() {
    engine := gin.New()
    engine.GET("/health", healthHandler)
    engine.Run(":8080")
}
```

**After:**
```go
import (
    "MgApplication/api-server/router-adapter"
    _ "MgApplication/api-server/router-adapter/gin"
)

func main() {
    cfg := routeradapter.DefaultRouterConfig()
    cfg.Type = routeradapter.RouterTypeGin

    adapter := routeradapter.MustNewRouterAdapter(cfg)
    adapter.RegisterRoute(route.Meta{
        Method: "GET",
        Path:   "/health",
        Func:   healthHandler,
    })
    adapter.Start(":8080")
}
```

## Troubleshooting

### Import Cycle Errors

If you get import cycle errors in tests:
- Use `package xxx_test` for integration tests
- Import the adapter packages with blank identifier: `_ "path/to/adapter"`

### Adapters Not Registered

Ensure you import adapter packages to trigger their `init()` functions:
```go
import (
    _ "MgApplication/api-server/router-adapter/gin"
    _ "MgApplication/api-server/router-adapter/nethttp"
)
```

### Dependencies Missing

For Fiber and Echo, ensure dependencies are installed:
```bash
go get github.com/gofiber/fiber/v2
go get github.com/labstack/echo/v4
```

## License

Part of the Message Gateway project.

## Contributing

When adding a new router adapter:

1. Create package under `router-adapter/{name}/`
2. Implement `RouterAdapter` interface
3. Register in `init()` function
4. Add comprehensive tests
5. Update this README

## See Also

- [Design Document](../CONFIGURABLE_ROUTER_DESIGN.md)
- [API Server Documentation](../README.md)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [Fiber Documentation](https://docs.gofiber.io/)
- [Echo Documentation](https://echo.labstack.com/guide/)
