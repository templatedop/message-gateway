# Trace Context Propagation Issue in api-db

## Problem Statement

The OpenTelemetry trace provider is not set as the global provider, causing **context propagation to break** when using pgx database library with tracing enabled.

## Root Cause Analysis

### Issue #1: Global Trace Provider Not Set

**Location**: `api-bootstrapper/fxtracer.go:51-101`

```go
func NewFxTracerProvider(param FxTraceParam) (*otelsdktrace.TracerProvider, error) {
    // ... creates tracerProvider ...

    return tracerProvider, nil  // ❌ Returns but doesn't set global provider
}
```

**Problem**:
- TracerProvider is created and returned via FX dependency injection
- **NOT** registered as global OTel provider using `otel.SetTracerProvider()`
- This causes two separate trace provider instances:
  1. **SDK TracerProvider** - injected into database module
  2. **Global TracerProvider** - default no-op provider

### Issue #2: Mixed Provider Usage

**Database tracer initialization** (`api-db/factory.go:106-114`):
```go
if cfg.Trace {
    var tracer dbtracer.Tracer
    tracer, err = dbtracer.NewDBTracer(
        cfg.DBDatabase,
        dbtracer.WithTraceProvider(osdktrace),  // Uses injected SDK provider
    )
    // ...
}
```

**Database tracer default** (`api-db/tracer/dbtracer.go:38-52`):
```go
optCtx := optionCtx{
    // ...
    traceProvider: otel.GetTracerProvider(),  // Falls back to GLOBAL provider
}
for _, opt := range opts {
    opt(&optCtx)  // Overrides with SDK provider if provided
}
```

**Problem**:
- When `WithTraceProvider(osdktrace)` is passed → uses SDK provider
- When NOT passed → uses global provider (no-op by default)
- **ANY code using `otel.GetTracerProvider()`** gets the global (no-op) provider

### Issue #3: Context Propagation Breaks

**Span creation in tracer** (`api-db/tracer/tracequery.go:25-46`):
```go
func (dt *dbTracer) TraceQueryStart(
    ctx context.Context,
    _ *pgx.Conn,
    data pgx.TraceQueryStartData,
) context.Context {
    queryName, queryType := queryNameFromSQL(data.SQL)
    ctx, span := dt.getTracer().Start(ctx, "postgresql.query")  // ⚠️ Creates span
    // ...
    return context.WithValue(ctx, dbTracerQueryCtxKey, &traceQueryData{
        span: span,  // ⚠️ Stores span in context
    })
}
```

**Where it breaks**:
```
HTTP Handler (uses otel.GetTracerProvider())
    └─> Creates parent span with GLOBAL provider (no-op)
        └─> Calls DB query with context
            └─> pgx tracer creates child span with SDK provider
                └─> ❌ Parent-child relationship BROKEN!
```

**Impact**:
- ❌ Database query spans are **NOT** connected to parent HTTP request spans
- ❌ Distributed tracing chain breaks at database boundary
- ❌ Cannot trace requests end-to-end through database
- ❌ Trace context (trace ID, span ID) **NOT** propagated to DB operations
- ❌ If using libraries that call `otel.GetTracerProvider()`, they get no-op provider

## Example Scenario

### Current Broken Behavior:

```go
// 1. HTTP handler creates span (global provider - no-op)
ctx, span := otel.GetTracerProvider().Tracer("http").Start(ctx, "HandleRequest")

// 2. Call database (pgx uses SDK provider)
err := db.WithTx(ctx, func(tx pgx.Tx) error {
    // 3. pgx tracer creates span with SDK provider
    // ❌ This span has NO parent! Context propagation broken!
    _, err := tx.Query(ctx, "SELECT * FROM users")
    return err
})
```

**Result**: Two disconnected trace trees instead of one unified trace

### Expected Correct Behavior:

```
TraceID: abc123
  └─> Span: HandleRequest (parent)
      └─> Span: postgresql.query (child)  ✅ Connected!
```

## How pgx Tracer Works

1. **Query Execution**:
   ```go
   tx.Query(ctx, "SELECT * FROM users")
   ```

2. **pgx calls TraceQueryStart**:
   ```go
   newCtx := tracer.TraceQueryStart(ctx, conn, data)
   ```

3. **Tracer creates span from incoming context**:
   ```go
   ctx, span := dt.getTracer().Start(ctx, "postgresql.query")
   ```
   - ✅ If `ctx` has parent span → child span created (context propagation works)
   - ❌ If parent span from different provider → **NOT CONNECTED**

4. **Query executes with new context**:
   - Contains span information

5. **pgx calls TraceQueryEnd**:
   ```go
   tracer.TraceQueryEnd(newCtx, conn, data)
   ```

6. **Tracer ends span**:
   ```go
   span.End()
   ```

## Why Global Provider Matters

### Libraries Using Global Provider:
- HTTP middleware (like `otelhttp`)
- gRPC interceptors (like `otelgrpc`)
- Custom instrumentation calling `otel.Tracer("name")`
- Any code using `otel.GetTracerProvider().Tracer("name")`

### If global provider is no-op:
- ❌ All spans from these libraries are discarded
- ❌ Context propagation chain breaks
- ❌ Parent-child span relationships lost

## Solution

### Fix #1: Set Global Trace Provider (Required)

**File**: `api-bootstrapper/fxtracer.go`

Add after creating the tracer provider:
```go
func NewFxTracerProvider(param FxTraceParam) (*otelsdktrace.TracerProvider, error) {
    // ... existing code to create tracerProvider ...

    // ✅ Set as global provider for context propagation
    otel.SetTracerProvider(tracerProvider)

    param.LifeCycle.Append(fx.Hook{
        OnStop: func(ctx context.Context) error {
            // ... existing shutdown code ...
        },
    })

    return tracerProvider, nil
}
```

**Why this works**:
- All code using `otel.GetTracerProvider()` gets the configured provider
- Spans from HTTP handlers, gRPC, and database all use same provider
- Parent-child relationships maintained across boundaries
- Trace context propagates correctly through entire request chain

### Fix #2: Consider Using Global Provider in DB Factory (Optional)

**Current** (`api-db/factory.go:106-114`):
```go
tracer, err = dbtracer.NewDBTracer(
    cfg.DBDatabase,
    dbtracer.WithTraceProvider(osdktrace),  // Explicit provider injection
)
```

**Alternative** (after Fix #1 is applied):
```go
tracer, err = dbtracer.NewDBTracer(
    cfg.DBDatabase,
    // ✅ No explicit provider - uses otel.GetTracerProvider() (global)
)
```

**Benefits**:
- Simpler code (no need to inject provider)
- Guaranteed to use same provider as rest of application
- Already defaults to global provider if no override

**Note**: This is optional because after Fix #1, both approaches work the same way since global provider is set.

## Testing the Fix

### Test 1: Verify Global Provider Set
```go
// After application starts
provider := otel.GetTracerProvider()
_, ok := provider.(*otelsdktrace.TracerProvider)
assert.True(t, ok, "Global provider should be SDK provider, not no-op")
```

### Test 2: Verify Context Propagation
```go
// Create parent span
ctx, parentSpan := otel.Tracer("test").Start(context.Background(), "parent")
defer parentSpan.End()

// Execute DB query (should create child span)
err := db.WithTx(ctx, func(tx pgx.Tx) error {
    _, err := tx.Query(ctx, "SELECT 1")
    return err
})

// Verify: DB span should be child of parent span
// (Check via trace exporter output)
```

### Test 3: End-to-End Trace
```bash
# Enable stdout trace exporter in config
trace:
  enabled: true
  processor:
    type: stdout
    options:
      pretty: true

# Make HTTP request
curl http://localhost:8080/api/users

# Check output: Should see complete trace tree:
# TraceID: xxx
#   Span: http.request (parent)
#     Span: postgresql.query (child)  ✅ Connected!
```

## Impact of NOT Fixing

### Without Fix:
- ❌ Database operations invisible in distributed traces
- ❌ Cannot debug slow queries via tracing
- ❌ Cannot correlate DB performance with HTTP requests
- ❌ Trace context (baggage, trace IDs) lost at DB boundary
- ❌ Observability severely limited

### With Fix:
- ✅ Complete end-to-end tracing through database
- ✅ See exact queries executed for each HTTP request
- ✅ Measure database latency in context of full request
- ✅ Trace context flows through entire application
- ✅ Full observability and debugging capability

## Related Files

1. **api-bootstrapper/fxtracer.go** - Trace provider initialization (NEEDS FIX)
2. **api-bootstrapper/bootstrapper.go** - DB module setup (passes SDK provider)
3. **api-db/factory.go** - DB tracer initialization (uses injected provider)
4. **api-db/tracer/dbtracer.go** - Tracer creation (defaults to global)
5. **api-db/tracer/tracequery.go** - Query span creation (where propagation happens)
6. **api-db/tracer/options.go** - Provider configuration options

## Recommendation

**Priority**: 🔴 **CRITICAL**

**Action**: Apply Fix #1 (set global trace provider) immediately

**Reason**: Without this fix, distributed tracing is fundamentally broken and database operations are not properly traced.
