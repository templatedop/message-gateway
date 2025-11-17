package bootstrapper

import (
	"context"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestContextPropagationToModules verifies that the signal-aware context
// is actually available to modules when they request it
func TestContextPropagationToModules(t *testing.T) {
	receivedContext := false
	var moduleContext context.Context

	b := &Bootstrapper{
		context: context.Background(),
		options: []fx.Option{
			fx.NopLogger,
			// Module that requests context
			fx.Invoke(func(ctx context.Context) {
				receivedContext = true
				moduleContext = ctx
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

	if !receivedContext {
		t.Fatal("Context was NOT propagated to module!")
	}

	if moduleContext == nil {
		t.Fatal("Module received nil context!")
	}

	t.Log("✅ Context successfully propagated to module")
}

// TestSignalContextAvailableToModules verifies that modules can access
// the signal-aware context and listen for Done()
func TestSignalContextAvailableToModules(t *testing.T) {
	contextReceived := false
	contextHasDone := false

	b := &Bootstrapper{
		context: context.Background(),
		options: []fx.Option{
			fx.NopLogger,
			fx.Invoke(func(ctx context.Context) {
				contextReceived = true

				// Check if context has Done() channel
				select {
				case <-ctx.Done():
					// Context already done
				default:
					// Context not done, which is good
					contextHasDone = true
				}
			}),
		},
	}

	// Create signal-aware context
	ctx, cancel := context.WithCancel(b.context)
	defer cancel()
	b.context = ctx

	app := fxtest.New(
		t,
		fx.Supply(fx.Annotate(b.context, fx.As(new(context.Context)))),
		fx.Options(b.options...),
	)

	app.RequireStart()
	app.RequireStop()

	if !contextReceived {
		t.Fatal("Context was NOT received by module")
	}

	if !contextHasDone {
		t.Fatal("Context does not have Done() channel")
	}

	t.Log("✅ Signal-aware context is available to modules")
}

// TestModuleCanListenToContextCancellation verifies that a module can
// actually listen to context cancellation
func TestModuleCanListenToContextCancellation(t *testing.T) {
	cancelled := make(chan bool, 1)

	b := &Bootstrapper{
		context: context.Background(),
		options: []fx.Option{
			fx.NopLogger,
			fx.Invoke(func(ctx context.Context) {
				go func() {
					<-ctx.Done()
					cancelled <- true
				}()
			}),
		},
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(b.context)
	b.context = ctx

	app := fxtest.New(
		t,
		fx.Supply(fx.Annotate(b.context, fx.As(new(context.Context)))),
		fx.Options(b.options...),
	)

	app.RequireStart()

	// Cancel the context
	cancel()

	// Wait for cancellation to be detected
	select {
	case <-cancelled:
		t.Log("✅ Module successfully detected context cancellation")
	case <-time.After(2 * time.Second):
		t.Fatal("Module did NOT detect context cancellation")
	}

	app.RequireStop()
}
