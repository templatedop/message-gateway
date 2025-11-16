package bootstrapper

import (
	"context"
	"testing"

	"MgApplication/api-server/router-adapter"
	"MgApplication/config"
)

// TestRouterAdapterCreation tests that the router adapter can be created from config
func TestRouterAdapterCreation(t *testing.T) {
	// Create a minimal config
	cfg, err := config.New(
		config.WithFileName("config"),
		config.WithAppEnv("test"),
		config.WithFilePaths("../configs"),
	)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify router.type is set in config
	if !cfg.Exists("router.type") {
		t.Fatal("router.type not found in config")
	}

	routerType := cfg.GetString("router.type")
	t.Logf("Router type from config: %s", routerType)

	if routerType != "fiber" {
		t.Errorf("Expected router type 'fiber', got '%s'", routerType)
	}
}

// TestFiberAdapterRegistered tests that the Fiber adapter factory is registered
func TestFiberAdapterRegistered(t *testing.T) {
	// Import adapter packages to register factories
	_ "MgApplication/api-server/router-adapter/gin"
	_ "MgApplication/api-server/router-adapter/fiber"
	_ "MgApplication/api-server/router-adapter/echo"
	_ "MgApplication/api-server/router-adapter/nethttp"

	registeredAdapters := routeradapter.GetRegisteredAdapters()
	t.Logf("Registered adapters: %v", registeredAdapters)

	// Check if fiber is registered
	fiberRegistered := false
	for _, adapter := range registeredAdapters {
		if adapter == routeradapter.RouterTypeFiber {
			fiberRegistered = true
			break
		}
	}

	if !fiberRegistered {
		t.Error("Fiber adapter is not registered")
	}
}

// TestNewRouterAdapterWithFiber tests creating a Fiber router adapter
func TestNewRouterAdapterWithFiber(t *testing.T) {
	// Import adapter packages to register factories
	_ "MgApplication/api-server/router-adapter/gin"
	_ "MgApplication/api-server/router-adapter/fiber"
	_ "MgApplication/api-server/router-adapter/echo"
	_ "MgApplication/api-server/router-adapter/nethttp"

	// Create router config for Fiber
	cfg := routeradapter.DefaultRouterConfig()
	cfg.Type = routeradapter.RouterTypeFiber

	// Create adapter
	adapter, err := routeradapter.NewRouterAdapter(cfg)
	if err != nil {
		t.Fatalf("Failed to create Fiber adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Adapter is nil")
	}

	// Set context (this should not panic)
	ctx := context.Background()
	adapter.SetContext(ctx)

	t.Log("Successfully created and configured Fiber adapter")
}

// TestRouterAdapterParams tests the router adapter FX parameters structure
func TestRouterAdapterParams(t *testing.T) {
	// Load config
	cfg, err := config.New(
		config.WithFileName("config"),
		config.WithAppEnv("test"),
		config.WithFilePaths("../configs"),
	)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create params structure (simulating what FX would inject)
	params := routerAdapterParams{
		Ctx:    context.Background(),
		Config: cfg,
		// Osdktrace and Registry would be provided by FX in real scenario
		Osdktrace: nil,
		Registry:  nil,
	}

	// Import adapter packages to register factories
	_ "MgApplication/api-server/router-adapter/gin"
	_ "MgApplication/api-server/router-adapter/fiber"
	_ "MgApplication/api-server/router-adapter/echo"
	_ "MgApplication/api-server/router-adapter/nethttp"

	// Call newRouterAdapter function
	adapter, err := newRouterAdapter(params)
	if err != nil {
		t.Fatalf("newRouterAdapter failed: %v", err)
	}

	if adapter == nil {
		t.Fatal("Adapter is nil")
	}

	t.Log("Successfully created router adapter via newRouterAdapter function")
}
