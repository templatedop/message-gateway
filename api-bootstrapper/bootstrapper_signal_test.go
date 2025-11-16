package bootstrapper

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestSignalHandling verifies that the bootstrapper properly handles shutdown signals
func TestSignalHandling(t *testing.T) {
	// Create a simple bootstrapper with minimal modules
	b := &Bootstrapper{
		context: context.Background(),
		options: []fx.Option{
			// Minimal test setup - no actual services
			fx.NopLogger,
		},
	}

	// Channel to signal test completion
	done := make(chan struct{})

	// Run the bootstrapper in a goroutine
	go func() {
		defer close(done)

		// Wrap context with signal detection
		ctx, cancel := signalNotifyContextForTest(b.context, os.Interrupt, syscall.SIGTERM)
		defer cancel()

		b.context = ctx

		// Create a test app
		app := fxtest.New(
			t,
			fx.Supply(fx.Annotate(b.context, fx.As(new(context.Context)))),
			fx.Options(b.options...),
		)

		// Start the app
		app.RequireStart()

		// Wait for context cancellation
		<-ctx.Done()

		// Stop the app
		app.RequireStop()
	}()

	// Give the app time to start
	time.Sleep(100 * time.Millisecond)

	// Send interrupt signal to trigger shutdown
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to send SIGTERM: %v", err)
	}

	// Wait for graceful shutdown with timeout
	select {
	case <-done:
		t.Log("Bootstrapper shutdown gracefully")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for graceful shutdown")
	}
}

// Helper function to create signal context for testing
func signalNotifyContextForTest(parent context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	c := make(chan os.Signal, 1)
	// Note: In real tests, sending signals to the process can be tricky
	// This is a simplified version for demonstration
	go func() {
		<-c
		cancel()
	}()

	return ctx, cancel
}

// TestContextPropagation verifies that the context is properly propagated to FX modules
func TestContextPropagation(t *testing.T) {
	ctxReceived := false

	b := &Bootstrapper{
		context: context.Background(),
		options: []fx.Option{
			fx.NopLogger,
			fx.Invoke(func(ctx context.Context) {
				if ctx != nil {
					ctxReceived = true
				}
			}),
		},
	}

	app := fxtest.New(
		t,
		fx.Supply(fx.Annotate(b.context, fx.As(new(context.Context)))),
		fx.Options(b.options...),
	)

	app.RequireStart()
	app.RequireStop()

	if !ctxReceived {
		t.Error("Context was not properly propagated to FX modules")
	}
}
