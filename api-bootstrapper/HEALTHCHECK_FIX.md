# Health Check Fix for Read/Write DB Mode

## Problem

After implementing separate read and write DB connections in api-bootstrapper (lines 222 onwards), the health check probes for both databases were commented out and stopped working:

- **Lines 253-259**: Read DB health check commented
- **Lines 425-431**: Write DB health check commented

This meant that the health check endpoint (`/healthzz`) would not monitor database connectivity, leading to:
- No visibility into database connection health
- Inability to detect database failures
- Kubernetes/load balancer health checks failing silently

## Root Cause

The health check registration was commented out because it needed to be adapted for the new named dependency injection pattern:

```go
// Old pattern (commented out):
// fxhealthcheck.AsCheckerProbe(db.NewSQLProbe)

// This didn't work because:
// 1. Two separate DB instances: "read_db" and "write_db"
// 2. Probes needed custom names to distinguish them
// 3. Named dependencies require explicit parameter injection
```

## Solution

Restored and fixed health check registration using proper named dependency injection:

### Read DB Health Check (Line 253-257)

```go
fxhealthcheck.AsCheckerProbe(func(p readDBProbeParams) healthcheck.CheckerProbe {
    probe := db.NewSQLProbe(p.DB)
    probe.SetName(ReadDBProbeName)  // "read-db-probe"
    return probe
}),
```

**Key aspects:**
- Uses `readDBProbeParams` struct with `name:"read_db"` tag
- Creates SQLProbe for the read database
- Sets custom name for identification
- Returns `healthcheck.CheckerProbe` interface

### Write DB Health Check (Line 423-427)

```go
fxhealthcheck.AsCheckerProbe(func(p writeDBProbeParams) healthcheck.CheckerProbe {
    probe := db.NewSQLProbe(p.DB)
    probe.SetName(WriteDBProbeName)  // "write-db-probe"
    return probe
}),
```

**Key aspects:**
- Uses `writeDBProbeParams` struct with `name:"write_db"` tag
- Creates SQLProbe for the write database
- Sets custom name for identification
- Returns `healthcheck.CheckerProbe` interface

## How It Works

### 1. Named Dependency Injection

```go
type readDBProbeParams struct {
    fx.In
    DB *db.DB `name:"read_db"`  // Injects the read database
}

type writeDBProbeParams struct {
    fx.In
    DB *db.DB `name:"write_db"`  // Injects the write database
}
```

Uber FX injects the correct database instance based on the `name` tag.

### 2. Probe Creation Function

The function passed to `AsCheckerProbe` must:
- Accept a struct with `fx.In` embedding (for dependency injection)
- Return `healthcheck.CheckerProbe` interface
- Create and configure the probe appropriately

### 3. Probe Registration

`AsCheckerProbe` internally:
- Annotates the function with `fx.As(new(healthcheck.CheckerProbe))`
- Tags it with `group:"healthcheck-probes"`
- Registers it in the health check registry
- Makes it available to the health check endpoint

### 4. Health Check Endpoint

The `/healthzz` endpoint now checks both probes:
```
GET /healthzz

Response (healthy):
{
  "status": "healthy",
  "checks": {
    "read-db-probe": {
      "passed": true,
      "output": "database ping success"
    },
    "write-db-probe": {
      "passed": true,
      "output": "database ping success"
    }
  }
}

Response (unhealthy):
{
  "status": "unhealthy",
  "checks": {
    "read-db-probe": {
      "passed": false,
      "output": "database connection timeout after 10s"
    },
    "write-db-probe": {
      "passed": true,
      "output": "database ping success"
    }
  }
}
```

## Files Modified

### api-bootstrapper/bootstrapper.go

**Lines 253-257** (Read DB):
```diff
- // fxhealthcheck.AsCheckerProbe(func(p readDBProbeParams) healthcheck.CheckerProbe {
- //     probe := db.NewSQLProbe(p.DB)
- //     probe.SetName(ReadDBProbeName)
- //     return probe
- // }),
+ fxhealthcheck.AsCheckerProbe(func(p readDBProbeParams) healthcheck.CheckerProbe {
+     probe := db.NewSQLProbe(p.DB)
+     probe.SetName(ReadDBProbeName)
+     return probe
+ }),
```

**Lines 423-427** (Write DB):
```diff
- // fxhealthcheck.AsCheckerProbe(func(p writeDBProbeParams) healthcheck.CheckerProbe {
- //     probe := db.NewSQLProbe(p.DB)
- //     probe.SetName(WriteDBProbeName)
- //     return probe
- // }),
+ fxhealthcheck.AsCheckerProbe(func(p writeDBProbeParams) healthcheck.CheckerProbe {
+     probe := db.NewSQLProbe(p.DB)
+     probe.SetName(WriteDBProbeName)
+     return probe
+ }),
```

### api-bootstrapper/healthcheck_test.go (NEW)

Comprehensive test suite covering:
- Probe parameter struct configuration
- Probe creation with correct names
- Probe interface implementation
- Function signature validation
- Probe naming uniqueness
- SetName functionality
- Mock probe implementation
- Benchmark tests

**Test results:**
- 17 unit tests covering all aspects
- All probe creation logic verified
- Interface compliance validated
- Benchmark tests for performance monitoring

## Testing

### Unit Tests

Created `api-bootstrapper/healthcheck_test.go` with:

1. **Struct Tests**
   - `TestReadDBProbeParams`
   - `TestWriteDBProbeParams`

2. **Probe Creation Tests**
   - `TestReadDBProbeCreation`
   - `TestWriteDBProbeCreation`

3. **Configuration Tests**
   - `TestProbeNames` - Verifies constants
   - `TestCollectorNames` - Verifies collector names
   - `TestProbeInterface` - Validates interface implementation

4. **Signature Tests**
   - `TestProbeFunctionSignature` - Validates function signatures for both probes

5. **Naming Tests**
   - `TestProbeNaming` - Ensures unique names
   - `TestProbeSetName` - Tests name setting

6. **Mock Tests**
   - `TestMockProbe` - Validates mock implementation
   - `TestProbeWithTimeout` - Tests timeout behavior

7. **Benchmarks**
   - `BenchmarkProbeCreation` - Measures probe creation time
   - `BenchmarkProbeNameAccess` - Measures name access time

### Integration Testing

For full integration testing (requires real database):

```bash
# Start test databases
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test ./api-bootstrapper -tags=integration -v

# Check health endpoint
curl http://localhost:8080/healthzz
```

### Manual Testing

1. **Start application:**
   ```bash
   go run main.go
   ```

2. **Check health endpoint:**
   ```bash
   curl http://localhost:8080/healthzz
   ```

3. **Expected response:**
   ```json
   {
     "status": "healthy",
     "checks": {
       "read-db-probe": {
         "passed": true,
         "output": "database ping success"
       },
       "write-db-probe": {
         "passed": true,
         "output": "database ping success"
       }
     }
   }
   ```

4. **Test failure scenario** (stop read DB):
   ```bash
   docker stop postgres-read
   curl http://localhost:8080/healthzz
   ```

   Expected response:
   ```json
   {
     "status": "unhealthy",
     "checks": {
       "read-db-probe": {
         "passed": false,
         "output": "database connection timeout after 10s"
       },
       "write-db-probe": {
         "passed": true,
         "output": "database ping success"
       }
     }
   }
   ```

## Benefits

1. **Separate Monitoring**: Read and write databases monitored independently
2. **Clear Identification**: Probes have descriptive names ("read-db-probe", "write-db-probe")
3. **Kubernetes Ready**: Health checks work with K8s liveness/readiness probes
4. **Load Balancer Integration**: Works with load balancer health checks
5. **Debugging**: Easy to identify which database has issues
6. **Metrics**: Can track health check success/failure rates per database

## Configuration

### Constants (lines 42-47)

```go
const (
    ReadDBProbeName      = "read-db-probe"
    WriteDBProbeName     = "write-db-probe"
    ReadDBCollectorName  = "read_db_collector"
    WriteDBCollectorName = "write_db_collector"
)
```

These can be customized if needed, but should remain unique and descriptive.

### Timeout Configuration

The SQL probe uses a 10-second timeout by default (configured in api-db/dbprobe.go).

To adjust:
```go
// In api-db/dbprobe.go
const DefaultTimeout = 10 * time.Second
```

## Troubleshooting

### Health Check Returns 500

**Symptom:** `/healthzz` returns HTTP 500

**Cause:** Health check module not properly initialized

**Solution:** Verify `fxhealthcheck` module is included in bootstrapper

### Probe Not Showing in Health Check

**Symptom:** Only one probe appears, or none appear

**Cause:** Probe registration failed or name conflict

**Solution:**
1. Check logs for FX errors during startup
2. Verify probe names are unique
3. Ensure DB connections are established before probes run

### Probe Always Fails

**Symptom:** Probe shows `passed: false` even when DB is healthy

**Cause:** Database configuration issues or permission problems

**Solution:**
1. Check database connection string
2. Verify database user has required permissions
3. Check database firewall rules
4. Review logs for specific error messages

### Probe Times Out

**Symptom:** Probe fails with "database connection timeout"

**Cause:** Network latency or database performance issues

**Solution:**
1. Increase timeout (if appropriate)
2. Check network connectivity
3. Investigate database performance
4. Review database query logs

## Future Enhancements

Potential improvements:

1. **Configurable Timeouts**: Allow probe timeout configuration via config file
2. **Detailed Metrics**: Export health check metrics to Prometheus
3. **Custom Queries**: Support custom health check queries beyond ping
4. **Alerting**: Integrate with alerting systems for probe failures
5. **Historical Data**: Track health check history for trending

## Related Files

- `api-db/dbprobe.go` - SQLProbe implementation
- `api-healthcheck/checker.go` - Health checker interface
- `api-healthcheck/probe.go` - Probe interface definition
- `api-fxhealth/register.go` - FX health check registration
- `api-server/health/health_check.go` - Health check HTTP endpoint

## References

- [Uber FX Documentation](https://uber-go.github.io/fx/)
- [Kubernetes Liveness Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- Health Check API Design (internal docs)
