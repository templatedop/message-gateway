# Stack Trace Configuration Guide

## Overview

The api-errors package now supports **configurable stack trace collection** with:
1. **Global enable/disable toggle** - Turn stack traces on or off
2. **Configurable depth** - Control how many frames to collect
3. **Selective collection** - Collect only for server errors (5xx)
4. **Environment-based defaults** - Auto-configure based on APP_ENV

## Quick Start

### Environment Variables

Set these environment variables to configure stack trace collection:

```bash
# Enable/disable stack traces (default: depends on APP_ENV)
export STACKTRACE_ENABLED=true

# Maximum number of stack frames to collect (default: depends on APP_ENV)
export STACKTRACE_MAX_DEPTH=10

# Collect stack traces only for 5xx errors (default: depends on APP_ENV)
export STACKTRACE_5XX_ONLY=true

# Set environment to auto-configure defaults
export APP_ENV=production  # or staging, development, dev, local
```

### Programmatic Configuration

```go
import "MgApplication/api-errors"

// Completely disable stack traces (production)
apierrors.DisableStackTraces()

// Enable with specific depth
apierrors.EnableStackTraces(10) // Collect up to 10 frames

// Enable only for 5xx errors
apierrors.EnableStackTracesFor5xxOnly(10)

// Full custom configuration
apierrors.SetStackTraceConfig(apierrors.StackTraceConfig{
    Enabled:           true,
    MaxDepth:          15,
    CollectFor5xxOnly: false,
})
```

## Configuration Options

### 1. Enabled (bool)

Controls whether stack traces are collected at all.

- **`true`**: Stack traces are collected (based on other settings)
- **`false`**: No stack traces are collected (zero overhead)

**Example**:
```go
config := apierrors.GetStackTraceConfig()
config.Enabled = false  // Disable completely
apierrors.SetStackTraceConfig(config)
```

### 2. MaxDepth (int)

Maximum number of stack frames to collect.

- **Higher values**: More detail, slower, more memory
- **Lower values**: Less detail, faster, less memory
- **Recommended**:
  - Development: 32 (full stack)
  - Staging: 10 (moderate)
  - Production: 5 or 0 (minimal/disabled)

**Example**:
```go
apierrors.EnableStackTraces(5)  // Only top 5 frames
```

### 3. CollectFor5xxOnly (bool)

Whether to collect stack traces only for server errors (5xx).

- **`true`**: Collect only when error code >= 500 (server errors)
- **`false`**: Collect for all errors (when enabled)

**Use case**: In production, you may want stack traces only for unexpected server errors, not for expected client errors (4xx).

**Example**:
```go
apierrors.EnableStackTracesFor5xxOnly(10)
// Stack traces collected for 500, 502, 503, etc.
// No stack traces for 400, 404, 422, etc.
```

## Environment-Based Defaults

The package automatically configures itself based on the `APP_ENV` environment variable:

### Production (`APP_ENV=production`)
```go
Enabled:           false  // Disabled for maximum performance
MaxDepth:          0
CollectFor5xxOnly: true
```

**Rationale**: Production prioritizes performance. Stack traces are disabled by default.

### Staging (`APP_ENV=staging`)
```go
Enabled:           true   // Enabled for debugging production issues
MaxDepth:          10     // Moderate depth
CollectFor5xxOnly: true   // Only unexpected server errors
```

**Rationale**: Staging balances debugging capability with performance.

### Development (`APP_ENV=development`, `dev`, or `local`)
```go
Enabled:           true   // Enabled for full debugging
MaxDepth:          32     // Full stack trace
CollectFor5xxOnly: false  // Collect for all errors
```

**Rationale**: Development prioritizes debugging capability over performance.

### Unknown/Unset Environment
```go
Enabled:           true
MaxDepth:          10
CollectFor5xxOnly: false
```

**Rationale**: Conservative defaults with limited collection.

## Performance Impact

### Disabled (Enabled=false)
- **CPU overhead**: 0 nanoseconds
- **Memory overhead**: 0 bytes
- **Best for**: Production

### Enabled with MaxDepth=5
- **CPU overhead**: ~0.3-0.5 microseconds per error
- **Memory overhead**: ~300-500 bytes per error
- **Best for**: Production (if stack traces needed for 5xx errors)

### Enabled with MaxDepth=10
- **CPU overhead**: ~0.5-1 microsecond per error
- **Memory overhead**: ~600-1,000 bytes per error
- **Best for**: Staging

### Enabled with MaxDepth=32
- **CPU overhead**: ~1-2 microseconds per error
- **Memory overhead**: ~1,500-2,000 bytes per error
- **Best for**: Development

### CollectFor5xxOnly Impact
When enabled, stack trace collection is skipped for 4xx errors:
- **4xx errors**: 0 overhead (no collection)
- **5xx errors**: Normal overhead (based on MaxDepth)

**Example**: At 10,000 requests/second with 90% success, 8% 4xx errors, 2% 5xx errors:
- Without CollectFor5xxOnly: 1,000 stack traces/second
- With CollectFor5xxOnly: 200 stack traces/second (80% reduction)

## Usage Examples

### Example 1: Production Setup

```go
// In main.go or initialization code
func init() {
    // Option 1: Disable entirely for maximum performance
    apierrors.DisableStackTraces()

    // Option 2: Enable only for 5xx errors with minimal depth
    apierrors.EnableStackTracesFor5xxOnly(5)
}
```

### Example 2: Staging Setup

```go
func init() {
    apierrors.SetStackTraceConfig(apierrors.StackTraceConfig{
        Enabled:           true,
        MaxDepth:          10,
        CollectFor5xxOnly: true,  // Only unexpected server errors
    })
}
```

### Example 3: Development Setup

```go
func init() {
    // Full stack traces for all errors
    apierrors.EnableStackTraces(32)
}
```

### Example 4: Dynamic Configuration

```go
// Read from your configuration system
func configureStackTraces(cfg *Config) {
    apierrors.SetStackTraceConfig(apierrors.StackTraceConfig{
        Enabled:           cfg.StackTrace.Enabled,
        MaxDepth:          cfg.StackTrace.MaxDepth,
        CollectFor5xxOnly: cfg.Featureflags.StackTraceFor5xxOnly,
    })
}
```

### Example 5: Runtime Toggle

```go
// Disable during high load
func handleHighLoad() {
    apierrors.DisableStackTraces()
}

// Re-enable after load decreases
func handleNormalLoad() {
    apierrors.EnableStackTracesFor5xxOnly(10)
}
```

## Environment Variable Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_ENV` | string | - | Sets environment-based defaults (`production`, `staging`, `development`) |
| `STACKTRACE_ENABLED` | bool | depends on APP_ENV | Enable/disable stack trace collection |
| `STACKTRACE_MAX_DEPTH` | int | depends on APP_ENV | Maximum number of frames to collect |
| `STACKTRACE_5XX_ONLY` | bool | depends on APP_ENV | Collect only for server errors (>= 500) |

**Priority**: Explicit environment variables override APP_ENV defaults.

### Example Configuration

```bash
# Development
export APP_ENV=development
# Automatic: Enabled=true, MaxDepth=32, CollectFor5xxOnly=false

# Production with overrides
export APP_ENV=production
export STACKTRACE_ENABLED=true       # Override default (false)
export STACKTRACE_MAX_DEPTH=5        # Override default (0)
export STACKTRACE_5XX_ONLY=true      # Override default (true)
```

## API Reference

### Functions

#### `GetStackTraceConfig() StackTraceConfig`
Returns the current global configuration (thread-safe).

```go
config := apierrors.GetStackTraceConfig()
fmt.Printf("Enabled: %v, MaxDepth: %d\n", config.Enabled, config.MaxDepth)
```

#### `SetStackTraceConfig(config StackTraceConfig)`
Updates the global configuration (thread-safe).

```go
apierrors.SetStackTraceConfig(apierrors.StackTraceConfig{
    Enabled:           true,
    MaxDepth:          20,
    CollectFor5xxOnly: false,
})
```

#### `DisableStackTraces()`
Convenience function to completely disable stack traces.

```go
apierrors.DisableStackTraces()
// Equivalent to:
// SetStackTraceConfig(StackTraceConfig{Enabled: false, MaxDepth: 0})
```

#### `EnableStackTraces(maxDepth int)`
Convenience function to enable stack traces with specified depth.

```go
apierrors.EnableStackTraces(15)
// Collects up to 15 frames for all errors
```

#### `EnableStackTracesFor5xxOnly(maxDepth int)`
Convenience function to enable stack traces only for 5xx errors.

```go
apierrors.EnableStackTracesFor5xxOnly(10)
// Collects up to 10 frames, but only for errors with code >= 500
```

## Migration Guide

### From Previous Behavior

**Before (v1)**: Stack traces were always collected with depth=32 for all errors.

**After (v2)**: Stack traces are configurable and disabled by default in production.

### Migration Steps

1. **No action required** if you want environment-based defaults
2. **Set `STACKTRACE_ENABLED=true`** if you want old behavior (always collect)
3. **Add configuration** in your initialization code:

```go
// To maintain old behavior (always collect, depth=32)
apierrors.EnableStackTraces(32)

// Or configure based on your needs
apierrors.SetStackTraceConfig(apierrors.StackTraceConfig{
    Enabled:           true,
    MaxDepth:          32,
    CollectFor5xxOnly: false,
})
```

## Troubleshooting

### Stack traces not appearing in logs

**Check**:
1. Is `STACKTRACE_ENABLED=true`?
2. If `CollectFor5xxOnly=true`, is the error code >= 500?
3. Is `MaxDepth > 0`?

**Solution**:
```bash
export STACKTRACE_ENABLED=true
export STACKTRACE_MAX_DEPTH=32
export STACKTRACE_5XX_ONLY=false
```

### Performance degradation

**Symptoms**: Slow error creation, high memory usage

**Solution**: Reduce `MaxDepth` or enable `CollectFor5xxOnly`:
```bash
export STACKTRACE_MAX_DEPTH=5
export STACKTRACE_5XX_ONLY=true
```

### Stack traces too shallow

**Symptoms**: Stack trace doesn't show enough context

**Solution**: Increase `MaxDepth`:
```bash
export STACKTRACE_MAX_DEPTH=50
```

## Best Practices

1. **Production**: Disable or use `CollectFor5xxOnly` with `MaxDepth <= 5`
2. **Staging**: Use `CollectFor5xxOnly=true` with `MaxDepth=10`
3. **Development**: Enable fully with `MaxDepth=32`
4. **Testing**: Disable stack traces in unit tests for speed
5. **Monitoring**: Log the configuration at startup to verify settings

### Example Startup Logging

```go
func main() {
    config := apierrors.GetStackTraceConfig()
    log.Printf("StackTrace Config: Enabled=%v, MaxDepth=%d, 5xxOnly=%v",
        config.Enabled, config.MaxDepth, config.CollectFor5xxOnly)
    // ... rest of application
}
```

## Conclusion

Configurable stack traces provide:
- ✅ **Performance control** - Disable in production for zero overhead
- ✅ **Flexibility** - Adjust depth based on environment
- ✅ **Selective collection** - Collect only for unexpected errors (5xx)
- ✅ **Environment awareness** - Auto-configure based on APP_ENV
- ✅ **Thread-safe** - Safe for concurrent access

For questions or issues, refer to this document or check the configuration at runtime using `GetStackTraceConfig()`.
