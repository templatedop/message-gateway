# Advanced Features Guide

This guide covers the advanced features added to the api-log package for production use:
- Configurable Middleware (Health Check Skipping)
- Performance Optimizations
- Sampling/Rate Limiting

## Table of Contents

- [Configurable Middleware](#configurable-middleware)
- [Log Sampling](#log-sampling)
- [Performance Optimizations](#performance-optimizations)
- [Real-World Examples](#real-world-examples)

---

## Configurable Middleware

Skip logging for specific endpoints like health checks, metrics, or internal routes.

### Basic Usage

```go
import (
    log "MgApplication/api-log"
    "github.com/gin-gonic/gin"
)

router := gin.New()

// Use default configuration (skips /healthz and /health)
router.Use(log.RequestResponseLoggerMiddleware)

// Or use custom configuration
config := &log.MiddlewareConfig{
    SkipPaths: []string{"/healthz", "/metrics", "/ready"},
    SkipPathPrefixes: []string{"/internal/", "/debug/"},
    SkipMethodPaths: map[string][]string{
        "GET": {"/status", "/ping"},
    },
}
router.Use(log.RequestResponseLoggerMiddlewareWithConfig(config))
```

### MiddlewareConfig Options

#### SkipPaths - Exact Path Matching
Skip logging for exact path matches:

```go
config := &log.MiddlewareConfig{
    SkipPaths: []string{
        "/healthz",
        "/health",
        "/metrics",
        "/ready",
        "/alive",
    },
}
```

#### SkipPathPrefixes - Prefix Matching
Skip all paths starting with specific prefixes:

```go
config := &log.MiddlewareConfig{
    SkipPathPrefixes: []string{
        "/internal/",    // Skips /internal/debug, /internal/metrics, etc.
        "/debug/",       // Skips /debug/pprof, /debug/vars, etc.
        "/_",            // Skips /_status, /_health, etc.
    },
}
```

#### SkipMethodPaths - Method-Specific Paths
Skip specific HTTP method + path combinations:

```go
config := &log.MiddlewareConfig{
    SkipMethodPaths: map[string][]string{
        "GET": {"/status", "/ping"},      // Skip GET /status and GET /ping
        "POST": {"/webhook"},              // Skip POST /webhook
        "HEAD": {"/healthz", "/ready"},    // Skip HEAD /healthz and HEAD /ready
    },
}
```

#### Combined Configuration
Use all options together:

```go
config := &log.MiddlewareConfig{
    SkipPaths:        []string{"/healthz"},
    SkipPathPrefixes: []string{"/internal/", "/debug/"},
    SkipMethodPaths: map[string][]string{
        "GET": {"/metrics"},
    },
}
```

### Default Configuration

When using `RequestResponseLoggerMiddleware()` or passing `nil` to `RequestResponseLoggerMiddlewareWithConfig()`, the default configuration is used:

```go
// Default skips these paths
DefaultMiddlewareConfig() = &MiddlewareConfig{
    SkipPaths: []string{"/healthz", "/health"},
    SkipPathPrefixes: []string{},
    SkipMethodPaths: make(map[string][]string),
}
```

---

## Log Sampling

Reduce log volume by sampling logs based on level, tags, or global rates.

### Why Use Sampling?

- **High-traffic services**: Log 1 in 100 requests instead of all
- **Cost optimization**: Reduce log storage costs by 90%+
- **Performance**: Reduce I/O overhead from excessive logging
- **Debug logs in production**: Sample debug logs at 10% instead of disabling completely

### Basic Usage

```go
factory := log.NewDefaultLoggerFactory()

samplingConfig := &log.SamplingConfig{
    GlobalRate: 1.0, // 100% - no global sampling
    LevelRates: map[zerolog.Level]float64{
        zerolog.DebugLevel: 0.1,  // Sample 10% of debug logs
        zerolog.InfoLevel:  1.0,  // Keep 100% of info logs
    },
}

factory.Create(
    log.WithServiceName("my-service"),
    log.WithSampling(samplingConfig),
)
```

### SamplingConfig Options

#### GlobalRate - Sample All Logs
Apply a global sampling rate to all logs:

```go
samplingConfig := &log.SamplingConfig{
    GlobalRate: 0.25, // Sample 25% of ALL logs
}
```

#### LevelRates - Per-Level Sampling
Different sampling rates for each log level:

```go
samplingConfig := &log.SamplingConfig{
    GlobalRate: 1.0, // No global sampling
    LevelRates: map[zerolog.Level]float64{
        zerolog.TraceLevel: 0.01,  // 1% of trace logs
        zerolog.DebugLevel: 0.1,   // 10% of debug logs
        zerolog.InfoLevel:  0.5,   // 50% of info logs
        zerolog.WarnLevel:  1.0,   // 100% of warnings
        zerolog.ErrorLevel: 1.0,   // 100% of errors
        zerolog.FatalLevel: 1.0,   // 100% of fatal logs
    },
}
```

#### TagRates - Tag-Based Sampling
Sample logs based on tags:

```go
samplingConfig := &log.SamplingConfig{
    GlobalRate: 1.0,
    TagRates: map[string]float64{
        "database":           0.5,   // 50% of database logs
        "cache":              0.1,   // 10% of cache logs
        "expensive-operation": 0.05, // 5% of expensive operations
    },
}

// Usage with tags
ctx := log.WithTags(context.Background(), "database")
log.Info(ctx, "query executed") // 50% chance of being logged
```

#### DisabledLevels - Completely Disable Levels
Turn off specific log levels entirely:

```go
samplingConfig := &log.SamplingConfig{
    DisabledLevels: []zerolog.Level{
        zerolog.TraceLevel,
        zerolog.DebugLevel,
    },
}

// All trace and debug logs are skipped, regardless of other settings
```

### Sampling Examples

#### Example 1: Production Environment
```go
// Production: Disable debug, sample info at 20%
samplingConfig := &log.SamplingConfig{
    GlobalRate: 1.0,
    LevelRates: map[zerolog.Level]float64{
        zerolog.InfoLevel: 0.2, // Only 20% of info logs
    },
    DisabledLevels: []zerolog.Level{
        zerolog.DebugLevel,
        zerolog.TraceLevel,
    },
}
```

#### Example 2: High-Traffic Endpoints
```go
// Sample expensive operations heavily
samplingConfig := &log.SamplingConfig{
    GlobalRate: 1.0,
    TagRates: map[string]float64{
        "api":      0.1,  // 10% of API logs
        "database": 0.05, // 5% of database logs
    },
}

ctx := log.WithTags(context.Background(), "api", "database")
log.InfoWithFields(ctx, "query executed", map[string]interface{}{
    "duration": duration,
    "rows": count,
}) // Only 5% chance of being logged (lowest tag rate wins)
```

#### Example 3: Development vs Production
```go
var samplingConfig *log.SamplingConfig

if os.Getenv("ENV") == "production" {
    samplingConfig = &log.SamplingConfig{
        GlobalRate: 0.5, // Sample 50% globally
        DisabledLevels: []zerolog.Level{zerolog.DebugLevel},
    }
} else {
    samplingConfig = nil // No sampling in dev
}

factory.Create(
    log.WithServiceName("my-service"),
    log.WithSampling(samplingConfig),
)
```

### How Sampling Works

Sampling decisions are made when creating a log event. The check order is:

1. **DisabledLevels**: If level is disabled, skip immediately
2. **GlobalRate**: Apply global sampling rate (if < 1.0)
3. **LevelRates**: Apply level-specific rate (if configured)
4. **TagRates**: Apply tag-specific rates (if tags present)

Each check is independent and multiplicative. For example:
- GlobalRate: 0.8 (80%)
- LevelRate (Debug): 0.5 (50%)
- TagRate ("expensive"): 0.2 (20%)
- **Final probability**: 0.8 × 0.5 × 0.2 = 0.08 (8%)

### Sampling Works Across All APIs

Sampling applies uniformly to all logging methods:

```go
// Simple API
log.Debug(ctx, "debug message")

// WithFields API
log.InfoWithFields(ctx, "user created", fields)

// Event API
log.InfoEvent(ctx).Str("key", "val").Msg("event")

// All use the same sampling logic
```

### Deterministic Testing

For testing, provide a custom random source:

```go
import "math/rand"

samplingConfig := &log.SamplingConfig{
    GlobalRate: 0.5,
    Rand: rand.New(rand.NewSource(12345)), // Deterministic
}

// Tests will have predictable sampling behavior
```

---

## Performance Optimizations

### String Builder (Automatic)

The middleware automatically uses `strings.Builder` for efficient path concatenation:

```go
// Old approach (2 allocations)
fullPath := path + "?" + raw

// New approach (1 allocation, pre-allocated)
var pathBuilder strings.Builder
pathBuilder.Grow(len(path) + len(raw) + 1)
pathBuilder.WriteString(path)
if raw != "" {
    pathBuilder.WriteByte('?')
    pathBuilder.WriteString(raw)
}
fullPath := pathBuilder.String()
```

**Performance Improvement**: ~6% faster for paths with query strings

### Pre-allocated Slices (Automatic)

Tags are accumulated with pre-allocated slices:

```go
// Old approach (potential slice growth)
allTags := append(existingTags, tags...)

// New approach (pre-allocated)
allTags := make([]string, 0, len(existingTags)+len(tags))
allTags = append(allTags, existingTags...)
allTags = append(allTags, tags...)
```

**Performance Improvement**: Predictable allocations, no growth overhead

### Benchmarks

Run benchmarks to see performance improvements:

```bash
go test ./api-log/... -bench=BenchmarkStringConcatenation -benchmem
go test ./api-log/... -bench=BenchmarkWithTags -benchmem
```

---

## Real-World Examples

### Example 1: Microservice with Health Checks

```go
package main

import (
    "github.com/gin-gonic/gin"
    log "MgApplication/api-log"
    "github.com/rs/zerolog"
)

func main() {
    // Initialize logger with sampling
    factory := log.NewDefaultLoggerFactory()

    samplingConfig := &log.SamplingConfig{
        GlobalRate: 1.0,
        LevelRates: map[zerolog.Level]float64{
            zerolog.DebugLevel: 0.1, // 10% of debug logs
        },
    }

    factory.Create(
        log.WithServiceName("user-service"),
        log.WithLevel(zerolog.DebugLevel),
        log.WithSampling(samplingConfig),
    )

    // Setup router with middleware config
    router := gin.New()

    middlewareConfig := &log.MiddlewareConfig{
        SkipPaths: []string{"/healthz", "/metrics", "/ready"},
        SkipPathPrefixes: []string{"/internal/"},
    }

    router.Use(log.SetCtxLoggerMiddleware)
    router.Use(log.RequestResponseLoggerMiddlewareWithConfig(middlewareConfig))

    // Routes
    router.GET("/healthz", healthCheck)
    router.GET("/ready", readinessCheck)
    router.GET("/api/users", getUsers)

    router.Run(":8080")
}
```

### Example 2: High-Traffic API with Tag-Based Sampling

```go
func handleRequest(c *gin.Context) {
    ctx := c.Request.Context()

    // Add tags based on request characteristics
    if isExpensiveOperation(c) {
        ctx = log.WithTags(ctx, "expensive")
    }

    if requiresDatabase(c) {
        ctx = log.WithTags(ctx, "database")
    }

    // These logs will be sampled based on tags
    log.InfoWithFields(ctx, "processing request", map[string]interface{}{
        "user_id": getUserID(c),
        "endpoint": c.Request.URL.Path,
    })

    // Do work...

    log.Info(ctx, "request completed")
}

// Sampling configuration
samplingConfig := &log.SamplingConfig{
    GlobalRate: 1.0,
    TagRates: map[string]float64{
        "expensive": 0.1,  // Only log 10% of expensive operations
        "database":  0.5,  // Log 50% of database operations
    },
}
```

### Example 3: Environment-Specific Configuration

```go
func createLogger(env string) {
    factory := log.NewDefaultLoggerFactory()

    var samplingConfig *log.SamplingConfig
    var middlewareConfig *log.MiddlewareConfig

    switch env {
    case "production":
        samplingConfig = &log.SamplingConfig{
            GlobalRate: 0.5, // Sample 50% of all logs
            DisabledLevels: []zerolog.Level{zerolog.DebugLevel},
        }
        middlewareConfig = &log.MiddlewareConfig{
            SkipPaths: []string{"/healthz", "/metrics", "/ready"},
            SkipPathPrefixes: []string{"/internal/", "/debug/"},
        }

    case "staging":
        samplingConfig = &log.SamplingConfig{
            LevelRates: map[zerolog.Level]float64{
                zerolog.DebugLevel: 0.2, // 20% of debug logs
            },
        }
        middlewareConfig = &log.MiddlewareConfig{
            SkipPaths: []string{"/healthz"},
        }

    case "development":
        samplingConfig = nil // No sampling
        middlewareConfig = nil // Use defaults
    }

    factory.Create(
        log.WithServiceName("my-service"),
        log.WithLevel(zerolog.DebugLevel),
        log.WithSampling(samplingConfig),
    )

    // Use middleware config in router setup
    if middlewareConfig != nil {
        router.Use(log.RequestResponseLoggerMiddlewareWithConfig(middlewareConfig))
    } else {
        router.Use(log.RequestResponseLoggerMiddleware)
    }
}
```

### Example 4: Cost Optimization for High-Volume Services

```go
// For services with millions of requests per day
samplingConfig := &log.SamplingConfig{
    GlobalRate: 1.0,
    LevelRates: map[zerolog.Level]float64{
        zerolog.InfoLevel: 0.01, // Only 1% of info logs (99% reduction!)
    },
    TagRates: map[string]float64{
        "critical": 1.0,     // Always log critical operations
        "error":    1.0,     // Always log errors
        "normal":   0.001,   // 0.1% of normal operations
    },
}

// Tag your logs appropriately
func processOrder(ctx context.Context, order Order) {
    if order.Value > 10000 {
        ctx = log.WithTags(ctx, "critical")
    } else {
        ctx = log.WithTags(ctx, "normal")
    }

    log.InfoWithFields(ctx, "processing order", map[string]interface{}{
        "order_id": order.ID,
        "value": order.Value,
    })

    // High-value orders are always logged
    // Normal orders: only 0.1% are logged
}
```

---

## Migration Guide

### Migrating to Configurable Middleware

**Before:**
```go
router.Use(log.RequestResponseLoggerMiddleware)
// Hardcoded to skip only /healthz
```

**After:**
```go
// Option 1: Use defaults (now skips /healthz AND /health)
router.Use(log.RequestResponseLoggerMiddleware)

// Option 2: Custom configuration
config := &log.MiddlewareConfig{
    SkipPaths: []string{"/healthz", "/metrics"},
}
router.Use(log.RequestResponseLoggerMiddlewareWithConfig(config))
```

### Adding Sampling to Existing Logger

**Before:**
```go
factory.Create(
    log.WithServiceName("my-service"),
    log.WithLevel(zerolog.InfoLevel),
)
```

**After:**
```go
samplingConfig := &log.SamplingConfig{
    LevelRates: map[zerolog.Level]float64{
        zerolog.DebugLevel: 0.1,
    },
}

factory.Create(
    log.WithServiceName("my-service"),
    log.WithLevel(zerolog.InfoLevel),
    log.WithSampling(samplingConfig), // Add this line
)
```

---

## Best Practices

### Middleware Configuration
1. **Always skip health checks** to reduce log noise
2. **Use prefixes for internal routes** (e.g., `/internal/`, `/debug/`)
3. **Be specific with method+path** combinations when needed
4. **Document your skip rules** in configuration files

### Sampling
1. **Start conservative** (e.g., 50% sampling) and adjust based on volume
2. **Never sample errors/warnings** unless absolutely necessary
3. **Use tags for fine-grained control** instead of global sampling
4. **Monitor your sampling rates** to ensure you're not missing critical issues
5. **Test sampling behavior** in staging before production

### Performance
1. **The optimizations are automatic** - no code changes needed
2. **Measure with benchmarks** if making custom modifications
3. **Profile your application** to identify bottlenecks

---

## Troubleshooting

### Logs Not Appearing

**Symptom**: Expected logs are missing

**Possible Causes**:
1. Sampling is too aggressive - check `SamplingConfig`
2. Path is being skipped - check `MiddlewareConfig`
3. Log level is disabled - check `DisabledLevels`

**Solution**:
```go
// Temporarily disable sampling to debug
factory.Create(
    log.WithServiceName("my-service"),
    log.WithSampling(nil), // No sampling
)
```

### Too Many Logs Despite Sampling

**Symptom**: Still seeing high log volume

**Possible Causes**:
1. Sampling rates are too high
2. Not all code paths use sampling
3. Tags not configured correctly

**Solution**:
```go
// Check your configuration
samplingConfig := &log.SamplingConfig{
    GlobalRate: 0.1, // Try more aggressive global sampling
    LevelRates: map[zerolog.Level]float64{
        zerolog.InfoLevel: 0.01, // Very aggressive
    },
}
```

### Middleware Not Skipping Paths

**Symptom**: Health checks are still being logged

**Possible Causes**:
1. Using old middleware function
2. Path doesn't match exactly
3. Method-specific configuration needed

**Solution**:
```go
// Ensure you're using the new function
router.Use(log.RequestResponseLoggerMiddlewareWithConfig(config))

// Check exact path matching
config := &log.MiddlewareConfig{
    SkipPaths: []string{"/healthz"}, // Must match exactly
}

// Or use prefixes for flexibility
config := &log.MiddlewareConfig{
    SkipPathPrefixes: []string{"/health"}, // Matches /health, /healthz, /health/check
}
```

---

## API Reference

### MiddlewareConfig

```go
type MiddlewareConfig struct {
    SkipPaths        []string              // Exact paths to skip
    SkipPathPrefixes []string              // Path prefixes to skip
    SkipMethodPaths  map[string][]string   // Method-specific paths to skip
}

func DefaultMiddlewareConfig() *MiddlewareConfig
func (c *MiddlewareConfig) ShouldSkip(method, path string) bool
```

### SamplingConfig

```go
type SamplingConfig struct {
    GlobalRate     float64                       // Global sampling rate (0.0-1.0)
    LevelRates     map[zerolog.Level]float64    // Per-level rates
    TagRates       map[string]float64           // Per-tag rates
    DisabledLevels []zerolog.Level              // Completely disabled levels
    Rand           *rand.Rand                   // Random source (for testing)
}

func DefaultSamplingConfig() *SamplingConfig
func (s *SamplingConfig) ShouldLog(level zerolog.Level, tags []string) bool
```

### Factory Options

```go
// New option for sampling
func WithSampling(config *SamplingConfig) loggerOption
```

### Middleware Functions

```go
// Existing (now uses DefaultMiddlewareConfig)
func RequestResponseLoggerMiddleware(c *gin.Context)

// New configurable version
func RequestResponseLoggerMiddlewareWithConfig(config *MiddlewareConfig) gin.HandlerFunc
```

---

## Further Reading

- [Tags and Structured Fields Guide](TAGS_AND_FIELDS_GUIDE.md)
- [Migration Guide](MIGRATION_GUIDE.md)
- [API Documentation](https://pkg.go.dev/github.com/rs/zerolog)
