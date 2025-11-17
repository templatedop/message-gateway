# Migration Guide: Event-based Logging API

This guide helps you migrate from the deprecated `ToZerolog()` pattern to the new Event-based API for structured logging.

## Overview

The new Event-based API provides direct access to `*zerolog.Event` without requiring verbose calls to `GetBaseLoggerInstance().ToZerolog()`. This improves code readability and maintains better abstraction.

## Quick Comparison

### ❌ Old Way (Deprecated)
```go
log.GetBaseLoggerInstance().ToZerolog().Info().Str("user", "john").Msg("login successful")
log.GetBaseLoggerInstance().ToZerolog().Error().Err(err).Msg("database error")
```

### ✅ New Way (Recommended)
```go
log.InfoEvent(ctx).Str("user", "john").Msg("login successful")
log.ErrorEvent(ctx).Err(err).Msg("database error")
```

## API Functions

### Simple Logging (Existing - No Changes)
For simple messages without structured fields, continue using the existing API:

```go
log.Debug(ctx, "processing item %d", itemID)
log.Info(ctx, "server started on port 8080")
log.Warn(ctx, "cache miss for key: %s", key)
log.Error(ctx, "failed to connect: %v", err)
log.Critical(ctx, "database connection lost")
log.Fatal(ctx, "cannot bind to port") // panics after logging
```

### Structured Logging (New Event API)
For structured logging with fields, use the new Event functions:

#### DebugEvent
```go
log.DebugEvent(ctx).
    Str("query", sql).
    Dur("duration", elapsed).
    Int("rows", count).
    Msg("query executed")
```

#### InfoEvent
```go
log.InfoEvent(ctx).
    Str("user_id", userID).
    Str("action", "purchase").
    Float64("amount", 99.99).
    Msg("transaction completed")
```

#### WarnEvent
```go
log.WarnEvent(ctx).
    Str("reason", "rate_limit").
    Int("attempts", attempts).
    Str("ip", clientIP).
    Msg("rate limit approaching")
```

#### ErrorEvent
```go
log.ErrorEvent(ctx).
    Err(err).
    Str("operation", "db_query").
    Str("table", "users").
    Msg("database operation failed")
```

#### CriticalEvent
```go
log.CriticalEvent(ctx).
    Err(err).
    Str("service", "payment-gateway").
    Bool("available", false).
    Msg("critical service unavailable")
```

## Migration Examples

### Example 1: Basic Error Logging

**Before:**
```go
log.GetBaseLoggerInstance().ToZerolog().Error().Err(err).Msg("database ping error")
```

**After:**
```go
log.ErrorEvent(ctx).Err(err).Msg("database ping error")
```

### Example 2: Info with Structured Fields

**Before:**
```go
log.GetBaseLoggerInstance().ToZerolog().Info().
    Str("module", "DBModule").
    Msg("Starting database module")
```

**After:**
```go
log.InfoEvent(ctx).
    Str("module", "DBModule").
    Msg("Starting database module")
```

### Example 3: Error with Multiple Fields

**Before:**
```go
logger := log.GetBaseLoggerInstance().ToZerolog()
logger.Error().
    Err(err).
    Str("bucket", bucketName).
    Str("operation", "create").
    Msg("bucket operation failed")
```

**After:**
```go
log.ErrorEvent(ctx).
    Err(err).
    Str("bucket", bucketName).
    Str("operation", "create").
    Msg("bucket operation failed")
```

### Example 4: Debug Logging

**Before:**
```go
log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("Bucket found")
```

**After:**
```go
log.DebugEvent(ctx).Msg("Bucket found")
```

## Context-Aware Logging

The Event API is context-aware and will automatically use the logger embedded in the context (if available):

```go
// In a Gin handler
func HandleRequest(c *gin.Context) {
    // Logger automatically includes request metadata (request-id, trace-id, etc.)
    log.InfoEvent(c).
        Str("endpoint", "/api/users").
        Int("user_count", count).
        Msg("fetched users")
}

// In a background goroutine
func ProcessJob(ctx context.Context, jobID string) {
    log.InfoEvent(ctx).
        Str("job_id", jobID).
        Msg("processing job")
}

// Without context (uses base logger)
func Startup() {
    log.InfoEvent(nil).
        Str("version", "v1.0.0").
        Msg("application starting")
}
```

## Available Field Types

The Event API supports all zerolog field types:

```go
log.InfoEvent(ctx).
    Str("string_field", "value").           // string
    Int("int_field", 42).                   // int
    Int64("int64_field", 1234567890).       // int64
    Float64("float_field", 3.14).           // float64
    Bool("bool_field", true).               // bool
    Err(err).                               // error
    Dur("duration", time.Second*5).         // time.Duration
    Time("timestamp", time.Now()).          // time.Time
    Interface("complex", customStruct).     // any type (JSON serialized)
    Msg("message")
```

## Backward Compatibility

The old API remains functional but is deprecated:

### Deprecated (still works)
```go
logger := log.GetBaseLoggerInstance().ToZerolog()
logger.Info().Msg("message")
```

### Preferred
```go
log.InfoEvent(ctx).Msg("message")
```

## Migration Checklist

- [ ] Replace `log.GetBaseLoggerInstance().ToZerolog().Info()` with `log.InfoEvent(ctx)`
- [ ] Replace `log.GetBaseLoggerInstance().ToZerolog().Error()` with `log.ErrorEvent(ctx)`
- [ ] Replace `log.GetBaseLoggerInstance().ToZerolog().Debug()` with `log.DebugEvent(ctx)`
- [ ] Replace `log.GetBaseLoggerInstance().ToZerolog().Warn()` with `log.WarnEvent(ctx)`
- [ ] Add context parameter where applicable
- [ ] Remove local `logger` variables that store `ToZerolog()` results
- [ ] Update tests to use new API

## Benefits

1. **Cleaner Code**: Less verbose, more readable
2. **Better Abstraction**: Hides zerolog implementation details
3. **Context-Aware**: Automatically includes request metadata
4. **Consistent API**: Matches the simple logging API pattern
5. **Easier Testing**: Mock at the package level, not zerolog level

## Need Help?

- See `api-log/logger_test.go` for comprehensive examples
- Check existing usages with: `grep -r "\.ToZerolog()" .`
- Review the API documentation in `api-log/logger.go`
