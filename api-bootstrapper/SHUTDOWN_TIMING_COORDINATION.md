# Graceful Shutdown Timing and Coordination

## Timeout Configuration

### Router Adapter Shutdown Timeout

**Location**: `api-bootstrapper/bootstrapper.go:623`

```go
OnStop: func(ctx context.Context) error {
    shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)  // ← 10 seconds
    defer cancel()

    if err := p.Adapter.Shutdown(shutdownCtx); err != nil {
        return err
    }

    return nil
}
```

**Timeout**: **10 seconds**

**Purpose**:
- Stop accepting new HTTP connections
- Drain in-flight HTTP requests
- Allow handlers to complete their work
- Close HTTP server gracefully

---

### Database Shutdown Timeout

**Location**: `api-bootstrapper/bootstrapper.go:424`

```go
OnStop: func(ctx context.Context) error {
    drainTimeout := 5 * time.Second  // ← 5 seconds
    drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
    defer cancel()

    // Wait for active connections to drain
    for {
        select {
        case <-drainCtx.Done():
            // Timeout - force close
            goto closeDB
        case <-ticker.C:
            if count.AcquiredConns() == 0 {
                // All connections idle - close immediately
                goto closeDB
            }
        }
    }

closeDB:
    p.DB.Close()
    return nil
}
```

**Timeout**: **5 seconds**

**Purpose**:
- Wait for active database connections to complete
- Allow in-flight queries/transactions to finish
- Monitor `AcquiredConns()` metric
- Close pool when all connections idle

---

## Module Order in Bootstrapper

**Location**: `api-bootstrapper/bootstrapper.go:52-61`

```go
options: []fx.Option{
    fxconfig,           // 1. Configuration
    fxlog,              // 2. Logging
    fxDB,               // 3. Database ← Position 3
    fxRouterAdapter,    // 4. Router  ← Position 4
    fxTrace,            // 5. Tracing
    fxMetrics,          // 6. Metrics
}
```

---

## FX Shutdown Order (CRITICAL!)

### How FX Works

**FX shuts down modules in REVERSE order of startup**

### Startup Sequence

```
1. fxconfig      → Initialize configuration
2. fxlog         → Initialize logging
3. fxDB          → Connect to database
4. fxRouterAdapter → Start HTTP server
5. fxTrace       → Initialize tracing
6. fxMetrics     → Initialize metrics
```

### Shutdown Sequence (REVERSE!)

```
6. fxMetrics     → Stop metrics collection
5. fxTrace       → Stop tracing
4. fxRouterAdapter → Shutdown HTTP server (10s timeout) ← STOPS FIRST
3. fxDB          → Drain and close database (5s timeout) ← STOPS SECOND
2. fxlog         → Close logging
1. fxconfig      → Cleanup configuration
```

---

## Coordination Mechanism

### Why This Order Matters

```
Position in options array:
  fxDB:            Position 3 (earlier)
  fxRouterAdapter: Position 4 (later)

Shutdown order:
  fxRouterAdapter: Shuts down FIRST  ← Router stops accepting requests
  fxDB:            Shuts down SECOND ← Database waits for connections
```

**Key Principle**: The router **MUST** stop before the database to allow in-flight requests to complete.

---

## Complete Shutdown Timeline

### Visual Timeline

```
Time    Event                           Module          Timeout
────────────────────────────────────────────────────────────────────
T+0s    SIGTERM/SIGINT received         Signal Handler  -
        Context cancelled               Bootstrapper    -
        FX begins shutdown              FX Runtime      -

T+0s    Router shutdown starts          fxRouterAdapter 10s
        ├─ Stop accepting new requests
        ├─ Drain in-flight requests
        └─ HTTP handlers processing...

        [HTTP Request 1 running]
        [HTTP Request 2 running]
        [HTTP Request 3 running]

T+2s    [HTTP Request 1 completes]
        [HTTP Request 2 completes]

T+4s    [HTTP Request 3 completes]

T+4s    All HTTP requests complete      fxRouterAdapter -
        Router shutdown complete        fxRouterAdapter ✓

T+4s    Database shutdown starts        fxDB            5s
        ├─ Log connection stats
        │  total=10, idle=8, acquired=2
        └─ Wait for AcquiredConns() == 0

        [DB Connection 1 in use]
        [DB Connection 2 in use]

T+5s    [DB Connection 1 released]

T+6s    [DB Connection 2 released]

T+6s    All connections idle            fxDB            -
        Database closes immediately     fxDB            ✓

T+6s    Shutdown complete               Application     -
────────────────────────────────────────────────────────────────────
Total time: 6 seconds (in this example)
```

---

## Detailed Coordination Flow

### Phase 1: Router Shutdown (T+0 to T+4s in example)

```go
// Router OnStop hook executes
OnStop: func(ctx context.Context) error {
    // Create 10-second timeout
    shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // Call adapter.Shutdown() with timeout
    if err := p.Adapter.Shutdown(shutdownCtx); err != nil {
        return err
    }

    return nil // Returns after HTTP server stops
}
```

**What happens during these 10 seconds:**

1. **T+0s**: HTTP server stops accepting new connections
2. **T+0s - T+10s**: Existing HTTP requests continue processing
   - Handlers can still access database
   - Context is signal-aware (can detect shutdown)
   - Handlers can finish gracefully
3. **T+0s - T+10s**: HTTP server waits for:
   - All active requests to complete
   - All response bodies to be sent
   - All connections to close cleanly
4. **T+10s or earlier**: Router shutdown completes
   - Returns from OnStop hook
   - FX proceeds to next module (fxDB)

**If timeout reached:**
- HTTP server force-closes remaining connections
- Returns from OnStop (may return error)
- FX proceeds to fxDB shutdown anyway

---

### Phase 2: Database Shutdown (T+4s to T+6s in example)

```go
// Database OnStop hook executes (only after router stops)
OnStop: func(ctx context.Context) error {
    // Log initial connection stats
    // total_conns=10, idle_conns=8, acquired_conns=2

    // Create 5-second drain timeout
    drainCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Poll every 100ms for idle connections
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-drainCtx.Done():
            // 5 seconds elapsed - force close
            logger.Warn().Msg("Drain timeout, forcing close")
            goto closeDB

        case <-ticker.C:
            // Check connection status
            if count.AcquiredConns() == 0 {
                // All connections idle - safe to close!
                logger.Info().Msg("All connections drained")
                goto closeDB
            }
            // Otherwise, continue waiting...
        }
    }

closeDB:
    p.DB.Close()
    return nil
}
```

**What happens during these 5 seconds:**

1. **T+0s (DB shutdown start)**: Log connection stats
   - Example: 10 total, 8 idle, 2 acquired (in use)

2. **T+0s - T+5s**: Poll every 100ms
   - Check `AcquiredConns()` count
   - If == 0, close immediately (don't wait full 5s!)
   - If > 0, continue waiting

3. **T+0s - T+5s**: Active connections finishing
   - Connection 1 releases (T+1s): acquired=1
   - Connection 2 releases (T+2s): acquired=0 ✓
   - Close immediately at T+2s

4. **T+5s or earlier**: Database closes
   - Returns from OnStop hook
   - FX proceeds to next module

**If timeout reached:**
- Log warning with remaining acquired count
- Force close database pool
- Active transactions may fail
- Returns from OnStop

---

## Maximum vs Typical Shutdown Times

### Maximum (Worst Case)

```
Router timeout:     10 seconds (all in-flight requests take full time)
Database timeout:   +5 seconds (connections don't drain in time)
──────────────────────────────────
Total maximum:      15 seconds
```

### Typical (Best Case)

```
Router shutdown:    2 seconds (requests complete quickly)
Database drain:     +0.5 seconds (connections release immediately)
──────────────────────────────────
Total typical:      2.5 seconds
```

### Observed Example

```
Router shutdown:    4 seconds (3 requests complete)
Database drain:     +2 seconds (2 connections drain)
──────────────────────────────────
Total observed:     6 seconds
```

---

## Why Database Gets 5 Seconds AFTER Router

### The Problem Without Coordination

```
❌ BAD: If they ran in parallel or database first

T+0s    Router shutdown starts (10s)
T+0s    Database shutdown starts (5s) ← Problem!

T+5s    Database timeout - closes! ← Connections still in use!

T+6s    HTTP Request tries to query
        ERROR: "connection closed" ← Failed request!

T+10s   Router shutdown completes
```

### The Solution: Sequential Shutdown

```
✓ GOOD: Router first, then database

T+0s    Router shutdown starts (10s)
        HTTP handlers processing...
        Database still available ← Handlers can use DB!

T+4s    All HTTP requests complete
        Router shutdown completes ✓

T+4s    NOW database shutdown starts (5s)
        Only residual connections remain

T+6s    Database closes cleanly ✓
```

---

## Coordination Through FX

### FX's Guarantee

**From Uber FX documentation:**

> "OnStop hooks are called in reverse order of OnStart hooks"

This means:
```
If startup order is:  [A, B, C, D]
Then shutdown order is: [D, C, B, A]
```

### Our Configuration

```go
options: []fx.Option{
    fxconfig,        // A
    fxlog,           // B
    fxDB,            // C ← Starts 3rd, stops 3rd from end
    fxRouterAdapter, // D ← Starts 4th, stops FIRST
    fxTrace,         // E
    fxMetrics,       // F
}

Startup:  [config, log, DB, Router, Trace, Metrics]
Shutdown: [Metrics, Trace, Router, DB, log, config]
                           ↑       ↑
                         FIRST   SECOND
```

### Why This Works

1. **FX calls Router's OnStop first**
   - Router has 10 seconds to drain HTTP requests
   - Database is still available during this time
   - HTTP handlers can complete their database operations

2. **FX waits for Router's OnStop to return**
   - Blocks until router finishes or timeout
   - Ensures HTTP server is fully stopped

3. **Only then FX calls Database's OnStop**
   - All HTTP traffic has stopped
   - Only residual database connections remain
   - Safe to drain and close

---

## Real-World Scenario

### Example: Slow API Endpoint

```go
// HTTP Handler that takes 8 seconds
func SlowHandler(c *fiber.Ctx) error {
    ctx := c.UserContext()

    // Long database query (8 seconds)
    var results []Data
    err := db.WithTx(ctx, func(tx pgx.Tx) error {
        return tx.Select(ctx, &results,
            "SELECT * FROM large_table WHERE complex_condition")
    })

    return c.JSON(results)
}
```

**Shutdown while request is running:**

```
T+0s    User sends request to /slow-endpoint
        Handler starts, begins 8-second query

T+2s    Admin presses Ctrl+C (SIGTERM)
        Context cancelled
        FX begins shutdown

T+2s    Router OnStop starts (10s timeout)
        - Stops accepting NEW requests ✓
        - Allows /slow-endpoint to continue ✓
        - Database still available ✓

T+6s    Query running...
T+8s    Query running...
T+10s   Query completes! ✓
        Handler returns response to user ✓

T+10s   Router shutdown completes

T+10s   Database OnStop starts (5s timeout)
        - Checks AcquiredConns() = 0 ✓
        - All connections idle ✓
        - Closes immediately (no 5s wait needed)

T+10s   Database shutdown completes

T+10s   Application exits gracefully ✓
        User received complete response ✓
        No errors logged ✓
```

---

## Timeout Adjustment Guidelines

### When to Increase Router Timeout

**Current**: 10 seconds

**Increase to 15-30 seconds if:**
- Long-running API endpoints (reports, exports)
- Large file uploads/downloads
- Complex data processing
- Third-party API calls with retries

**Example:**
```go
shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
```

### When to Increase Database Timeout

**Current**: 5 seconds

**Increase to 10-15 seconds if:**
- Complex long-running queries
- Large transaction processing
- Many concurrent connections
- Slow network to database

**Example:**
```go
drainTimeout := 10 * time.Second
```

### When to Decrease (Not Recommended)

**Risks:**
- Failed requests during shutdown
- Incomplete transactions
- Data inconsistency
- Poor user experience

**Only decrease if:**
- Fast shutdown is critical (dev environment)
- All requests are <1 second
- Acceptable to drop in-flight requests

---

## Monitoring Coordination

### Log Markers to Watch

**Successful coordination:**
```
[T+0s]  INFO  Shutdown signal received, initiating graceful shutdown...
[T+0s]  INFO  Router adapter shutdown initiated
[T+4s]  INFO  Router adapter shutdown complete              ← Router done
[T+4s]  INFO  Database connection stats at shutdown start   ← DB starts
[T+4s]  INFO  Waiting for active database connections to drain...
[T+6s]  INFO  All database connections drained successfully ← DB done
[T+6s]  INFO  Database shutdown complete
[T+6s]  INFO  Application shutdown complete
```

**Timing issue (too fast):**
```
[T+0s]  INFO  Shutdown signal received...
[T+0s]  INFO  Router adapter shutdown initiated
[T+0s]  WARN  Router shutdown timeout reached!              ← Problem!
[T+10s] INFO  Router adapter shutdown complete (forced)
[T+10s] INFO  Database connection stats... acquired_conns=5 ← Still active!
[T+15s] WARN  Drain timeout reached, forcing close          ← Had to force
```

---

## Summary

### Timeouts

| Component | Timeout | Purpose |
|-----------|---------|---------|
| Router Adapter | **10 seconds** | Drain HTTP requests |
| Database | **5 seconds** | Drain DB connections |
| **Total Maximum** | **15 seconds** | Complete graceful shutdown |

### Coordination Method

**Mechanism**: FX's reverse-order shutdown

**Order**:
1. Router stops first (fxRouterAdapter OnStop)
2. Database stops second (fxDB OnStop)

**Guarantee**: Database has full 5 seconds AFTER router completes

**Result**:
- Zero failed HTTP requests
- Complete database operations
- Clean graceful shutdown
- Full observability

### Key Insight

The coordination happens **automatically through module ordering**:

```go
options: []fx.Option{
    ...
    fxDB,            // Earlier position → Stops later
    fxRouterAdapter, // Later position → Stops earlier
    ...
}
```

This simple ordering ensures proper shutdown sequence without manual coordination code!
