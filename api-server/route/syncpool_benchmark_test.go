package route

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// SYNC.POOL PERFORMANCE COMPARISON BENCHMARKS
// ============================================================================
// This file compares the performance of:
// 1. buildImproved (WITH sync.Pool) - Current optimized implementation
// 2. build (WITHOUT sync.Pool) - Original implementation
//
// The goal is to measure the actual performance improvement from sync.Pool
// ============================================================================

// Test request/response types
type SyncPoolBenchRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Age     int    `json:"age" binding:"required,min=1,max=150"`
	Country string `json:"country" binding:"required"`
}

type SyncPoolBenchResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

func (r SyncPoolBenchResponse) Status() int                   { return 200 }
func (r SyncPoolBenchResponse) GetContentType() string        { return "application/json" }
func (r SyncPoolBenchResponse) GetContentDisposition() string { return "" }
func (r SyncPoolBenchResponse) ResponseType() string          { return "json" }
func (r SyncPoolBenchResponse) Object() []byte                { return nil }

func syncPoolBenchHandler(ctx *Context, req SyncPoolBenchRequest) (SyncPoolBenchResponse, error) {
	return SyncPoolBenchResponse{
		Success: true,
		Message: "User " + req.Name + " processed successfully",
		UserID:  "user-12345",
	}, nil
}

// ============================================================================
// BENCHMARK 1: JSON Request Processing
// ============================================================================

func BenchmarkWithSyncPool_JSON(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildImproved(syncPoolBenchHandler)
	router.POST("/api/users", handler)

	body := `{"name":"John Doe","email":"john@example.com","age":30,"country":"USA"}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutSyncPool_JSON(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := build(syncPoolBenchHandler)
	router.POST("/api/users", handler)

	body := `{"name":"John Doe","email":"john@example.com","age":30,"country":"USA"}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ============================================================================
// BENCHMARK 2: Form Data Processing
// ============================================================================

func BenchmarkWithSyncPool_FormData(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildImproved(syncPoolBenchHandler)
	router.POST("/api/users", handler)

	body := "name=John+Doe&email=john@example.com&age=30&country=USA"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutSyncPool_FormData(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := build(syncPoolBenchHandler)
	router.POST("/api/users", handler)

	body := "name=John+Doe&email=john@example.com&age=30&country=USA"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ============================================================================
// BENCHMARK 3: Large Payload Processing
// ============================================================================

func BenchmarkWithSyncPool_LargePayload(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildImproved(syncPoolBenchHandler)
	router.POST("/api/users", handler)

	// Create a larger JSON payload
	var buf bytes.Buffer
	buf.WriteString(`{"name":"`)
	buf.WriteString(strings.Repeat("A", 500)) // 500 char name
	buf.WriteString(`","email":"john@example.com","age":30,"country":"USA"}`)
	body := buf.String()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutSyncPool_LargePayload(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := build(syncPoolBenchHandler)
	router.POST("/api/users", handler)

	// Create a larger JSON payload
	var buf bytes.Buffer
	buf.WriteString(`{"name":"`)
	buf.WriteString(strings.Repeat("A", 500)) // 500 char name
	buf.WriteString(`","email":"john@example.com","age":30,"country":"USA"}`)
	body := buf.String()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ============================================================================
// BENCHMARK 4: Parallel Processing (High Concurrency)
// ============================================================================

func BenchmarkWithSyncPool_Parallel(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildImproved(syncPoolBenchHandler)
	router.POST("/api/users", handler)

	body := `{"name":"John Doe","email":"john@example.com","age":30,"country":"USA"}`

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkWithoutSyncPool_Parallel(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := build(syncPoolBenchHandler)
	router.POST("/api/users", handler)

	body := `{"name":"John Doe","email":"john@example.com","age":30,"country":"USA"}`

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// ============================================================================
// BENCHMARK 5: GET Requests (No Body)
// ============================================================================

type GetBenchResponse struct {
	StatusText string `json:"status"`
	Data       string `json:"data"`
}

func (r GetBenchResponse) Status() int                   { return 200 }
func (r GetBenchResponse) GetContentType() string        { return "application/json" }
func (r GetBenchResponse) GetContentDisposition() string { return "" }
func (r GetBenchResponse) ResponseType() string          { return "json" }
func (r GetBenchResponse) Object() []byte                { return nil }

func getBenchHandler(ctx *Context, req struct{}) (GetBenchResponse, error) {
	return GetBenchResponse{
		StatusText: "ok",
		Data:       "response data",
	}, nil
}

func BenchmarkWithSyncPool_GET(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := buildImproved(getBenchHandler)
	router.GET("/api/status", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkWithoutSyncPool_GET(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handler := build(getBenchHandler)
	router.GET("/api/status", handler)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ============================================================================
// BENCHMARK 6: Memory Allocation Test
// ============================================================================

func BenchmarkWithSyncPool_AllocationsOnly(b *testing.B) {
	b.ReportAllocs()

	// Create a proper gin context with request
	ginCtx := &gin.Context{}
	ginCtx.Request = httptest.NewRequest("GET", "/", nil)

	for i := 0; i < b.N; i++ {
		ctx := getContext()
		ctx.fromGinCtx(ginCtx)
		if ctx.cancel != nil {
			ctx.cancel()
		}
		putContext(ctx)
	}
}

func BenchmarkWithoutSyncPool_AllocationsOnly(b *testing.B) {
	b.ReportAllocs()

	// Create a proper gin context with request
	ginCtx := &gin.Context{}
	ginCtx.Request = httptest.NewRequest("GET", "/", nil)

	for i := 0; i < b.N; i++ {
		ctx := &Context{}
		ctx.fromGinCtx(ginCtx)
		if ctx.cancel != nil {
			ctx.cancel()
		}
		// No pool to return to - just let GC handle it
	}
}
