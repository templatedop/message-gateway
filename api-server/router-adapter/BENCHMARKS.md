# Router Adapter Performance Benchmarks

Comprehensive performance analysis of all four supported web frameworks: Gin, Fiber, Echo, and net/http.

**Test Environment:**
- CPU: Intel(R) Xeon(R) CPU @ 2.60GHz
- Architecture: linux/amd64
- Go Version: 1.24.7
- Test Mode: Release (Gin in ReleaseMode)

## Table of Contents

- [Executive Summary](#executive-summary)
- [Benchmark Results](#benchmark-results)
- [Performance Analysis](#performance-analysis)
- [Router Recommendations](#router-recommendations)
- [Use Case Matrix](#use-case-matrix)

---

## Executive Summary

### 🏆 Performance Winners

| Category | Winner | Runner-Up |
|----------|--------|-----------|
| **Fastest Startup** | net/http | Gin |
| **Lowest Latency** | Fiber | Gin |
| **Best Throughput** | Fiber | net/http |
| **Lowest Memory** | net/http | Fiber |
| **Middleware Efficiency** | Fiber | net/http |
| **Path Parameters** | Fiber | Gin |
| **Real-World Performance** | Fiber | net/http |

### 📊 Quick Comparison (Lower is Better)

```
Adapter Creation (ns/op):
net/http:  293 ████
Gin:       682 █████████
Fiber:    4552 ██████████████████████████████████████████████
Echo:   125449 (initialization overhead)

Simple Request (ns/op):
Fiber:    1301 ████████████
Gin:      1899 ████████████████
Echo:     2201 ██████████████████
net/http: 2321 ███████████████████

Memory per Request (B/op):
Fiber:    1057 ██████████
Gin:      1450 ██████████████
net/http: 1763 █████████████████
Echo:     1747 █████████████████
```

---

## Benchmark Results

### 1. Adapter Creation

Time to create a new router adapter instance.

| Framework | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| **net/http** | **293** | **368** | **4** |
| Gin | 682 | 832 | 7 |
| Fiber | 4,552 | 5,252 | 41 |
| Echo | 125,449 | 3,283 | 46 |

**Analysis:**
- net/http is **2.3x faster** than Gin and uses **57% less memory**
- Echo has significant initialization overhead due to template compilation
- Fiber's overhead is acceptable for long-running servers (microseconds vs milliseconds)

### 2. Simple Request Handling

Performance for handling a basic GET request without path parameters.

| Framework | ns/op | B/op | allocs/op | req/s (est) |
|-----------|-------|------|-----------|-------------|
| **Fiber** | **1,301** | **1,057** | **11** | **769k** |
| Gin | 1,899 | 1,450 | 15 | 527k |
| Echo | 2,201 | 1,747 | 19 | 454k |
| net/http | 2,321 | 1,763 | 20 | 431k |

**Analysis:**
- Fiber is **31% faster** than Gin and **44% faster** than net/http
- Fiber uses **27% less memory** per request than Gin
- All frameworks handle 400k+ req/s - excellent for most use cases

### 3. Path Parameter Handling

Performance when extracting path parameters (e.g., `/users/:id/posts/:postid`).

| Framework | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| **Fiber** | **1,255** | **1,057** | **11** |
| Gin | 1,841 | 1,450 | 15 |
| Echo | 2,516 | 2,035 | 20 |
| net/http | 2,898 | 2,436 | 23 |

**Analysis:**
- Fiber is **32% faster** than Gin for parameterized routes
- net/http has higher overhead due to regex-based routing
- Echo allocates more memory for parameter extraction

### 4. Middleware Overhead

Performance impact of middleware chain (framework-agnostic middleware).

#### Single Middleware

| Framework | ns/op | B/op | allocs/op | Overhead |
|-----------|-------|------|-----------|----------|
| **Fiber** | **1,218** | **1,057** | **11** | **-6%** |
| Gin | 2,281 | 1,780 | 22 | +20% |
| net/http | 2,666 | 1,851 | 24 | +15% |
| Echo | 2,834 | 2,115 | 26 | +29% |

#### Five Middlewares

| Framework | ns/op | B/op | allocs/op | Overhead |
|-----------|-------|------|-----------|----------|
| **Fiber** | **1,350** | **1,057** | **11** | **+4%** |
| net/http | 3,308 | 2,148 | 36 | +43% |
| Gin | 3,954 | 3,034 | 50 | +108% |
| Echo | 4,704 | 3,526 | 54 | +114% |

**Analysis:**
- Fiber shows **near-zero overhead** for multiple middlewares
- Gin and Echo show linear growth with middleware count
- Fiber's fasthttp foundation provides superior middleware performance

### 5. Real-World API Scenario

Realistic API with:
- 3 middleware layers (logging, auth, CORS)
- 10 typical REST endpoints
- Mix of request types (simple, parameterized, POST)

| Framework | ns/op | B/op | allocs/op | Relative |
|-----------|-------|------|-----------|----------|
| **Fiber** | **1,293** | **1,057** | **11** | **1.0x** |
| net/http | 3,681 | 2,596 | 32 | 2.8x |
| Gin | 4,183 | 3,142 | 38 | 3.2x |
| Echo | 5,587 | 3,702 | 43 | 4.3x |

**Analysis:**
- Fiber is **2.8-4.3x faster** than other frameworks in production scenarios
- Memory usage: Fiber uses **66% less memory** than Gin
- Fiber maintains performance advantage under realistic conditions

### 6. RouterContext Operations

Performance of the framework-agnostic context abstraction.

| Operation | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| Creation | 682 | 656 | 7 |
| Set Param | 14 | 0 | 0 |
| Get Param | 9 | 0 | 0 |
| Query Param | 10 | 0 | 0 |
| Set Data | 41 | 0 | 0 |
| Get Data | 24 | 0 | 0 |
| JSON Response | 3,243 | 2,098 | 27 |

**Analysis:**
- RouterContext operations are **extremely fast** (< 50ns)
- Zero allocations for parameter and data operations
- JSON serialization dominates response time

---

## Performance Analysis

### Memory Efficiency

**Memory per Request (B/op):**

```
Fiber:      1,057 B  ████████████
Gin:        1,450 B  ████████████████
net/http:   1,763 B  ███████████████████
Echo:       1,747 B  ███████████████████
```

**Key Insights:**
- Fiber uses **27-40% less memory** than alternatives
- Memory efficiency crucial for high-concurrency servers
- Lower allocations reduce GC pressure

### Latency Comparison

**Average Request Latency (microseconds):**

| Framework | P50 | P95 (est) | P99 (est) |
|-----------|-----|-----------|-----------|
| Fiber | 1.3 | ~2.0 | ~3.0 |
| Gin | 1.9 | ~3.0 | ~4.5 |
| Echo | 2.2 | ~3.5 | ~5.2 |
| net/http | 2.3 | ~3.7 | ~5.5 |

### Throughput Estimates

Based on single-core performance:

| Framework | Req/s (single core) | Req/s (16 cores) |
|-----------|---------------------|------------------|
| Fiber | 770,000 | ~12.3M |
| Gin | 527,000 | ~8.4M |
| Echo | 454,000 | ~7.3M |
| net/http | 431,000 | ~6.9M |

**Real-world throughput** will be lower due to:
- Business logic overhead
- Database queries
- Network I/O
- External API calls

Expect **10-50k req/s** per instance for typical APIs.

---

## Router Recommendations

### 🚀 Fiber - Maximum Performance

**Use When:**
- Performance is the top priority
- Building high-throughput APIs (100k+ req/s)
- Need low latency (< 2ms p99)
- Handling many concurrent connections
- Microservices with strict SLA requirements
- WebSocket or Server-Sent Events (SSE)

**Advantages:**
- ✅ Fastest request handling (1.3μs)
- ✅ Lowest memory per request (1,057 B)
- ✅ Near-zero middleware overhead
- ✅ Built on fasthttp (optimized HTTP parser)
- ✅ Express.js-like API (familiar to Node.js developers)

**Trade-offs:**
- ⚠️ fasthttp incompatible with standard http.Handler
- ⚠️ Slightly higher startup time (4.5μs)
- ⚠️ Larger dependency tree

**Example Use Cases:**
- Real-time chat applications
- Gaming backends
- IoT data ingestion
- High-frequency trading APIs
- Proxy/gateway services

**Configuration:**
```go
cfg.Type = routeradapter.RouterTypeFiber
cfg.Fiber = &routeradapter.FiberConfig{
    Prefork:      false,  // Use for multi-core (SO_REUSEPORT)
    Concurrency:  256 * 1024,
    BodyLimit:    4 * 1024 * 1024,
}
```

---

### ⚖️ Gin - Balanced Choice

**Use When:**
- Need balance between performance and ecosystem
- Want extensive middleware library
- Team familiar with Gin
- Need good documentation and community
- Standard CRUD APIs
- Rapid prototyping

**Advantages:**
- ✅ Great performance (1.9μs)
- ✅ Huge ecosystem of middleware
- ✅ Excellent documentation
- ✅ Large community support
- ✅ Fast startup (682ns)
- ✅ Battle-tested in production

**Trade-offs:**
- ⚠️ Higher middleware overhead vs Fiber
- ⚠️ More allocations per request

**Example Use Cases:**
- REST APIs for web/mobile apps
- Internal business applications
- Content management systems
- E-commerce platforms
- Admin dashboards

**Configuration:**
```go
cfg.Type = routeradapter.RouterTypeGin
cfg.Gin = &routeradapter.GinConfig{
    Mode:                "release",
    RemoveExtraSlash:    true,
    ForwardedByClientIP: true,
}
```

---

### 🎯 Echo - Minimalist & Flexible

**Use When:**
- Want a minimalist framework
- Need HTTP/2 support out of the box
- Prefer clean, simple APIs
- Building straightforward web services
- Standard business applications

**Advantages:**
- ✅ Minimal and easy to understand
- ✅ Built-in HTTP/2 support
- ✅ Auto TLS with Let's Encrypt
- ✅ Centralized error handling
- ✅ Good for beginners

**Trade-offs:**
- ⚠️ Slower than Fiber and Gin
- ⚠️ Higher initialization time (125μs)
- ⚠️ Higher memory per request

**Example Use Cases:**
- Simple web services
- Proof-of-concepts
- Internal tools
- Learning projects
- Microservices with moderate load

**Configuration:**
```go
cfg.Type = routeradapter.RouterTypeEcho
cfg.Echo = &routeradapter.EchoConfig{
    Debug:      false,
    HideBanner: true,
}
```

---

### 🔧 net/http - Zero Dependencies

**Use When:**
- Want zero third-party dependencies
- Building libraries or frameworks
- Need maximum portability
- Strict security/compliance requirements
- Minimal attack surface desired
- Educational purposes

**Advantages:**
- ✅ Fastest startup (293ns)
- ✅ Zero dependencies
- ✅ Part of Go standard library
- ✅ Maximum stability
- ✅ Smallest binary size

**Trade-offs:**
- ⚠️ Slower request handling than Fiber/Gin
- ⚠️ More manual routing code needed
- ⚠️ Fewer built-in features
- ⚠️ Regex-based routing has overhead

**Example Use Cases:**
- Libraries and frameworks
- Embedded systems
- Security-critical applications
- Long-term maintenance projects
- Government/compliance-heavy sectors

**Configuration:**
```go
cfg.Type = routeradapter.RouterTypeNetHTTP
cfg.NetHTTP = &routeradapter.NetHTTPConfig{
    EnableHTTP2: true,
}
```

---

## Use Case Matrix

### By Traffic Volume

| Traffic Level | Recommended | Alternative |
|---------------|-------------|-------------|
| < 1k req/s | Any | net/http for simplicity |
| 1k-10k req/s | Gin | Echo |
| 10k-50k req/s | Fiber | Gin |
| 50k-100k req/s | Fiber | - |
| > 100k req/s | Fiber + Prefork | - |

### By Application Type

| Application Type | Primary Choice | Secondary Choice |
|------------------|----------------|------------------|
| REST API (CRUD) | Gin | Fiber |
| GraphQL Server | Fiber | Gin |
| Microservice | Fiber | net/http |
| Monolith | Gin | Echo |
| Real-time (WebSocket) | Fiber | - |
| Static Site | net/http | Echo |
| Proxy/Gateway | Fiber | net/http |
| Admin Panel | Gin | Echo |
| Mobile Backend | Fiber | Gin |
| IoT Platform | Fiber | net/http |

### By Team & Project Factors

| Factor | Recommended Router |
|--------|-------------------|
| Team new to Go | Gin (best docs) |
| Performance critical | Fiber |
| Security audit required | net/http |
| Quick prototype | Gin or Echo |
| Long-term maintenance | Gin or net/http |
| Minimize dependencies | net/http |
| JavaScript background | Fiber (Express-like) |
| Large ecosystem needed | Gin |

### By Technical Requirements

| Requirement | Best Choice | Reason |
|-------------|-------------|--------|
| HTTP/2 Support | Echo | Built-in |
| HTTP/3 (QUIC) | Fiber | Via external lib |
| WebSocket | Fiber | Native fasthttp support |
| Server-Sent Events | Fiber | Low overhead |
| Large File Upload | Fiber | Streaming support |
| gRPC Gateway | Gin | Better ecosystem |
| TLS/Auto-Cert | Echo | Built-in Let's Encrypt |
| Graceful Shutdown | All | All support context-based |

---

## Performance Optimization Tips

### General (All Routers)

1. **Use Release Mode**
   ```go
   gin.SetMode(gin.ReleaseMode)  // For Gin
   ```

2. **Enable Response Compression**
   - Use gzip middleware for text responses
   - Benchmark compression level (1-9)

3. **Connection Pooling**
   - Reuse HTTP connections
   - Configure keep-alive timeouts

4. **Reduce Allocations**
   - Pool common objects (sync.Pool)
   - Avoid unnecessary conversions

### Fiber-Specific

1. **Enable Prefork** (multi-core)
   ```go
   cfg.Fiber.Prefork = true  // SO_REUSEPORT
   ```

2. **Tune Concurrency**
   ```go
   cfg.Fiber.Concurrency = 1024 * 1024  // Max connections
   ```

3. **Optimize Body Limit**
   ```go
   cfg.Fiber.BodyLimit = 1 * 1024 * 1024  // 1MB
   ```

### Gin-Specific

1. **Disable Logging in Production**
   - Remove gin.Logger() middleware
   - Use structured logging

2. **Use Trusted Proxies**
   ```go
   cfg.Gin.TrustedProxies = []string{"10.0.0.0/8"}
   ```

### net/http-Specific

1. **Set Timeouts**
   ```go
   cfg.ReadTimeout = 5 * time.Second
   cfg.WriteTimeout = 10 * time.Second
   cfg.IdleTimeout = 120 * time.Second
   ```

2. **Optimize MaxHeaderBytes**
   ```go
   cfg.MaxHeaderBytes = 1 << 20  // 1MB
   ```

---

## Running Benchmarks

### Quick Benchmark

```bash
# Run all benchmarks
go test -bench=. -benchmem ./api-server/router-adapter/

# Specific benchmark
go test -bench=BenchmarkSimpleRequest -benchmem ./api-server/router-adapter/

# With longer run time for accuracy
go test -bench=. -benchmem -benchtime=5s ./api-server/router-adapter/
```

### Comparison

```bash
# Save baseline
go test -bench=. -benchmem ./api-server/router-adapter/ > old.txt

# Make changes...

# Compare
go test -bench=. -benchmem ./api-server/router-adapter/ > new.txt
benchcmp old.txt new.txt
```

### CPU Profiling

```bash
go test -bench=BenchmarkRealWorldScenario -cpuprofile=cpu.prof ./api-server/router-adapter/
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -bench=BenchmarkMemoryUsage -memprofile=mem.prof ./api-server/router-adapter/
go tool pprof mem.prof
```

---

## Conclusion

### Quick Decision Guide

**Choose Fiber if:**
- Performance is critical
- Handling high concurrency
- Need lowest latency

**Choose Gin if:**
- Want ecosystem and community
- Need balance of performance and features
- Team has Gin experience

**Choose Echo if:**
- Want minimalist framework
- Need HTTP/2 out of box
- Building standard web services

**Choose net/http if:**
- Want zero dependencies
- Need maximum portability
- Security/compliance is paramount

### Final Recommendations

For **most production applications**, we recommend:

1. **Fiber** - If performance matters and you can accept the dependencies
2. **Gin** - For general-purpose APIs with good ecosystem support
3. **net/http** - For libraries, frameworks, or security-critical applications
4. **Echo** - For simple services or when you prefer minimalism

All four routers are production-ready and perform excellently. The "best" choice depends on your specific requirements, team expertise, and project constraints.

---

*Benchmarks performed on: 2025-11-16*
*Router Adapter Version: 1.0.0*
*Go Version: 1.24.7*
