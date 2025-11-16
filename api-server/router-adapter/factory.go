package routeradapter

import (
	"fmt"
)

// NewRouterAdapter creates a new RouterAdapter based on the provided configuration
// This is the factory function that selects and instantiates the appropriate adapter
//
// Supported router types:
//   - gin: Gin web framework (default)
//   - fiber: Fiber web framework
//   - echo: Echo web framework
//   - nethttp: Standard library net/http
//
// Returns error if:
//   - Configuration is invalid
//   - Unsupported router type is specified
//   - Framework-specific initialization fails
func NewRouterAdapter(cfg *RouterConfig) (RouterAdapter, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid router config: %w", err)
	}

	// Create adapter based on type
	switch cfg.Type {
	case RouterTypeGin, "":
		// Gin is the default framework
		return NewGinAdapter(cfg)

	case RouterTypeFiber:
		return NewFiberAdapter(cfg)

	case RouterTypeEcho:
		return NewEchoAdapter(cfg)

	case RouterTypeNetHTTP:
		return NewNetHTTPAdapter(cfg)

	default:
		return nil, fmt.Errorf("unsupported router type: %s", cfg.Type)
	}
}

// MustNewRouterAdapter is like NewRouterAdapter but panics on error
// Useful for initialization where you want to fail fast
func MustNewRouterAdapter(cfg *RouterConfig) RouterAdapter {
	adapter, err := NewRouterAdapter(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create router adapter: %v", err))
	}
	return adapter
}

// NewRouterAdapterFromType creates a router adapter from just the type
// Uses default configuration for the specified type
// Useful for testing or simple use cases
func NewRouterAdapterFromType(routerType RouterType) (RouterAdapter, error) {
	cfg := DefaultRouterConfig()
	cfg.Type = routerType
	return NewRouterAdapter(cfg)
}

// Adapter factory functions
// These are implemented in separate files for each framework

// NewGinAdapter creates a new Gin router adapter
// Implemented in gin/adapter.go
func NewGinAdapter(cfg *RouterConfig) (RouterAdapter, error) {
	// Import guard: prevent circular dependency
	// Actual implementation is in gin/adapter.go
	return nil, fmt.Errorf("gin adapter not yet implemented - see gin/adapter.go")
}

// NewFiberAdapter creates a new Fiber router adapter
// Implemented in fiber/adapter.go
func NewFiberAdapter(cfg *RouterConfig) (RouterAdapter, error) {
	return nil, fmt.Errorf("fiber adapter not yet implemented - see fiber/adapter.go")
}

// NewEchoAdapter creates a new Echo router adapter
// Implemented in echo/adapter.go
func NewEchoAdapter(cfg *RouterConfig) (RouterAdapter, error) {
	return nil, fmt.Errorf("echo adapter not yet implemented - see echo/adapter.go")
}

// NewNetHTTPAdapter creates a new net/http router adapter
// Implemented in nethttp/adapter.go
func NewNetHTTPAdapter(cfg *RouterConfig) (RouterAdapter, error) {
	return nil, fmt.Errorf("nethttp adapter not yet implemented - see nethttp/adapter.go")
}
