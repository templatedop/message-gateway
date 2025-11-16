# Router Shutdown Behavior - Critical Understanding

## Question 1: Database Execution During Router Shutdown

### Short Answer
**YES** - Database operations continue executing during router shutdown. Router shutdown is a **WAITING** phase, not a **STOPPING** phase.

---

## What "Router Shutdown" Actually Does

### Code: Router OnStop Hook

**Location**: `api-bootstrapper/bootstrapper.go:622-634`

```go
OnStop: func(ctx context.Context) error {
    shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // This call BLOCKS and WAITS!
    if err := p.Adapter.Shutdown(shutdownCtx); err != nil {
        return err
    }

    log.GetBaseLoggerInstance().ToZerolog().Info().
        Msg("Router adapter shutdown complete")

    return nil  // Only returns AFTER handlers finish
}
```

### What `Adapter.Shutdown()` Does Internally

**For Gin/Echo/net/http adapters:**

```go
func (a *GinAdapter) Shutdown(ctx context.Context) error {
    if a.server == nil {
        return nil
    }

    // This is http.Server.Shutdown()
    // It does NOT kill handlers immediately!
    return a.server.Shutdown(ctx)
}
```

### http.Server.Shutdown() Behavior

**From Go documentation:**

```
Shutdown gracefully shuts down the server without interrupting any
active connections. Shutdown works by:

1. First closing all open listeners (no NEW connections)
2. Then closing all idle connections
3. Then WAITING for active connections to become idle
4. Only returns when all connections idle (or context timeout)
```

**Key Points:**

- ✅ Does NOT kill active handlers
- ✅ Does NOT interrupt database operations
- ✅ WAITS for handlers to complete (up to timeout)
- ✅ Database remains available during wait

---

## Visual Flow: What Happens During Router Shutdown

### Timeline with Active Database Operation

```
T+0s    HTTP Request arrives: POST /create-user
        ├─ Handler starts executing
        └─ Begin database transaction

T+1s    Handler processing...
        ├─ INSERT INTO users (...)
        └─ INSERT INTO audit_log (...)

T+2s    Admin presses Ctrl+C (SIGTERM)
        ├─ Signal received
        └─ Router OnStop hook triggered

T+2s    router.Shutdown(10s timeout) called
        ├─ Server stops listening for NEW requests ✓
        ├─ Existing handler STILL RUNNING ✓
        └─ Database STILL AVAILABLE ✓

        ┌─────────────────────────────────────────┐
        │  SHUTDOWN WAITING PHASE                 │
        │                                         │
        │  - Handler continues executing          │
        │  - Database operations proceed normally │
        │  - Transaction continues                │
        │  - Router.Shutdown() is BLOCKED here    │
        └─────────────────────────────────────────┘

T+3s    Handler still executing...
        ├─ Database query continues
        └─ tx.Commit() executing

T+4s    Handler completes!
        ├─ Transaction committed ✓
        ├─ Response generated ✓
        └─ Response written to connection ✓

T+4s    Connection becomes idle
        └─ Router.Shutdown() RETURNS NOW ✓

T+4s    Router OnStop returns to FX

T+4s    FX proceeds to Database OnStop
        └─ Safe to close DB (no handlers running)
```

---

## Code Example: Handler During Shutdown

### Handler Code

```go
func CreateUser(c *fiber.Ctx) error {
    var user User
    if err := c.BodyParser(&user); err != nil {
        return err
    }

    // This database operation continues even during router shutdown!
    err := db.WithTx(c.UserContext(), func(tx pgx.Tx) error {
        // T+1s: Insert user
        _, err := tx.Exec(ctx,
            "INSERT INTO users (name, email) VALUES ($1, $2)",
            user.Name, user.Email)
        if err != nil {
            return err
        }

        // T+2s: SIGTERM received - Router.Shutdown() called
        //       But this handler CONTINUES EXECUTING!

        // T+3s: Insert audit log
        _, err = tx.Exec(ctx,
            "INSERT INTO audit_log (action, user_id) VALUES ($1, $2)",
            "create_user", user.ID)
        if err != nil {
            return err
        }

        // T+4s: Commit transaction - SUCCEEDS!
        return nil
    })

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // T+4s: Return response - User receives it!
    return c.Status(201).JSON(user)
}
```

### Execution Flow

```
Request arrives at T+0s
├─ T+0s: Handler starts
├─ T+1s: First INSERT executes
├─ T+2s: SIGTERM received → Router.Shutdown() called
│         ├─ Stops NEW requests
│         ├─ THIS handler continues ✓
│         └─ Database available ✓
├─ T+3s: Second INSERT executes (during shutdown!)
├─ T+4s: COMMIT succeeds (during shutdown!)
├─ T+4s: Response sent to user ✓
└─ T+4s: Handler completes → Router.Shutdown() returns
```

**Result**:
- ✅ All database operations complete
- ✅ Transaction committed successfully
- ✅ User receives 201 Created response
- ✅ No errors!

---

## Question 2: If Router Closes First, Does Response Reach User?

### Short Answer
**YES** - The response reaches the user because router shutdown waits for BOTH handler completion AND response delivery.

---

## How HTTP Response Delivery Works

### Complete Request/Response Cycle

```
1. User sends request
   ↓
2. TCP connection established
   ↓
3. Handler processes request
   ↓
4. Handler generates response
   ↓
5. Response written to TCP buffer
   ↓
6. TCP sends response packets to user
   ↓
7. User receives response
   ↓
8. Connection closes
```

### Router Shutdown Waits for Step 8!

**Router.Shutdown() waits for:**
- ✅ Handler execution complete
- ✅ Response generated
- ✅ Response written to connection
- ✅ Response fully transmitted
- ✅ Connection closed gracefully

---

## Proof: Go's http.Server.Shutdown() Implementation

### From Go Source Code (net/http/server.go)

```go
// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
func (srv *Server) Shutdown(ctx context.Context) error {
    srv.mu.Lock()
    lnerr := srv.closeListenersLocked()  // Stop accepting NEW connections
    srv.closeDoneChanLocked()

    // Wait for all connections to become idle
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        if srv.closeIdleConns() {  // Try to close idle connections
            return lnerr
        }
        select {
        case <-ctx.Done():
            return ctx.Err()  // Timeout
        case <-ticker.C:
            // Continue waiting for active connections
        }
    }
}
```

**Key insight**: It polls until connections are idle, meaning response has been sent!

---

## Visual Flow: Response Delivery During Shutdown

### Detailed Timeline

```
T+0s    Request: GET /users/123
        ├─ Connection established
        └─ Handler starts

T+1s    Handler queries database
        └─ SELECT * FROM users WHERE id = 123

T+2s    SIGTERM received
        └─ Router.Shutdown() called
            ├─ Stop listening for NEW requests
            └─ Start waiting for active connections

T+3s    Database query completes
        ├─ Handler receives user data
        └─ Handler generates JSON response

T+3.1s  Response written to TCP buffer
        └─ HTTP/1.1 200 OK
            Content-Type: application/json
            Content-Length: 156

            {"id":123,"name":"John","email":"john@example.com"}

T+3.2s  TCP transmission
        ├─ Response packets sent over network
        └─ User's browser receives packets

T+3.3s  User receives complete response ✓
        └─ Browser displays: {"id":123,...}

T+3.4s  Connection marked as idle
        └─ Router.Shutdown() detects idle connection

T+3.4s  Router.Shutdown() RETURNS ✓
        └─ OnStop hook completes

T+3.4s  FX proceeds to Database OnStop
```

**Result**: User received complete response before shutdown completed!

---

## Example: Slow Response During Shutdown

### Code

```go
func GenerateReport(c *fiber.Ctx) error {
    ctx := c.UserContext()

    // Long database query - 8 seconds
    var report ReportData
    err := db.WithTx(ctx, func(tx pgx.Tx) error {
        return tx.Get(ctx, &report, `
            SELECT ...
            FROM large_table
            WHERE complex_conditions
        `)
    })

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // Generate large JSON response
    return c.JSON(report)  // 5MB response
}
```

### Timeline

```
T+0s    Request: GET /generate-report
T+0s    Handler starts 8-second query
T+1s    SIGTERM received → Router.Shutdown(10s) called
T+1s    Router waits... (Handler still running)
T+2s    Router waits... (Query executing)
T+3s    Router waits... (Query executing)
...
T+8s    Query completes! ✓
T+8s    Response generated (5MB JSON)
T+8.0s  Response writing starts
T+8.5s  Response transmission ongoing...
T+9.0s  Response transmission ongoing...
T+9.5s  Response fully transmitted ✓
T+9.5s  User receives complete 5MB response ✓
T+9.5s  Connection idle
T+9.5s  Router.Shutdown() RETURNS ✓

Total: 9.5 seconds (within 10s timeout)
```

**Result**:
- ✅ User receives complete 5MB report
- ✅ No timeout (9.5s < 10s)
- ✅ Clean shutdown

---

## What If Timeout Is Reached?

### Scenario: Handler Takes Too Long

```go
func VerySlowOperation(c *fiber.Ctx) error {
    // This takes 15 seconds
    time.Sleep(15 * time.Second)
    return c.JSON(fiber.Map{"status": "done"})
}
```

### Timeline

```
T+0s    Request arrives
T+0s    Handler starts (15s operation)
T+1s    SIGTERM → Router.Shutdown(10s timeout)
T+1s    Router waits...
T+2s    Router waits...
...
T+11s   TIMEOUT REACHED! (10s elapsed)
        ├─ Router.Shutdown() returns with error
        ├─ FX logs error
        ├─ Connection FORCE-CLOSED
        └─ User receives: Connection reset ❌

T+15s   Handler completes (too late)
        └─ Response discarded
```

**Result**: User gets connection error after 11 seconds

---

## Best Practices to Ensure Response Delivery

### 1. Set Appropriate Timeout

```go
// For APIs with long operations
shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
```

### 2. Use Request Timeouts

```go
func Handler(c *fiber.Ctx) error {
    // Set maximum request duration
    ctx, cancel := context.WithTimeout(c.UserContext(), 25*time.Second)
    defer cancel()

    // Use timeout context
    err := db.WithTx(ctx, func(tx pgx.Tx) error {
        // Operations limited to 25s
    })
}
```

### 3. Check Context Cancellation

```go
func LongHandler(c *fiber.Ctx) error {
    ctx := c.UserContext()

    for i := 0; i < 100; i++ {
        // Check if shutdown signal received
        select {
        case <-ctx.Done():
            // Shutdown in progress - return quickly
            return c.Status(503).JSON(fiber.Map{
                "error": "Service shutting down",
            })
        default:
            // Continue processing
            processItem(i)
        }
    }
}
```

---

## Summary

### Question 1: Database Continues During Router Shutdown?

**YES** ✅

- Router.Shutdown() **waits** for handlers to complete
- Database operations **continue normally**
- Handlers can **finish their work**
- Database remains **available** during shutdown
- Only when handlers complete does shutdown proceed

### Question 2: Response Reaches User?

**YES** ✅

- Router.Shutdown() waits for **complete response delivery**
- Not just handler completion, but also:
  - Response generation ✓
  - Response transmission ✓
  - Connection closure ✓
- User receives **full response** before shutdown completes
- Only if **timeout** is reached does connection force-close

### The Guarantee

```
Router shutdown guarantees:
├─ Handler completes execution
├─ Database operations finish
├─ Response generated
├─ Response transmitted to user
└─ Only then: shutdown proceeds

As long as: Total time < Shutdown timeout (10s)
```

### Why Current Order Is Correct

```
1. Router OnStop starts
   ├─ Waits for handlers (up to 10s)
   ├─ Database available during wait
   └─ Responses delivered during wait

2. Router OnStop completes
   └─ All handlers finished, responses sent

3. Database OnStop starts
   └─ Safe to close (no handlers running)

Result: Zero failed requests, all responses delivered ✓
```

---

## Monitoring Response Delivery

### Logs to Watch

**Successful response delivery:**
```
INFO  Shutdown signal received
INFO  Router adapter shutdown initiated
INFO  Handler /generate-report completed in 8.5s
INFO  Response sent: 200 OK, 5MB
INFO  Connection closed gracefully
INFO  Router adapter shutdown complete
```

**Timeout scenario:**
```
INFO  Shutdown signal received
INFO  Router adapter shutdown initiated
WARN  Handler /slow-operation timeout after 10s
WARN  Connection force-closed
ERROR Router adapter shutdown: context deadline exceeded
```

### Metrics to Track

- Average handler completion time during shutdown
- Number of connections force-closed due to timeout
- Number of successful response deliveries during shutdown
- Shutdown duration distribution

---

## Conclusion

Both database execution AND response delivery continue during router shutdown. The shutdown process is designed to ensure graceful completion of all in-flight work, not abrupt termination.

**Router shutdown = Graceful drain, not immediate stop!**
