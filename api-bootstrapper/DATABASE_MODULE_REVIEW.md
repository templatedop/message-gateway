# Database Module Review and Recommendations

## Current Implementation Analysis

### Overview

The api-bootstrapper uses FX annotations to manage two separate database connections:
- **Write DB** (fxDB): For write operations
- **Read DB** (FxReadDB): For read operations (read replicas)

This implements a **Read-Write Splitting** pattern for database scalability.

---

## Current Implementation

### 1. Write DB Module (`fxDB`)

**Location**: `api-bootstrapper/bootstrapper.go:338-381`

```go
var fxDB = fx.Module(
    "Write DBModule",
    fx.Provide(
        fx.Annotated{
            Name:   "write_config",
            Target: dbconfig,
        },
        fx.Annotated{
            Name:   "write_prepared_config",
            Target: db.NewDefaultDbFactory().NewPreparedDBConfig,
        },
        fx.Annotated{
            Name: "write_db",
            Target: func(params struct {
                fx.In
                Config    db.DBConfig `name:"write_config"`
                Osdktrace *otelsdktrace.TracerProvider
                Registry  *prometheus.Registry
            }) (*db.DB, error) {
                factory := db.NewDefaultDbFactory()
                factory.SetCollectorName(WriteDBCollectorName)
                return factory.CreateConnection(&params.Config, params.Osdktrace, params.Registry)
            },
        },
        // Bridge provider: expose named write_db as default *db.DB
        func(p struct {
            fx.In
            Write *db.DB `name:"write_db"`
        }) *db.DB {
            return p.Write
        },
    ),
    fx.Invoke(dblifecycle),
)
```

**Features**:
- ✅ Context-aware lifecycle (signal propagation)
- ✅ Graceful connection draining (5s timeout)
- ✅ Detailed logging with connection metrics
- ✅ Bridge provider for default injection
- ✅ **ACTIVE** in bootstrapper options

### 2. Read DB Module (`FxReadDB`)

**Location**: `api-bootstrapper/bootstrapper.go:222-260`

```go
var FxReadDB = fx.Module(
    "Read DBModule",
    fx.Provide(
        fx.Annotated{
            Name:   "read_config",
            Target: dbreadconfig,
        },
        fx.Annotated{
            Name:   "read_prepared_config",
            Target: db.NewDefaultDbFactory().NewPreparedDBConfig,
        },
        fx.Annotated{
            Name: "read_db",
            Target: func(params struct {
                fx.In
                Config    db.DBConfig `name:"read_config"`
                Osdktrace *otelsdktrace.TracerProvider
                Registry  *prometheus.Registry
            }) (*db.DB, error) {
                factory := db.NewDefaultDbFactory()
                factory.SetCollectorName(ReadDBCollectorName)
                return factory.CreateConnection(&params.Config, params.Osdktrace, params.Registry)
            },
        },
    ),
    fx.Invoke(readdblifecycle),
)
```

**Features**:
- ❌ NO context in lifecycle params
- ❌ NO graceful connection draining
- ❌ OLD shutdown implementation (immediate close)
- ❌ NO bridge provider
- ❌ **NOT ACTIVE** in bootstrapper options (not included)

---

## Issues Identified

### 🔴 Critical Issue #1: FxReadDB Not Active

**Problem**:
```go
func New() *Bootstrapper {
    return &Bootstrapper{
        context: context.Background(),
        options: []fx.Option{
            fxconfig,
            fxlog,
            fxDB,              // ✅ Write DB included
            // FxReadDB,       // ❌ Read DB NOT included!
            fxRouterAdapter,
            fxTrace,
            fxMetrics,
        },
    }
}
```

**Impact**:
- Read DB connection never created
- Applications cannot use read replicas
- All queries go to write DB (no load distribution)
- Read-write splitting not functional

**Severity**: **HIGH** - Feature not working

---

### 🔴 Critical Issue #2: Read DB Missing Graceful Shutdown

**Current Implementation** (`bootstrapper.go:268-296`):

```go
type readDBLifecycleParams struct {
    fx.In
    DB *db.DB `name:"read_db"`  // ❌ NO Context!
    LC fx.Lifecycle
}

func readdblifecycle(p readDBLifecycleParams) {
    p.LC.Append(
        fx.Hook{
            OnStart: func(ctx context.Context) error {
                err := p.DB.Ping()  // ❌ Not context-aware!
                if err != nil {
                    return err
                }
                return nil
            },
            OnStop: func(ctx context.Context) error {
                // ❌ Immediate close - no draining!
                p.DB.Close()
                return nil
            },
        },
    )
}
```

**Problems**:
- No context propagation (can't detect shutdown signals)
- No graceful connection draining
- Immediate close (same issue we fixed for write DB)
- No detailed logging during shutdown
- Inconsistent with write DB implementation

**Severity**: **HIGH** - Failed requests during shutdown

---

### 🟡 Medium Issue #3: Code Duplication

**Duplicated Code**:
- Config functions: `dbconfig()` vs `dbreadconfig()` (almost identical)
- Provider functions: Anonymous functions in both modules (identical structure)
- Lifecycle functions: `dblifecycle()` vs `readdblifecycle()` (should be same)

**Impact**:
- Harder to maintain
- Bug fixes must be applied twice
- Inconsistencies (as seen with graceful shutdown)

**Severity**: **MEDIUM** - Maintenance burden

---

### 🟡 Medium Issue #4: No Read DB Bridge Provider

**Write DB has**:
```go
func(p struct {
    fx.In
    Write *db.DB `name:"write_db"`
}) *db.DB {
    return p.Write
}
```

**Read DB doesn't have** bridge provider

**Impact**:
- Consumers must always use `name:"read_db"` annotation
- Cannot inject default DB for read-only repositories
- Less flexible dependency injection

**Severity**: **MEDIUM** - Usability issue

---

### 🟢 Minor Issue #5: Inconsistent Naming

- Write DB module: `fxDB` (lowercase 'fx')
- Read DB module: `FxReadDB` (uppercase 'Fx')

**Impact**: Minor style inconsistency

---

## Recommended Improvements

### ✅ Recommendation #1: Unified Database Module with Strategy Pattern

**Better Approach**: Create a single, reusable database module factory

```go
// dbModuleParams holds parameters for creating a DB module
type dbModuleParams struct {
    Name              string
    ConfigReader      func(*config.Config) db.DBConfig
    CollectorName     string
    IsDefaultProvider bool // If true, also provide as unnamed *db.DB
}

// createDBModule creates a database module with given parameters
func createDBModule(params dbModuleParams) fx.Option {
    configName := params.Name + "_config"
    preparedConfigName := params.Name + "_prepared_config"
    dbName := params.Name + "_db"

    providers := []interface{}{
        // Config provider
        fx.Annotated{
            Name:   configName,
            Target: params.ConfigReader,
        },

        // Prepared config provider
        fx.Annotated{
            Name:   preparedConfigName,
            Target: db.NewDefaultDbFactory().NewPreparedDBConfig,
        },

        // DB connection provider
        fx.Annotated{
            Name: dbName,
            Target: func(p struct {
                fx.In
                Config    db.DBConfig `name:"<<CONFIG_NAME>>"`  // Dynamically set
                Osdktrace *otelsdktrace.TracerProvider
                Registry  *prometheus.Registry
            }) (*db.DB, error) {
                factory := db.NewDefaultDbFactory()
                factory.SetCollectorName(params.CollectorName)
                return factory.CreateConnection(&p.Config, p.Osdktrace, p.Registry)
            },
        },
    }

    // Add bridge provider if requested
    if params.IsDefaultProvider {
        providers = append(providers, func(p struct {
            fx.In
            DB *db.DB `name:"<<DB_NAME>>"`  // Dynamically set
        }) *db.DB {
            return p.DB
        })
    }

    return fx.Module(
        params.Name+" DBModule",
        fx.Provide(providers...),
        fx.Invoke(createDBLifecycle(dbName, params.Name)),
    )
}

// createDBLifecycle creates lifecycle hook for a named DB
func createDBLifecycle(dbName, moduleName string) interface{} {
    return func(ctx context.Context, db *db.DB `name:"<<DB_NAME>>"`, lc fx.Lifecycle) {
        lc.Append(fx.Hook{
            OnStart: func(startCtx context.Context) error {
                logger := log.GetBaseLoggerInstance().ToZerolog()
                logger.Info().
                    Str("module", moduleName).
                    Msg("Starting database module")

                // Context-aware ping
                if err := db.PingContext(startCtx); err != nil {
                    return err
                }

                logger.Info().
                    Str("module", moduleName).
                    Msg("Successfully connected to database")
                return nil
            },
            OnStop: func(stopCtx context.Context) error {
                logger := log.GetBaseLoggerInstance().ToZerolog()

                // Graceful connection draining (same as write DB)
                if count := db.Stat(); count != nil {
                    logger.Info().
                        Str("module", moduleName).
                        Int32("total_conns", count.TotalConns()).
                        Int32("idle_conns", count.IdleConns()).
                        Int32("acquired_conns", count.AcquiredConns()).
                        Msg("Database connection stats at shutdown start")
                }

                drainTimeout := 5 * time.Second
                drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
                defer cancel()

                logger.Info().
                    Str("module", moduleName).
                    Dur("drain_timeout", drainTimeout).
                    Msg("Waiting for active database connections to drain...")

                ticker := time.NewTicker(100 * time.Millisecond)
                defer ticker.Stop()

                for {
                    select {
                    case <-drainCtx.Done():
                        if count := db.Stat(); count != nil {
                            logger.Warn().
                                Str("module", moduleName).
                                Int32("remaining_acquired", count.AcquiredConns()).
                                Msg("Drain timeout reached, forcing database closure")
                        }
                        goto closeDB

                    case <-ticker.C:
                        if count := db.Stat(); count != nil {
                            if count.AcquiredConns() == 0 {
                                logger.Info().
                                    Str("module", moduleName).
                                    Msg("All database connections drained successfully")
                                goto closeDB
                            }
                        }
                    }
                }

            closeDB:
                db.Close()

                if count := db.Stat(); count != nil {
                    logger.Info().
                        Str("module", moduleName).
                        Int32("final_total_conns", count.TotalConns()).
                        Msg("Database connection pool closed")
                }

                logger.Info().
                    Str("module", moduleName).
                    Msg("Database shutdown complete")
                return nil
            },
        })
    }
}

// Usage:
var fxWriteDB = createDBModule(dbModuleParams{
    Name:              "write",
    ConfigReader:      dbconfig,
    CollectorName:     WriteDBCollectorName,
    IsDefaultProvider: true,  // Default *db.DB points to write DB
})

var fxReadDB = createDBModule(dbModuleParams{
    Name:              "read",
    ConfigReader:      dbreadconfig,
    CollectorName:     ReadDBCollectorName,
    IsDefaultProvider: false,  // Must use name:"read_db" explicitly
})
```

**Benefits**:
- ✅ Zero code duplication
- ✅ Consistent behavior (both get graceful shutdown)
- ✅ Easy to add more DB connections (e.g., analytics DB)
- ✅ Single place to fix bugs
- ✅ Type-safe and maintainable

---

### ✅ Recommendation #2: Simpler Fix (Minimal Changes)

If you don't want to refactor, just fix the critical issues:

#### Fix #1: Add FxReadDB to Options

```go
func New() *Bootstrapper {
    return &Bootstrapper{
        context: context.Background(),
        options: []fx.Option{
            fxconfig,
            fxlog,
            fxDB,              // Write DB
            FxReadDB,          // ✅ ADD Read DB
            fxRouterAdapter,
            fxTrace,
            fxMetrics,
        },
    }
}
```

#### Fix #2: Update Read DB Lifecycle

```go
type readDBLifecycleParams struct {
    fx.In
    Ctx context.Context  // ✅ ADD context
    DB  *db.DB `name:"read_db"`
    LC  fx.Lifecycle
}

func readdblifecycle(p readDBLifecycleParams) {
    p.LC.Append(
        fx.Hook{
            OnStart: func(ctx context.Context) error {
                logger := log.GetBaseLoggerInstance().ToZerolog()
                logger.Info().Str("module", "ReadDBModule").Msg("Starting read database module")

                // ✅ Context-aware ping
                err := p.DB.PingContext(ctx)
                if err != nil {
                    return err
                }

                logger.Info().Msg("Successfully connected to read database")
                return nil
            },
            OnStop: func(ctx context.Context) error {
                logger := log.GetBaseLoggerInstance().ToZerolog()

                // ✅ Copy full graceful shutdown from write DB
                if count := p.DB.Stat(); count != nil {
                    logger.Info().
                        Int32("total_conns", count.TotalConns()).
                        Int32("idle_conns", count.IdleConns()).
                        Int32("acquired_conns", count.AcquiredConns()).
                        Msg("Read database connection stats at shutdown start")
                }

                drainTimeout := 5 * time.Second
                drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
                defer cancel()

                logger.Info().
                    Dur("drain_timeout", drainTimeout).
                    Msg("Waiting for read database connections to drain...")

                ticker := time.NewTicker(100 * time.Millisecond)
                defer ticker.Stop()

                for {
                    select {
                    case <-drainCtx.Done():
                        if count := p.DB.Stat(); count != nil {
                            logger.Warn().
                                Int32("remaining_acquired", count.AcquiredConns()).
                                Msg("Read DB drain timeout, forcing closure")
                        }
                        goto closeDB

                    case <-ticker.C:
                        if count := p.DB.Stat(); count != nil {
                            if count.AcquiredConns() == 0 {
                                logger.Info().Msg("Read DB connections drained")
                                goto closeDB
                            }
                        }
                    }
                }

            closeDB:
                p.DB.Close()

                if count := p.DB.Stat(); count != nil {
                    logger.Info().
                        Int32("final_total_conns", count.TotalConns()).
                        Msg("Read database connection pool closed")
                }

                logger.Info().Msg("Read database shutdown complete")
                return nil
            },
        },
    )
}
```

---

### ✅ Recommendation #3: Repository-Level Abstraction

**Better Pattern**: Let repositories decide read vs write

```go
// In repository layer
type UserRepository struct {
    writeDB *db.DB
    readDB  *db.DB
}

func NewUserRepository(params struct {
    fx.In
    WriteDB *db.DB `name:"write_db"`
    ReadDB  *db.DB `name:"read_db"`
}) *UserRepository {
    return &UserRepository{
        writeDB: params.WriteDB,
        readDB:  params.ReadDB,
    }
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    // Use read DB for queries
    return r.readDB.QueryContext(ctx, "SELECT * FROM users WHERE id = $1", id)
}

func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
    // Use write DB for mutations
    return r.writeDB.WithTx(ctx, func(tx pgx.Tx) error {
        return tx.Exec(ctx, "INSERT INTO users (...) VALUES (...)")
    })
}
```

**Benefits**:
- Clear separation of concerns
- Type-safe dependency injection
- Easy to test (mock either DB)
- Flexible (can add caching, circuit breakers, etc.)

---

## Configuration Changes Needed

### Current Config

```yaml
db:
  username: "msggateway_rw_user"
  password: "DoPrw@123"
  host: "172.28.13.156"
  port: "5432"
  database: "msggatewaydb"
  schema: "msggateway"
  maxconns: 10
  minconns: 1
  # ... other settings
```

### ✅ Recommended Config (Read Replica Support)

```yaml
db:
  # Write DB (primary)
  username: "msggateway_rw_user"
  password: "DoPrw@123"
  host: "172.28.13.156"  # Primary DB
  port: "5432"
  database: "msggatewaydb"
  schema: "msggateway"
  maxconns: 10
  minconns: 1
  maxconnlifetime: 30
  maxconnidletime: 10
  healthcheckperiod: 5

  # Read DB (replica) - NEW
  read:
    username: "msggateway_ro_user"      # Read-only user
    password: "DoPro@123"
    host: "172.28.13.157"               # Read replica host
    port: "5432"
    database: "msggatewaydb"
    schema: "msggateway"
    maxconns: 20                        # More connections for reads
    minconns: 2
    maxconnlifetime: 30
    maxconnidletime: 10
    healthcheckperiod: 5
```

**Note**: If read replica has same host, use:
```yaml
db:
  read:
    host: "172.28.13.156"  # Same as primary
    username: "msggateway_ro_user"  # But use read-only user
```

---

## Summary Table

| Aspect | Current State | Issues | Recommended |
|--------|---------------|--------|-------------|
| **Write DB** | ✅ Fully functional | None | Keep as-is |
| **Read DB** | ⚠️ Defined but inactive | Not in options array | Add to options |
| **Read DB Lifecycle** | ❌ Missing features | No context, no graceful shutdown | Copy from write DB |
| **Code Duplication** | ❌ High | Duplicated logic | Create factory function |
| **Consistency** | ❌ Inconsistent | Read/Write behave differently | Unify implementation |
| **Naming** | ⚠️ Inconsistent | fxDB vs FxReadDB | Standardize (fxWriteDB, fxReadDB) |

---

## Action Plan

### 🔥 CRITICAL (Fix Immediately)

1. ✅ Add `FxReadDB` to bootstrapper options array
2. ✅ Update `readDBLifecycleParams` to include Context
3. ✅ Implement graceful shutdown for read DB (copy from write DB)

### 🟡 HIGH PRIORITY (Fix Soon)

4. ✅ Refactor to use factory pattern (eliminate duplication)
5. ✅ Add read DB bridge provider (if needed)
6. ✅ Standardize naming (fxWriteDB, fxReadDB)

### 🟢 NICE TO HAVE (Future)

7. ✅ Add read DB health check probe
8. ✅ Create repository-level abstraction
9. ✅ Add connection pool metrics per DB
10. ✅ Document read-write splitting strategy

---

## Conclusion

### Current Implementation: ⚠️ PARTIALLY CORRECT

**What's Working**:
- ✅ Write DB fully functional with graceful shutdown
- ✅ Correct use of FX annotations
- ✅ Proper dependency injection

**What's Broken**:
- ❌ Read DB not active (FxReadDB not in options)
- ❌ Read DB missing graceful shutdown
- ❌ Code duplication and inconsistency

### Better Model: YES

**Recommended Approach**:
1. **Short-term**: Fix critical issues (add FxReadDB, graceful shutdown)
2. **Long-term**: Refactor to factory pattern (eliminate duplication)

The FX annotation approach is correct, but the implementation is incomplete and inconsistent.

---

## Related Documentation

- `DATABASE_SHUTDOWN_COORDINATION.md` - Write DB graceful shutdown
- `SHUTDOWN_TIMING_COORDINATION.md` - Shutdown timing details
- `SHUTDOWN_COORDINATION_TEST_RESULTS.md` - Test verification

---

**Reviewed**: 2025-11-16
**Status**: Critical issues identified
**Priority**: HIGH - Fix read DB immediately
