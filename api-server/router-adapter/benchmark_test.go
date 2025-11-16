package routeradapter_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"MgApplication/api-server/route"
	"MgApplication/api-server/router-adapter"

	"github.com/gin-gonic/gin"

	// Import all adapters
	_ "MgApplication/api-server/router-adapter/echo"
	_ "MgApplication/api-server/router-adapter/fiber"
	_ "MgApplication/api-server/router-adapter/gin"
	_ "MgApplication/api-server/router-adapter/nethttp"
)

// Test data for benchmarks
var (
	dummyHandler = func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "hello"})
	}

	testPayload = map[string]interface{}{
		"user":    "testuser",
		"email":   "test@example.com",
		"active":  true,
		"balance": 1234.56,
		"items":   []string{"item1", "item2", "item3"},
	}
)

// BenchmarkAdapterCreation benchmarks adapter creation for each router type
func BenchmarkAdapterCreation(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType
			cfg.Port = 8080

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				adapter, err := routeradapter.NewRouterAdapter(cfg)
				if err != nil {
					b.Fatalf("Failed to create adapter: %v", err)
				}
				_ = adapter
			}
		})
	}
}

// BenchmarkRouteRegistration benchmarks route registration performance
func BenchmarkRouteRegistration(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				meta := route.Meta{
					Method: "GET",
					Path:   fmt.Sprintf("/route-%d", i),
					Func:   dummyHandler,
				}
				_ = adapter.RegisterRoute(meta)
			}
		})
	}
}

// BenchmarkSimpleRequest benchmarks simple GET request handling
func BenchmarkSimpleRequest(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)
			adapter.RegisterRoute(route.Meta{
				Method: "GET",
				Path:   "/hello",
				Func:   dummyHandler,
			})

			req := httptest.NewRequest("GET", "/hello", nil)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				adapter.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkPathParameters benchmarks request handling with path parameters
func BenchmarkPathParameters(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)
			adapter.RegisterRoute(route.Meta{
				Method: "GET",
				Path:   "/users/:id/posts/:postid",
				Func:   dummyHandler,
			})

			req := httptest.NewRequest("GET", "/users/123/posts/456", nil)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				adapter.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkMiddleware benchmarks middleware overhead
func BenchmarkMiddleware(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	// Simple middleware that adds a header
	middleware := func(ctx *routeradapter.RouterContext, next func() error) error {
		ctx.SetHeader("X-Test", "middleware")
		return next()
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType)+"/1-middleware", func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)
			adapter.RegisterMiddleware(middleware)
			adapter.RegisterRoute(route.Meta{
				Method: "GET",
				Path:   "/hello",
				Func:   dummyHandler,
			})

			req := httptest.NewRequest("GET", "/hello", nil)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				adapter.ServeHTTP(w, req)
			}
		})

		b.Run(string(routerType)+"/5-middlewares", func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)
			for j := 0; j < 5; j++ {
				adapter.RegisterMiddleware(middleware)
			}
			adapter.RegisterRoute(route.Meta{
				Method: "GET",
				Path:   "/hello",
				Func:   dummyHandler,
			})

			req := httptest.NewRequest("GET", "/hello", nil)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				adapter.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkJSONSerialization benchmarks JSON response performance
func BenchmarkJSONSerialization(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)
			adapter.RegisterRoute(route.Meta{
				Method: "GET",
				Path:   "/data",
				Func:   dummyHandler,
			})

			req := httptest.NewRequest("GET", "/data", nil)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				adapter.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	concurrencyLevels := []int{10, 100, 1000}

	for _, routerType := range routerTypes {
		for _, concurrency := range concurrencyLevels {
			name := fmt.Sprintf("%s/concurrent-%d", routerType, concurrency)
			b.Run(name, func(b *testing.B) {
				cfg := routeradapter.DefaultRouterConfig()
				cfg.Type = routerType

				adapter, _ := routeradapter.NewRouterAdapter(cfg)
				adapter.RegisterRoute(route.Meta{
					Method: "GET",
					Path:   "/hello",
					Func:   dummyHandler,
				})

				b.ResetTimer()
				b.ReportAllocs()

				b.RunParallel(func(pb *testing.PB) {
					req := httptest.NewRequest("GET", "/hello", nil)
					for pb.Next() {
						w := httptest.NewRecorder()
						adapter.ServeHTTP(w, req)
					}
				})
			})
		}
	}
}

// BenchmarkRouterContext benchmarks RouterContext operations
func BenchmarkRouterContext(b *testing.B) {
	req := httptest.NewRequest("GET", "/test?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	b.Run("creation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ctx := routeradapter.NewRouterContext(w, req)
			_ = ctx
		}
	})

	ctx := routeradapter.NewRouterContext(w, req)

	b.Run("set-param", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ctx.SetParam("id", "123")
		}
	})

	b.Run("get-param", func(b *testing.B) {
		ctx.SetParam("id", "123")
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ctx.Param("id")
		}
	})

	b.Run("query-param", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ctx.QueryParam("page")
		}
	})

	b.Run("set-data", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ctx.Set("user", "testuser")
		}
	})

	b.Run("get-data", func(b *testing.B) {
		ctx.Set("user", "testuser")
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ctx.Get("user")
		}
	})

	b.Run("json-response", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			ctx := routeradapter.NewRouterContext(w, req)
			_ = ctx.JSON(200, testPayload)
		}
	})
}

// BenchmarkComplexRouting benchmarks routing with many routes
func BenchmarkComplexRouting(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	numRoutes := []int{10, 100, 1000}

	for _, routerType := range routerTypes {
		for _, n := range numRoutes {
			name := fmt.Sprintf("%s/%d-routes", routerType, n)
			b.Run(name, func(b *testing.B) {
				cfg := routeradapter.DefaultRouterConfig()
				cfg.Type = routerType

				adapter, _ := routeradapter.NewRouterAdapter(cfg)

				// Register many routes
				for i := 0; i < n; i++ {
					adapter.RegisterRoute(route.Meta{
						Method: "GET",
						Path:   fmt.Sprintf("/route-%d", i),
						Func:   dummyHandler,
					})
				}

				// Test last route (worst case for linear search)
				req := httptest.NewRequest("GET", fmt.Sprintf("/route-%d", n-1), nil)

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					w := httptest.NewRecorder()
					adapter.ServeHTTP(w, req)
				}
			})
		}
	}
}

// BenchmarkRealWorldScenario benchmarks a realistic API scenario
func BenchmarkRealWorldScenario(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	// Middleware stack: logging, auth, CORS
	loggingMiddleware := func(ctx *routeradapter.RouterContext, next func() error) error {
		ctx.SetHeader("X-Request-ID", "test-123")
		return next()
	}

	authMiddleware := func(ctx *routeradapter.RouterContext, next func() error) error {
		ctx.Set("user", "authenticated-user")
		return next()
	}

	corsMiddleware := func(ctx *routeradapter.RouterContext, next func() error) error {
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		return next()
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)

			// Add middleware stack
			adapter.RegisterMiddleware(loggingMiddleware)
			adapter.RegisterMiddleware(authMiddleware)
			adapter.RegisterMiddleware(corsMiddleware)

			// Register typical API routes
			routes := []struct {
				method string
				path   string
			}{
				{"GET", "/api/v1/users"},
				{"GET", "/api/v1/users/:id"},
				{"POST", "/api/v1/users"},
				{"PUT", "/api/v1/users/:id"},
				{"DELETE", "/api/v1/users/:id"},
				{"GET", "/api/v1/posts"},
				{"GET", "/api/v1/posts/:id"},
				{"POST", "/api/v1/posts"},
				{"GET", "/health"},
				{"GET", "/metrics"},
			}

			for _, r := range routes {
				adapter.RegisterRoute(route.Meta{
					Method: r.method,
					Path:   r.path,
					Func:   dummyHandler,
				})
			}

			// Test mix of requests
			requests := []*http.Request{
				httptest.NewRequest("GET", "/api/v1/users", nil),
				httptest.NewRequest("GET", "/api/v1/users/123", nil),
				httptest.NewRequest("GET", "/api/v1/posts/456", nil),
				httptest.NewRequest("GET", "/health", nil),
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				req := requests[i%len(requests)]
				w := httptest.NewRecorder()
				adapter.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkServerStartup benchmarks server startup time
func BenchmarkServerStartup(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				cfg := routeradapter.DefaultRouterConfig()
				cfg.Type = routerType
				cfg.Port = 38888 + i // Use different ports

				adapter, _ := routeradapter.NewRouterAdapter(cfg)
				adapter.RegisterRoute(route.Meta{
					Method: "GET",
					Path:   "/health",
					Func:   dummyHandler,
				})

				b.StartTimer()
				err := adapter.Start(fmt.Sprintf(":%d", 38888+i))
				b.StopTimer()

				if err != nil {
					b.Fatalf("Failed to start: %v", err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				adapter.Shutdown(ctx)
				cancel()

				time.Sleep(50 * time.Millisecond) // Allow port to be released
			}
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory allocations per request
func BenchmarkMemoryUsage(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)
			adapter.RegisterRoute(route.Meta{
				Method: "GET",
				Path:   "/test",
				Func:   dummyHandler,
			})

			req := httptest.NewRequest("GET", "/test", nil)

			b.ReportAllocs()
			b.ResetTimer()

			var m1, m2 sync.Map
			b.ReportMetric(0, "B/request")

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				adapter.ServeHTTP(w, req)
			}

			// Store metrics
			_ = m1
			_ = m2
		})
	}
}

// BenchmarkThroughput benchmarks requests per second capability
func BenchmarkThroughput(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	routerTypes := []routeradapter.RouterType{
		routeradapter.RouterTypeGin,
		routeradapter.RouterTypeFiber,
		routeradapter.RouterTypeEcho,
		routeradapter.RouterTypeNetHTTP,
	}

	for _, routerType := range routerTypes {
		b.Run(string(routerType), func(b *testing.B) {
			cfg := routeradapter.DefaultRouterConfig()
			cfg.Type = routerType

			adapter, _ := routeradapter.NewRouterAdapter(cfg)

			// Set up realistic API
			adapter.RegisterRoute(route.Meta{
				Method: "GET",
				Path:   "/api/users/:id",
				Func:   dummyHandler,
			})

			server := httptest.NewServer(http.HandlerFunc(adapter.ServeHTTP))
			defer server.Close()

			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					resp, err := client.Get(server.URL + "/api/users/123")
					if err != nil {
						b.Fatalf("Request failed: %v", err)
					}
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
			})
		})
	}
}
