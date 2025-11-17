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
go test -v
```

### Test Coverage

The test suite includes:
- `TestSignalHandling` - Verifies signal detection and graceful shutdown
- `TestContextPropagation` - Ensures context is passed to FX modules
- `TestWithContext` - Validates WithContext method
- `TestOptionsMethod` - Checks Options appending
- `TestBootstrapAppCreation` - Verifies FX app creation

## Deployment Scenarios

### Docker Container
```dockerfile
# The application will respond to docker stop gracefully
CMD ["./app"]
```

When you run `docker stop <container>`:
1. Docker sends SIGTERM to the application
2. Application receives signal and initiates graceful shutdown
3. All resources are cleaned up within the grace period (default 10s)
4. If timeout expires, Docker sends SIGKILL

### Kubernetes Pod
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: message-gateway
spec:
  containers:
  - name: app
    image: message-gateway:latest
  terminationGracePeriodSeconds: 30  # Allow 30s for graceful shutdown
```

When pod is terminated:
1. Kubernetes sends SIGTERM to the main process
2. Application initiates graceful shutdown
3. Waits up to `terminationGracePeriodSeconds`
4. If not terminated, Kubernetes sends SIGKILL

### Systemd Service
```ini
[Unit]
Description=Message Gateway Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/message-gateway
Restart=on-failure
RestartSec=5s

# Shutdown configuration
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
SendSIGKILL=yes

[Install]
WantedBy=multi-user.target
```

### Manual Execution
```bash
# Start the application
./message-gateway

# Graceful shutdown with Ctrl+C
^C

# Or send signal from another terminal
kill -TERM <pid>
```

## Example Lifecycle

### Startup
```
INFO Starting fxdb module
INFO Successfully connected to the database
INFO Starting HTTP server on :8080
```

### Normal Operation
```
INFO Request: GET /health
INFO Request: POST /api/v1/messages
```

### Shutdown (Ctrl+C)
```
^C
INFO Shutdown signal received, initiating graceful shutdown...
INFO Connection stats during shutdown: Total connections: 10
INFO Database shutdown complete!!
INFO Read Database shutdown complete!!
INFO Application shutdown complete
```

## Module Integration

### Accessing Context in Custom Modules

If you create custom FX modules that need to react to shutdown signals:

```go
var MyCustomModule = fx.Module(
    "my-module",
    fx.Invoke(func(lc fx.Lifecycle, ctx context.Context) {
        lc.Append(fx.Hook{
            OnStart: func(startCtx context.Context) error {
                // Start your service
                go func() {
                    // Monitor for shutdown
                    <-ctx.Done()
                    log.Info("Shutdown signal received in my module")
                    // Perform cleanup
                }()
                return nil
            },
            OnStop: func(stopCtx context.Context) error {
                // Additional cleanup if needed
                return nil
            },
        })
    }),
)
```

### Database Module Example

The existing database modules already use this pattern:

```go
fx.Hook{
    OnStart: func(ctx context.Context) error {
        log.Info("Starting database connection")
        return db.Ping()
    },
    OnStop: func(ctx context.Context) error {
        log.Info("Closing database connections")
        db.Close()
        return nil
    },
}
```

## Troubleshooting

### Application doesn't shutdown

**Symptoms:**
- Application hangs after receiving SIGTERM
- No shutdown logs appear

**Possible Causes:**
1. Goroutines blocking without checking context
2. Database queries without timeout
3. Network calls without context

**Solutions:**
```go
// Bad - no context awareness
go func() {
    for {
        doWork()
    }
}()

// Good - context-aware goroutine
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            doWork()
        }
    }
}()
```

### Shutdown takes too long

**Symptoms:**
- Shutdown completes but takes 10+ seconds
- Timeout errors in logs

**Possible Causes:**
1. Long-running operations in OnStop hooks
2. Database queries without timeout
3. HTTP clients without timeout

**Solutions:**
```go
// Add timeout to OnStop operations
OnStop: func(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    return service.Shutdown(ctx)
}
```

### Context not propagated

**Symptoms:**
- Modules don't receive context
- Cannot access context.Done()

**Solutions:**
```go
// Ensure you request context in FX dependency injection
fx.Invoke(func(ctx context.Context, db *db.DB) {
    // Now you have access to context
})
```

### Database connections leak

**Symptoms:**
- Database shows active connections after shutdown
- Connection pool not releasing

**Solutions:**
- Ensure `OnStop` hooks call `db.Close()`
- Check for goroutines holding connections
- Verify context is cancelled properly

## Performance Considerations

### Goroutine Management

The signal handler creates one monitoring goroutine:
```go
go func() {
    <-ctx.Done()
    log.Info("Shutdown signal received")
}()
```

This is lightweight and has minimal overhead.

### Context Propagation

Using `signal.NotifyContext` is efficient:
- No polling required
- OS signals are delivered directly
- Context cancellation propagates instantly

### Memory Impact

Signal handling adds:
- One goroutine (~2KB stack)
- One context wrapper (~48 bytes)
- Signal channel (~24 bytes)

Total overhead: ~2KB (negligible)

## Best Practices

### 1. Always Use Context in Long-Running Operations
```go
// Good
result, err := db.QueryContext(ctx, query)

// Bad
result, err := db.Query(query)
```

### 2. Set Timeouts for Cleanup Operations
```go
OnStop: func(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return cleanup(ctx)
}
```

### 3. Monitor Context Cancellation
```go
go func() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            doPeriodicWork()
        }
    }
}()
```

### 4. Log Shutdown Progress
```go
OnStop: func(ctx context.Context) error {
    log.Info("Starting service shutdown")
    err := service.Close()
    if err != nil {
        log.Error("Shutdown error", err)
        return err
    }
    log.Info("Service shutdown complete")
    return nil
}
```

## Future Enhancements

1. **Configurable Shutdown Timeout**: Allow configuration of shutdown grace period
2. **Shutdown Hooks**: Add custom pre-shutdown and post-shutdown hooks
3. **Health Check Integration**: Update health checks during shutdown to return 503
4. **Metrics**: Track shutdown duration and success rate
5. **Custom Signal Handlers**: Support for custom signal handling per module
6. **Graceful Degradation**: Partial shutdown support for non-critical services

## References

- [Go signal package](https://pkg.go.dev/os/signal)
- [Uber FX Lifecycle](https://uber-go.github.io/fx/lifecycle.html)
- [Go Context](https://pkg.go.dev/context)
- [Kubernetes Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
