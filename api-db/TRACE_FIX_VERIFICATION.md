# Verification Guide for Trace Context Propagation Fix

## Changes Made

### File: `api-bootstrapper/fxtracer.go`

**Added**: Global trace provider registration

```go
// Set as global trace provider to ensure context propagation works
// This is CRITICAL for distributed tracing:
// - All libraries using otel.GetTracerProvider() will get this provider
// - Ensures parent-child span relationships across boundaries (HTTP -> DB)
// - Enables trace context propagation through pgx database operations
otel.SetTracerProvider(tracerProvider)
```

**Import Added**: `"go.opentelemetry.io/otel"`

## Verification Methods

### Method 1: Code Compilation Test

```bash
cd /home/user/message-gateway
go build ./...
```

**Expected**: All packages compile successfully ✅

**Status**: ✅ **PASSED** - Code compiles without errors

### Method 2: Manual Integration Test

**Prerequisites**:
1. Enable tracing in `configs/config.yaml`:
   ```yaml
   trace:
     enabled: true
     processor:
       type: stdout
       options:
         pretty: true
   ```

2. Start the application:
   ```bash
   go run main.go
   ```

3. Make an HTTP request that triggers a database query:
   ```bash
   curl http://localhost:8080/api/endpoint
   ```

4. Check stdout for trace output:
   ```json
   {
     "Name": "http.request",
     "SpanContext": {
       "TraceID": "abc123...",
       "SpanID": "def456..."
     },
     "Parent": {},
     "SpanKind": 2,
     "StartTime": "...",
     "EndTime": "...",
     "ChildSpanCount": 1
   }
   {
     "Name": "postgresql.query",
     "SpanContext": {
       "TraceID": "abc123...",  // ✅ Same TraceID as parent!
       "SpanID": "ghi789..."
     },
     "Parent": {
       "TraceID": "abc123...",
       "SpanID": "def456..."    // ✅ References parent span!
     },
     "SpanKind": 3,
     "StartTime": "...",
     "EndTime": "...",
     "Attributes": {
       "db.name": "your_database",
       "db.query_name": "your_query",
       "db.query_type": "SELECT"
     }
   }
   ```

**Verification Points**:
- ✅ `postgresql.query` span should have **same TraceID** as parent `http.request`
- ✅ `postgresql.query` span should have **Parent.SpanID** matching parent's SpanID
- ✅ Both spans should appear in the output (not just parent)

**Before Fix**: Database spans would have different TraceID or no parent reference
**After Fix**: Database spans properly nested under parent HTTP span

### Method 3: Programmatic Verification

Add this code to your application startup to verify:

```go
package main

import (
    "log"
    "go.opentelemetry.io/otel"
    otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func verifyGlobalTraceProvider() {
    provider := otel.GetTracerProvider()

    // Check if it's the SDK provider (not no-op)
    if _, ok := provider.(*otelsdktrace.TracerProvider); ok {
        log.Println("✅ Global trace provider is SDK TracerProvider (correct)")
    } else {
        log.Println("❌ Global trace provider is NOT set or is no-op (broken)")
    }

    // Verify we can create spans
    tracer := otel.Tracer("verification")
    ctx, span := tracer.Start(context.Background(), "test")
    defer span.End()

    if span.SpanContext().IsValid() {
        log.Println("✅ Can create valid spans (correct)")
    } else {
        log.Println("❌ Spans are invalid/no-op (broken)")
    }
}
```

Call this function after bootstrapper starts and before serving requests.

**Expected Output**:
```
✅ Global trace provider is SDK TracerProvider (correct)
✅ Can create valid spans (correct)
```

### Method 4: Context Propagation Test

Create a test endpoint that verifies context propagation:

```go
func TestContextPropagationHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Create a parent span
    tracer := otel.Tracer("test-handler")
    ctx, parentSpan := tracer.Start(ctx, "test-request")
    defer parentSpan.End()

    parentSpanContext := parentSpan.SpanContext()
    log.Printf("Parent span - TraceID: %s, SpanID: %s",
        parentSpanContext.TraceID(),
        parentSpanContext.SpanID())

    // Execute a database query
    err := db.WithTx(ctx, func(tx pgx.Tx) error {
        _, err := tx.Query(ctx, "SELECT 1")
        return err
    })

    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.Write([]byte("Check logs for trace IDs - they should match!"))
}
```

**Verification**:
- Parent span TraceID and DB span TraceID should be **identical**
- DB span should be child of parent span
- Check trace exporter output to confirm parent-child relationship

## Expected Behavior After Fix

### ✅ Correct Behavior (With Fix)

```
TraceID: abc123def456...
├─ Span: http.request (parent)
│  ├─ SpanID: span001
│  ├─ Duration: 150ms
│  └─ Children: 1
│
└─ Span: postgresql.query (child)
   ├─ SpanID: span002
   ├─ Parent: span001
   ├─ Duration: 45ms
   └─ Attributes:
      ├─ db.name: "users_db"
      ├─ db.query_name: "GetUserByID"
      └─ db.query_type: "SELECT"
```

**Characteristics**:
- ✅ Single TraceID for entire request
- ✅ Parent-child relationship established
- ✅ Complete trace tree visible
- ✅ Can track request from HTTP to database

### ❌ Broken Behavior (Without Fix)

```
TraceID: abc123... (HTTP span - no-op provider)
└─ Span: http.request
   ├─ Duration: 150ms
   └─ Children: 0  ❌ No children!

TraceID: xyz789... (DB span - different provider)
└─ Span: postgresql.query  ❌ Disconnected!
   ├─ Duration: 45ms
   └─ Parent: NONE  ❌ Orphaned!
```

**Characteristics**:
- ❌ Multiple TraceIDs (trace fragmentation)
- ❌ No parent-child relationship
- ❌ Orphaned database spans
- ❌ Cannot correlate HTTP requests with DB queries

## Troubleshooting

### Issue: Still seeing orphaned DB spans

**Possible Causes**:
1. Application not restarted after fix
2. Config has `trace.enabled: false`
3. Multiple trace provider instances being created
4. Global provider being overwritten somewhere

**Solution**:
```bash
# Rebuild completely
go clean -cache
go build ./...

# Verify fix is applied
grep -A 5 "otel.SetTracerProvider" api-bootstrapper/fxtracer.go
```

### Issue: No spans appearing at all

**Possible Causes**:
1. Trace processor not configured
2. Exporter not set up
3. Sampler set to "always-off"

**Solution**:
Check config:
```yaml
trace:
  enabled: true  # ✅ Must be true
  processor:
    type: stdout  # or otlpgrpc
  sampler:
    type: always-on  # For testing
    options:
      ratio: 1.0
```

### Issue: Spans appear but with different TraceIDs

**This means the fix didn't work!**

**Debug Steps**:
1. Verify `otel.SetTracerProvider(tracerProvider)` is called
2. Add logging after the call:
   ```go
   otel.SetTracerProvider(tracerProvider)
   log.Println("Global trace provider set:", otel.GetTracerProvider())
   ```
3. Check that no other code is calling `otel.SetTracerProvider` later
4. Verify FX module load order (fxTrace should load before fxDB)

## Success Criteria

- ✅ Code compiles without errors
- ✅ Application starts successfully
- ✅ Database queries create spans
- ✅ DB spans have same TraceID as parent HTTP spans
- ✅ DB spans reference parent span ID
- ✅ Complete trace tree visible in exporter output
- ✅ Can trace requests end-to-end through database

## Regression Testing

To ensure the fix doesn't break anything:

1. **Run all existing tests**:
   ```bash
   go test ./... -v
   ```

2. **Test with tracing disabled**:
   ```yaml
   trace:
     enabled: false
   ```
   Application should still work normally (no-op tracer)

3. **Test with different exporters**:
   - stdout
   - otlpgrpc
   - jaeger (if available)

4. **Test under load**:
   ```bash
   # Ensure tracing doesn't cause performance issues
   ab -n 1000 -c 10 http://localhost:8080/api/endpoint
   ```

## Documentation Updates Needed

After verification, update:

1. **README.md** - Add tracing setup instructions
2. **config.yaml.example** - Add trace configuration example
3. **ARCHITECTURE.md** - Document trace provider setup
4. **OBSERVABILITY.md** - Explain distributed tracing setup

## Related Files Modified

- `api-bootstrapper/fxtracer.go` - Added `otel.SetTracerProvider()` call
- `api-db/TRACE_CONTEXT_ISSUE.md` - Problem analysis (new)
- `api-db/TRACE_FIX_VERIFICATION.md` - This verification guide (new)
