package route

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"MgApplication/api-server/util/diutil/typlect"
)

// ============================================================================
// BENCHMARK SETUP
// ============================================================================

type BenchmarkRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type BenchmarkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (r BenchmarkResponse) Status() int                   { return 200 }
func (r BenchmarkResponse) GetContentType() string        { return "application/json" }
func (r BenchmarkResponse) GetContentDisposition() string { return "" }
func (r BenchmarkResponse) ResponseType() string          { return "json" }
func (r BenchmarkResponse) Object() []byte                { return nil }

func benchmarkHandler(ctx *Context, req BenchmarkRequest) (BenchmarkResponse, error) {
	return BenchmarkResponse{
		Success: true,
		Message: "Processed: " + req.Name,
	}, nil
}

// ============================================================================
// VERSION 1: DEFERRED CLEANUP (CURRENT IMPLEMENTATION)
// ============================================================================

func buildWithDeferredCleanup[Req, Res any](f HandlerFunc[Req, Res], defaultStatus ...int) gin.HandlerFunc {
	ds := http.StatusOK
	if len(defaultStatus) == 1 {
		ds = defaultStatus[0]
	}

	hasInput := typlect.GetType[Req]() != typlect.TypeNoParam

	return func(c *gin.Context) {
		// Get pooled context
		ctx := getContext()
		defer func() {
			if ctx.cancel != nil {
				ctx.cancel()
			}
			putContext(ctx)
		}()

		ctx.fromGinCtx(c)

		var req Req
		if hasInput {
			if c.Request.Method != http.MethodGet && c.Request.Body != nil {
				if err := c.ShouldBindJSON(&req); err != nil {
					return
				}
			}
		}

		res, err := f(ctx, req)
		if err != nil {
			c.Error(err)
			return
		}

		handleResponse(c, res, ds)
	}
}

// ============================================================================
// VERSION 2: EXPLICIT CLEANUP (NO DEFER)
// ============================================================================

func buildWithExplicitCleanup[Req, Res any](f HandlerFunc[Req, Res], defaultStatus ...int) gin.HandlerFunc {
	ds := http.StatusOK
	if len(defaultStatus) == 1 {
		ds = defaultStatus[0]
	}

	hasInput := typlect.GetType[Req]() != typlect.TypeNoParam

	return func(c *gin.Context) {
		// Get pooled context
		ctx := getContext()
		ctx.fromGinCtx(c)

		var req Req
		if hasInput {
			if c.Request.Method != http.MethodGet && c.Request.Body != nil {
				if err := c.ShouldBindJSON(&req); err != nil {
					// Explicit cleanup on error path
					if ctx.cancel != nil {
						ctx.cancel()
					}
					putContext(ctx)
					return
				}
			}
		}

		res, err := f(ctx, req)
		if err != nil {
			// Explicit cleanup on error path
			if ctx.cancel != nil {
				ctx.cancel()
			}
			putContext(ctx)
			c.Error(err)
			return
		}

		handleResponse(c, res, ds)

		// Explicit cleanup on success path
		if ctx.cancel != nil {
			ctx.cancel()
		}
		putContext(ctx)
	}
}

// ============================================================================
// VERSION 3: DEFERRED WITH NAMED RETURN (OPTIMIZATION)
// ============================================================================

func buildWithNamedDeferCleanup[Req, Res any](f HandlerFunc[Req, Res], defaultStatus ...int) gin.HandlerFunc {
	ds := http.StatusOK
	if len(defaultStatus) == 1 {
		ds = defaultStatus[0]
	}

	hasInput := typlect.GetType[Req]() != typlect.TypeNoParam

	return func(c *gin.Context) {
		ctx := getContext()
		defer putContextWithCancel(ctx)

		ctx.fromGinCtx(c)

		var req Req
		if hasInput {
			if c.Request.Method != http.MethodGet && c.Request.Body != nil {
				if err := c.ShouldBindJSON(&req); err != nil {
					return
				}
			}
		}

		res, err := f(ctx, req)
		if err != nil {
			c.Error(err)
			return
		}

		handleResponse(c, res, ds)
	}
}

// Helper function to reduce defer overhead
func putContextWithCancel(ctx *Context) {
	if ctx.cancel != nil {
		ctx.cancel()
	}
	putContext(ctx)
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkDeferredCleanup(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildWithDeferredCleanup(benchmarkHandler)
	router.POST("/test", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test","email":"test@example.com"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkExplicitCleanup(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildWithExplicitCleanup(benchmarkHandler)
	router.POST("/test", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test","email":"test@example.com"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkNamedDeferCleanup(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildWithNamedDeferCleanup(benchmarkHandler)
	router.POST("/test", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test","email":"test@example.com"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ============================================================================
// MICRO-BENCHMARKS: ISOLATED POOL OPERATIONS
// ============================================================================

func BenchmarkPoolOperations_Deferred(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() {
			ctx := getContext()
			defer func() {
				if ctx.cancel != nil {
					ctx.cancel()
				}
				putContext(ctx)
			}()

			// Simulate some work
			ctx.Ctx = context.Background()
			var cancel context.CancelFunc
			ctx.Ctx, cancel = context.WithCancel(ctx.Ctx)
			ctx.cancel = cancel
		}()
	}
}

func BenchmarkPoolOperations_Explicit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() {
			ctx := getContext()

			// Simulate some work
			ctx.Ctx = context.Background()
			var cancel context.CancelFunc
			ctx.Ctx, cancel = context.WithCancel(ctx.Ctx)
			ctx.cancel = cancel

			// Explicit cleanup
			if ctx.cancel != nil {
				ctx.cancel()
			}
			putContext(ctx)
		}()
	}
}

func BenchmarkPoolOperations_NamedDefer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() {
			ctx := getContext()
			defer putContextWithCancel(ctx)

			// Simulate some work
			ctx.Ctx = context.Background()
			var cancel context.CancelFunc
			ctx.Ctx, cancel = context.WithCancel(ctx.Ctx)
			ctx.cancel = cancel
		}()
	}
}

// ============================================================================
// CONCURRENT BENCHMARKS
// ============================================================================

func BenchmarkDeferredCleanup_Parallel(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildWithDeferredCleanup(benchmarkHandler)
	router.POST("/test", handler)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test","email":"test@example.com"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkExplicitCleanup_Parallel(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildWithExplicitCleanup(benchmarkHandler)
	router.POST("/test", handler)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test","email":"test@example.com"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkNamedDeferCleanup_Parallel(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildWithNamedDeferCleanup(benchmarkHandler)
	router.POST("/test", handler)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test","email":"test@example.com"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}
