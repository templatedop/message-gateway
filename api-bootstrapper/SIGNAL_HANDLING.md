# Signal Handling in API Bootstrapper

## Overview

The bootstrapper now includes graceful shutdown signal detection, allowing the application to respond to OS signals (SIGINT, SIGTERM) and perform a clean shutdown.

## Implementation

### Changes Made

**File:** `api-bootstrapper/bootstrapper.go`

#### 1. Added Signal Detection Imports
```go
import (
    "os/signal"
    "syscall"
    // ... other imports
)
```

#### 2. Modified `Run()` Method

The `Run()` method now wraps the incoming context with signal detection:

```go
func (b *Bootstrapper) Run(options ...fx.Option) {
    // Wrap the context with signal detection for graceful shutdown
    // Listen for SIGINT (Ctrl+C) and SIGTERM (kill command)
    ctx, cancel := signal.NotifyContext(b.context, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
    defer cancel()

    // Update the bootstrapper context with signal-aware context
    b.context = ctx

    // Monitor context cancellation in a separate goroutine
    go func() {
        <-ctx.Done()
        if ctx.Err() == context.Canceled {
            log.GetBaseLoggerInstance().ToZerolog().Info().Msg("Shutdown signal received, initiating graceful shutdown...")
        }
    }()

    // Create and run the FX application
    app := b.BootstrapApp(options...)

    // Run the application with signal handling
    // When a signal is received, the context will be cancelled and fx will gracefully shutdown
    app.Run()

    log.GetBaseLoggerInstance().ToZerolog().Info().Msg("Application shutdown complete")
}
```

## How It Works

### Signal Flow

1. **Main Application** sends `context.Background()` to bootstrapper:
   ```go
   // main.go
   app := bootstrapper.New().Options(
       bootstrap.Fxvalidator,
       bootstrap.FxHandler,
       bootstrap.FxRepo,
   ).WithContext(context.Background()).Run()
   ```

2. **Bootstrapper** wraps the context with signal detection:
   - Creates a new context using `signal.NotifyContext()`
   - Listens for `os.Interrupt`, `syscall.SIGTERM`, and `syscall.SIGINT`
   - Updates the bootstrapper's internal context

3. **Signal Received**:
   - OS signal triggers context cancellation
   - Monitoring goroutine logs the shutdown initiation
   - FX application receives shutdown signal via context
   - All FX lifecycle hooks execute their `OnStop` functions

4. **Graceful Shutdown**:
   - HTTP server closes connections (10s timeout)
   - Database connections are closed
   - Tracer provider flushes pending spans
   - All resources are cleaned up

### Signals Handled

| Signal | Source | Description |
|--------|--------|-------------|
| `os.Interrupt` | Ctrl+C in terminal | User interruption |
| `syscall.SIGTERM` | `kill <pid>` | Termination request |
| `syscall.SIGINT` | Ctrl+C | Interrupt signal |

## Usage in Modules

Any FX module can now access the signal-aware context:

```go
fx.Invoke(func(ctx context.Context) {
    // This context will be cancelled when shutdown signal is received
    go func() {
        <-ctx.Done()
        // Perform cleanup
    }()
})
```

## Benefits

### 1. Graceful Shutdown
- HTTP server completes in-flight requests
- Database transactions are committed or rolled back properly
- Telemetry data is flushed before exit

### 2. Resource Cleanup
- All database connections are closed
- File handles are released
- Network connections are properly terminated

### 3. Observability
- Logs indicate when shutdown is initiated
- Logs confirm when shutdown is complete
- Easy to debug shutdown issues

### 4. Container Orchestration Support
- Works seamlessly with Kubernetes pod termination
- Responds to Docker stop commands
- Compatible with systemd service management

## Testing

Test file: `api-bootstrapper/bootstrapper_signal_test.go`

### Running Tests

```bash
cd api-bootstrapper
go test -v -run TestSignalHandling
go test -v -run TestContextPropagation
```

## Deployment Scenarios

### Docker Container
```dockerfile
# The application will respond to docker stop gracefully
CMD ["./app"]
```

### Kubernetes Pod
```yaml
# The application will respond to pod termination gracefully
spec:
  terminationGracePeriodSeconds: 30
```

### Systemd Service
```ini
[Service]
Type=notify
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
```

## Troubleshooting

### Application doesn't shutdown
- Check if any goroutines are blocking
- Verify database connections are being closed
- Increase shutdown timeout in lifecycle hooks

### Shutdown takes too long
- Review `OnStop` lifecycle hooks
- Check for long-running operations
- Consider reducing timeout values

### Context not propagated
- Ensure modules request `context.Context` via FX dependency injection
- Verify `fx.Supply` is providing the context correctly

## Example Output

### Normal Shutdown
```
INFO Shutdown signal received, initiating graceful shutdown...
INFO Database shutdown complete!!
INFO Application shutdown complete
```

### During Development (Ctrl+C)
```
^CINFO Shutdown signal received, initiating graceful shutdown...
INFO Connection stats during shutdown: Total connections: 10
INFO Read Database shutdown complete!!
INFO Database shutdown complete!!
INFO Application shutdown complete
```

## Future Enhancements

1. **Configurable Shutdown Timeout**: Allow configuration of shutdown grace period
2. **Shutdown Hooks**: Add custom pre-shutdown and post-shutdown hooks
3. **Health Check Integration**: Update health checks during shutdown
4. **Metrics**: Track shutdown duration and success rate
