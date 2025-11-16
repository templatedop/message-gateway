package routeradapter

import (
	"fmt"
	"sync"
)

// AdapterFactory is a function that creates a RouterAdapter from configuration
type AdapterFactory func(*RouterConfig) (RouterAdapter, error)

var (
	adapterFactories = make(map[RouterType]AdapterFactory)
	factoryMu        sync.RWMutex
)

// RegisterAdapterFactory registers a factory function for a router type
// This allows adapter implementations to register themselves
func RegisterAdapterFactory(routerType RouterType, factory AdapterFactory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	adapterFactories[routerType] = factory
}

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

	// Default to Gin if not specified
	if cfg.Type == "" {
		cfg.Type = RouterTypeGin
	}

	// Get factory for router type
	factoryMu.RLock()
	factory, exists := adapterFactories[cfg.Type]
	factoryMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no adapter registered for router type: %s", cfg.Type)
	}

	// Create adapter
	return factory(cfg)
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

// GetRegisteredAdapters returns a list of registered adapter types
// Useful for debugging and displaying available options
func GetRegisteredAdapters() []RouterType {
	factoryMu.RLock()
	defer factoryMu.RUnlock()

	adapters := make([]RouterType, 0, len(adapterFactories))
	for routerType := range adapterFactories {
		adapters = append(adapters, routerType)
	}
	return adapters
}
