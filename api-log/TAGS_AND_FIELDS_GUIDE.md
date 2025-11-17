# Tags and Structured Fields Guide

This guide explains how to use the Tags and WithFields features for enhanced structured logging.

## 📋 Table of Contents

- [Tags Support](#tags-support)
- [Structured Fields API](#structured-fields-api)
- [Combining Tags and Fields](#combining-tags-and-fields)
- [API Comparison](#api-comparison)
- [Best Practices](#best-practices)
- [Examples](#examples)

---

## 🏷️ Tags Support

Tags allow you to label log entries with searchable keywords for better filtering and organization.

### Adding Tags to Context

```go
// Add tags to context
ctx = log.WithTags(ctx, "database", "payment")

// All logs with this context will include tags
log.Info(ctx, "processing transaction")
// Output: {"level":"info","tags":["database","payment"],"message":"processing transaction"}
```

### Tags Accumulate

```go
ctx = log.WithTags(ctx, "api", "v1")
ctx = log.WithTags(ctx, "auth")

// Tags accumulate: ["api", "v1", "auth"]
log.Info(ctx, "user authenticated")
```

### Retrieving Tags

```go
tags := log.GetTags(ctx)
if len(tags) > 0 {
    // Tags are present
}
```

### Use Cases for Tags

1. **Service/Module Identification**
   ```go
   ctx = log.WithTags(ctx, "payment-service", "stripe")
   ```

2. **Environment/Stage**
   ```go
   ctx = log.WithTags(ctx, "production", "us-west-2")
   ```

3. **Feature Flags**
   ```go
   ctx = log.WithTags(ctx, "feature-beta", "experiment-123")
   ```

4. **Request Classification**
   ```go
   ctx = log.WithTags(ctx, "api", "internal", "admin")
   ```

---

## 📊 Structured Fields API

The WithFields API provides a simple way to add structured data to logs using maps.

### Basic Usage

```go
log.InfoWithFields(ctx, "user logged in", map[string]interface{}{
    "user_id": "12345",
    "ip": "192.168.1.1",
    "success": true,
})
```

### All Log Levels

```go
// Debug
log.DebugWithFields(ctx, "cache lookup", map[string]interface{}{
    "key": "user:123",
    "hit": true,
})

// Info
log.InfoWithFields(ctx, "request processed", map[string]interface{}{
    "duration": 45 * time.Millisecond,
    "status": 200,
})

// Warn
log.WarnWithFields(ctx, "rate limit approaching", map[string]interface{}{
    "attempts": 4,
    "limit": 5,
})

// Error
log.ErrorWithFields(ctx, "database error", map[string]interface{}{
    "error": err,
    "table": "users",
})

// Critical
log.CriticalWithFields(ctx, "service down", map[string]interface{}{
    "service": "payment-gateway",
    "available": false,
})
```

### Supported Field Types

The WithFields API automatically handles type conversion:

```go
log.InfoWithFields(ctx, "all types", map[string]interface{}{
    // Strings
    "string": "value",

    // Integers
    "int": 42,
    "int64": int64(9876543210),
    "int32": int32(12345),

    // Unsigned integers
    "uint": uint(100),
    "uint64": uint64(9876543210),

    // Floats
    "float32": float32(3.14),
    "float64": 3.14159,

    // Boolean
    "bool": true,

    // Errors
    "error": err,

    // Arrays/Slices
    "strings": []string{"a", "b", "c"},
    "ints": []int{1, 2, 3},

    // Complex types (JSON marshaled)
    "custom": customStruct,

    // Nil values (skipped)
    "optional": nil,
})
```

---

## 🔗 Combining Tags and Fields

Tags and fields work seamlessly together:

```go
// Set up context with tags
ctx = log.WithTags(ctx, "database", "query")

// Log with both tags and fields
log.InfoWithFields(ctx, "query executed", map[string]interface{}{
    "table": "users",
    "rows": 150,
    "duration": 45 * time.Millisecond,
})

// Output includes both tags and fields:
// {
//   "level": "info",
//   "tags": ["database", "query"],
//   "table": "users",
//   "rows": 150,
//   "duration": 45,
//   "message": "query executed"
// }
```

### With Event API

Tags work with all logging APIs:

```go
ctx = log.WithTags(ctx, "payment")

// Event API
log.InfoEvent(ctx).
    Str("transaction_id", "tx-123").
    Float64("amount", 99.99).
    Msg("payment processed")
// Output includes tags automatically
```

---

## 🎯 API Comparison

### When to Use Each API

| API | Use Case | Example |
|-----|----------|---------|
| **Simple API** | Basic messages, no fields needed | `log.Info(ctx, "server started")` |
| **WithFields API** | Simple structured data from map | `log.InfoWithFields(ctx, "user created", fields)` |
| **Event API** | Complex structured data, many fields | `log.InfoEvent(ctx).Str(...).Int(...).Msg(...)` |

### Feature Comparison

| Feature | Simple | WithFields | Event |
|---------|--------|------------|-------|
| Tags support | ✅ | ✅ | ✅ |
| Simple messages | ✅ | ✅ | ✅ |
| Format strings | ✅ | ❌ | ❌ |
| Structured fields | ❌ | ✅ (map) | ✅ (chained) |
| Type safety | ❌ | ⚠️ (interface{}) | ✅ |
| Auto type conversion | N/A | ✅ | N/A |
| Flexibility | Low | Medium | High |

### Examples

```go
ctx = log.WithTags(ctx, "api")

// Simple API - basic message
log.Info(ctx, "request received from %s", clientIP)

// WithFields API - moderate complexity
log.InfoWithFields(ctx, "request processed", map[string]interface{}{
    "user_id": userID,
    "duration": elapsed,
    "status": 200,
})

// Event API - complex structured logging
log.InfoEvent(ctx).
    Str("user_id", userID).
    Dur("duration", elapsed).
    Int("status", 200).
    Str("endpoint", "/api/users").
    Int("response_size", size).
    Msg("request processed")
```

---

## ✅ Best Practices

### 1. Tag Naming

```go
// Good: lowercase, descriptive
ctx = log.WithTags(ctx, "database", "postgres", "read-replica")

// Avoid: uppercase, symbols
ctx = log.WithTags(ctx, "DATABASE", "db-postgres", "read_replica")
```

### 2. Tag Scope

```go
// Set broad tags early
func HandleRequest(ctx context.Context) {
    ctx = log.WithTags(ctx, "http", "api-v2")

    // Add specific tags as needed
    if isAdmin {
        ctx = log.WithTags(ctx, "admin")
    }

    processRequest(ctx)
}
```

### 3. Field Naming

```go
// Good: snake_case, descriptive
log.InfoWithFields(ctx, "user action", map[string]interface{}{
    "user_id": "123",
    "action_type": "login",
    "timestamp": time.Now(),
})

// Avoid: camelCase, abbreviations
log.InfoWithFields(ctx, "user action", map[string]interface{}{
    "userId": "123",
    "actType": "login",
    "ts": time.Now(),
})
```

### 4. Error Handling

```go
// Include error in fields
if err != nil {
    log.ErrorWithFields(ctx, "operation failed", map[string]interface{}{
        "error": err,
        "operation": "user_create",
        "user_id": userID,
    })
}
```

### 5. Performance Considerations

```go
// WithFields - good for dynamic data
fields := map[string]interface{}{
    "count": count,
    "total": total,
}
log.InfoWithFields(ctx, "batch processed", fields)

// Event API - better for static structure, slightly faster
log.InfoEvent(ctx).
    Int("count", count).
    Int("total", total).
    Msg("batch processed")
```

---

## 📝 Examples

### Example 1: HTTP Request Logging

```go
func LogHTTPRequest(ctx context.Context, r *http.Request, statusCode int, duration time.Duration) {
    ctx = log.WithTags(ctx, "http", "incoming")

    log.InfoWithFields(ctx, "http request", map[string]interface{}{
        "method": r.Method,
        "path": r.URL.Path,
        "status": statusCode,
        "duration": duration,
        "client_ip": r.RemoteAddr,
        "user_agent": r.UserAgent(),
    })
}
```

### Example 2: Database Operations

```go
func LogDatabaseQuery(ctx context.Context, query string, duration time.Duration, rowCount int) {
    ctx = log.WithTags(ctx, "database", "postgres")

    log.DebugWithFields(ctx, "query executed", map[string]interface{}{
        "query": query,
        "duration": duration,
        "rows": rowCount,
        "slow_query": duration > 1*time.Second,
    })
}
```

### Example 3: Payment Processing

```go
func ProcessPayment(ctx context.Context, paymentID string, amount float64) error {
    ctx = log.WithTags(ctx, "payment", "stripe")

    log.InfoWithFields(ctx, "payment started", map[string]interface{}{
        "payment_id": paymentID,
        "amount": amount,
        "currency": "USD",
    })

    // Process payment...
    err := chargeCard(amount)

    if err != nil {
        log.ErrorWithFields(ctx, "payment failed", map[string]interface{}{
            "error": err,
            "payment_id": paymentID,
            "amount": amount,
            "retry": true,
        })
        return err
    }

    log.InfoWithFields(ctx, "payment succeeded", map[string]interface{}{
        "payment_id": paymentID,
        "amount": amount,
    })

    return nil
}
```

### Example 4: Background Jobs

```go
func RunBackgroundJob(ctx context.Context, jobID string) {
    ctx = log.WithTags(ctx, "background-job", "email-sender")

    log.InfoWithFields(ctx, "job started", map[string]interface{}{
        "job_id": jobID,
        "start_time": time.Now(),
    })

    // Process items
    processed := 0
    failed := 0

    for _, item := range items {
        if err := processItem(item); err != nil {
            failed++
            log.ErrorWithFields(ctx, "item failed", map[string]interface{}{
                "error": err,
                "item_id": item.ID,
            })
        } else {
            processed++
        }
    }

    log.InfoWithFields(ctx, "job completed", map[string]interface{}{
        "job_id": jobID,
        "processed": processed,
        "failed": failed,
        "duration": time.Since(startTime),
    })
}
```

### Example 5: Microservice Communication

```go
func CallExternalService(ctx context.Context, serviceURL string) (*Response, error) {
    ctx = log.WithTags(ctx, "external-call", "payment-service")

    log.InfoWithFields(ctx, "calling external service", map[string]interface{}{
        "service": serviceURL,
        "timeout": 30 * time.Second,
    })

    resp, err := httpClient.Get(serviceURL)

    if err != nil {
        log.ErrorWithFields(ctx, "service call failed", map[string]interface{}{
            "error": err,
            "service": serviceURL,
            "retry_count": retries,
        })
        return nil, err
    }

    log.InfoWithFields(ctx, "service call succeeded", map[string]interface{}{
        "service": serviceURL,
        "status": resp.StatusCode,
        "duration": elapsed,
    })

    return resp, nil
}
```

---

## 🔍 Searching and Filtering Logs

With tags and structured fields, you can easily search and filter logs:

### By Tags

```bash
# Find all database-related logs
jq 'select(.tags | contains(["database"]))' logs.json

# Find logs with specific tag combination
jq 'select(.tags | contains(["payment", "stripe"]))' logs.json
```

### By Fields

```bash
# Find slow queries
jq 'select(.duration > 1000)' logs.json

# Find errors for specific user
jq 'select(.level == "error" and .user_id == "123")' logs.json
```

### Combined

```bash
# Find failed payments with high amounts
jq 'select(.tags | contains(["payment"]) and .level == "error" and .amount > 100)' logs.json
```

---

## 🚀 Summary

- **Tags**: Use for categorization and filtering (`log.WithTags()`)
- **WithFields**: Use for simple structured logging with maps
- **Combine**: Use tags for broad categorization, fields for specific data
- **Choose API**: Simple → WithFields → Event based on complexity

For more examples and advanced features, see:
- `api-log/tags_and_fields_example_test.go`
- `api-log/example_test.go`
- `api-log/MIGRATION_GUIDE.md`
- `api-log/ADVANCED_FEATURES_GUIDE.md` - Middleware configuration, sampling, and performance
