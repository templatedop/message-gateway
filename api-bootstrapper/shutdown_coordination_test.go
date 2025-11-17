package bootstrapper

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestShutdownCoordination verifies that router shuts down before database
func TestShutdownCoordination(t *testing.T) {
	var (
		routerStartTime    time.Time
		routerStopTime     time.Time
		dbStartTime        time.Time
		dbStopTime         time.Time
		routerStopComplete time.Time
		dbStopComplete     time.Time
		mu                 sync.Mutex
	)

	// Create a test app with mock router and database modules
	app := fxtest.New(
		t,
		fx.NopLogger,

		// Mock Database Module (simulates fxDB)
		fx.Module("test-db",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						mu.Lock()
						dbStartTime = time.Now()
						mu.Unlock()
						t.Log("Database OnStart executed")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						mu.Lock()
						dbStopTime = time.Now()
						mu.Unlock()
						t.Log("Database OnStop started")

						// Simulate database drain (5 seconds timeout, but complete quickly)
						drainTimeout := 5 * time.Second
						drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
						defer cancel()

						// Simulate checking for active connections
						time.Sleep(500 * time.Millisecond) // Simulate some drain time

						select {
						case <-drainCtx.Done():
							t.Log("Database drain timeout")
						default:
							t.Log("Database connections drained")
						}

						mu.Lock()
						dbStopComplete = time.Now()
						mu.Unlock()
						t.Log("Database OnStop completed")
						return nil
					},
				})
			}),
		),

		// Mock Router Module (simulates fxRouterAdapter)
		fx.Module("test-router",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						mu.Lock()
						routerStartTime = time.Now()
						mu.Unlock()
						t.Log("Router OnStart executed")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						mu.Lock()
						routerStopTime = time.Now()
						mu.Unlock()
						t.Log("Router OnStop started")

						// Simulate router shutdown (10 seconds timeout)
						shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
						defer cancel()

						// Simulate waiting for handlers to complete
						handlerCompletionTime := 2 * time.Second
						select {
						case <-time.After(handlerCompletionTime):
							t.Log("All HTTP handlers completed")
						case <-shutdownCtx.Done():
							t.Log("Router shutdown timeout")
						}

						mu.Lock()
						routerStopComplete = time.Now()
						mu.Unlock()
						t.Log("Router OnStop completed")
						return nil
					},
				})
			}),
		),
	)

	// Start the app
	app.RequireStart()

	// Give modules time to start
	time.Sleep(100 * time.Millisecond)

	// Stop the app (triggers shutdown)
	t.Log("Triggering application shutdown...")
	app.RequireStop()

	// Verify shutdown coordination
	mu.Lock()
	defer mu.Unlock()

	// Verify both modules started
	if routerStartTime.IsZero() {
		t.Error("Router OnStart was not called")
	}
	if dbStartTime.IsZero() {
		t.Error("Database OnStart was not called")
	}

	// Verify both modules stopped
	if routerStopTime.IsZero() {
		t.Error("Router OnStop was not called")
	}
	if dbStopTime.IsZero() {
		t.Error("Database OnStop was not called")
	}

	// CRITICAL: Verify router completed shutdown BEFORE database started shutdown
	if !routerStopComplete.Before(dbStopTime) {
		t.Errorf("Router OnStop should complete BEFORE Database OnStop starts\nRouter complete: %v\nDB start: %v",
			routerStopComplete, dbStopTime)
	}

	// Log timing for analysis
	t.Logf("\n=== Shutdown Timing ===")
	t.Logf("Router OnStop start:    %v", routerStopTime.Format("15:04:05.000"))
	t.Logf("Router OnStop complete: %v (duration: %v)",
		routerStopComplete.Format("15:04:05.000"),
		routerStopComplete.Sub(routerStopTime))
	t.Logf("DB OnStop start:        %v (delay: %v after router complete)",
		dbStopTime.Format("15:04:05.000"),
		dbStopTime.Sub(routerStopComplete))
	t.Logf("DB OnStop complete:     %v (duration: %v)",
		dbStopComplete.Format("15:04:05.000"),
		dbStopComplete.Sub(dbStopTime))
	t.Logf("Total shutdown time:    %v", dbStopComplete.Sub(routerStopTime))
}

// TestShutdownOrderWithModulePositioning verifies FX shutdown order based on module position
func TestShutdownOrderWithModulePositioning(t *testing.T) {
	var (
		shutdownOrder []string
		mu            sync.Mutex
	)

	recordShutdown := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		shutdownOrder = append(shutdownOrder, name)
	}

	// Create app with modules in specific order (simulating actual bootstrapper)
	app := fxtest.New(
		t,
		fx.NopLogger,

		// Module 1: Config (position 1)
		fx.Module("config",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						t.Log("Config started")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						recordShutdown("config")
						t.Log("Config stopped")
						return nil
					},
				})
			}),
		),

		// Module 2: Log (position 2)
		fx.Module("log",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						t.Log("Log started")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						recordShutdown("log")
						t.Log("Log stopped")
						return nil
					},
				})
			}),
		),

		// Module 3: Database (position 3)
		fx.Module("database",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						t.Log("Database started")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						recordShutdown("database")
						t.Log("Database stopped")
						return nil
					},
				})
			}),
		),

		// Module 4: Router (position 4)
		fx.Module("router",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						t.Log("Router started")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						recordShutdown("router")
						t.Log("Router stopped")
						return nil
					},
				})
			}),
		),

		// Module 5: Metrics (position 5)
		fx.Module("metrics",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						t.Log("Metrics started")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						recordShutdown("metrics")
						t.Log("Metrics stopped")
						return nil
					},
				})
			}),
		),
	)

	// Start and stop
	app.RequireStart()
	time.Sleep(100 * time.Millisecond)
	app.RequireStop()

	// Verify shutdown order is reverse of startup order
	mu.Lock()
	defer mu.Unlock()

	expectedOrder := []string{"metrics", "router", "database", "log", "config"}
	if len(shutdownOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d shutdown events, got %d", len(expectedOrder), len(shutdownOrder))
	}

	for i, expected := range expectedOrder {
		if shutdownOrder[i] != expected {
			t.Errorf("Shutdown order[%d]: expected %s, got %s", i, expected, shutdownOrder[i])
		}
	}

	t.Logf("✓ Shutdown order correct: %v", shutdownOrder)

	// Verify router shut down before database
	routerIndex := -1
	dbIndex := -1
	for i, module := range shutdownOrder {
		if module == "router" {
			routerIndex = i
		}
		if module == "database" {
			dbIndex = i
		}
	}

	if routerIndex == -1 || dbIndex == -1 {
		t.Fatal("Router or Database not found in shutdown order")
	}

	if routerIndex >= dbIndex {
		t.Errorf("Router should shut down BEFORE database. Router: %d, Database: %d", routerIndex, dbIndex)
	}

	t.Logf("✓ Router (index %d) shut down before Database (index %d)", routerIndex, dbIndex)
}

// TestDatabaseAvailabilityDuringRouterShutdown verifies database remains available
func TestDatabaseAvailabilityDuringRouterShutdown(t *testing.T) {
	var (
		dbAvailable        bool
		dbOperationSuccess bool
		mu                 sync.Mutex
	)

	app := fxtest.New(
		t,
		fx.NopLogger,

		// Mock database that stays available
		fx.Module("database",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						mu.Lock()
						dbAvailable = true
						mu.Unlock()
						t.Log("Database started and available")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						t.Log("Database OnStop started - closing database")
						time.Sleep(100 * time.Millisecond) // Simulate drain
						mu.Lock()
						dbAvailable = false
						mu.Unlock()
						t.Log("Database closed")
						return nil
					},
				})
			}),
		),

		// Mock router that uses database during shutdown
		fx.Module("router",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						t.Log("Router started")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						t.Log("Router OnStop started - simulating handler using database")

						// Simulate a handler that's still processing during shutdown
						time.Sleep(500 * time.Millisecond)

						// Try to use database (simulating handler DB operation)
						mu.Lock()
						if dbAvailable {
							dbOperationSuccess = true
							t.Log("✓ Database operation during router shutdown: SUCCESS")
						} else {
							t.Log("✗ Database operation during router shutdown: FAILED (DB not available)")
						}
						mu.Unlock()

						t.Log("Router OnStop completed")
						return nil
					},
				})
			}),
		),
	)

	app.RequireStart()
	time.Sleep(100 * time.Millisecond)
	app.RequireStop()

	mu.Lock()
	defer mu.Unlock()

	if !dbOperationSuccess {
		t.Error("Database should be available during router shutdown, but operation failed")
	} else {
		t.Log("✓ Database was available during router shutdown")
	}
}

// TestShutdownTimeouts verifies timeout behavior
func TestShutdownTimeouts(t *testing.T) {
	var (
		routerTimedOut bool
		mu             sync.Mutex
	)

	app := fxtest.New(
		t,
		fx.NopLogger,

		fx.Module("slow-router",
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return nil
					},
					OnStop: func(ctx context.Context) error {
						shutdownCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
						defer cancel()

						// Simulate a handler that takes too long
						select {
						case <-time.After(2 * time.Second):
							// This would only happen if timeout doesn't work
							t.Log("Handler completed (shouldn't happen)")
						case <-shutdownCtx.Done():
							mu.Lock()
							routerTimedOut = true
							mu.Unlock()
							t.Log("Router shutdown timeout reached (expected)")
						}

						return shutdownCtx.Err()
					},
				})
			}),
		),
	)

	app.RequireStart()
	time.Sleep(100 * time.Millisecond)

	// Note: RequireStop expects clean shutdown, so we expect error here
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Stop(stopCtx)
	if err == nil {
		t.Log("App stopped with timeout (expected)")
	}

	mu.Lock()
	defer mu.Unlock()

	if !routerTimedOut {
		t.Error("Router shutdown should have timed out")
	} else {
		t.Log("✓ Router shutdown timeout worked correctly")
	}
}

// TestContextPropagationInShutdown verifies signal-aware context reaches all modules during shutdown
func TestContextPropagationInShutdown(t *testing.T) {
	var (
		routerReceivedContext bool
		dbReceivedContext     bool
		mu                    sync.Mutex
	)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := fxtest.New(
		t,
		fx.Supply(fx.Annotate(ctx, fx.As(new(context.Context)))),
		fx.NopLogger,

		fx.Module("database",
			fx.Invoke(func(lc fx.Lifecycle, appCtx context.Context) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						mu.Lock()
						dbReceivedContext = (appCtx != nil)
						mu.Unlock()
						t.Logf("Database received context: %v", appCtx != nil)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return nil
					},
				})
			}),
		),

		fx.Module("router",
			fx.Invoke(func(lc fx.Lifecycle, appCtx context.Context) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						mu.Lock()
						routerReceivedContext = (appCtx != nil)
						mu.Unlock()
						t.Logf("Router received context: %v", appCtx != nil)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return nil
					},
				})
			}),
		),
	)

	app.RequireStart()
	app.RequireStop()

	mu.Lock()
	defer mu.Unlock()

	if !routerReceivedContext {
		t.Error("Router did not receive context")
	}
	if !dbReceivedContext {
		t.Error("Database did not receive context")
	}

	if routerReceivedContext && dbReceivedContext {
		t.Log("✓ Context properly propagated to both router and database")
	}
}
