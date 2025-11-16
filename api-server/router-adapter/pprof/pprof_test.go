package pprof

import (
	"net/http/httptest"
	"strings"
	"testing"

	"MgApplication/api-server/router-adapter"
)

// TestPprofIndexHandler tests the pprof index handler
func TestPprofIndexHandler(t *testing.T) {
	middleware := PprofIndexHandler("/debug/pprof/")

	req := httptest.NewRequest("GET", "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	nextCalled := false
	err := middleware(ctx, func() error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if nextCalled {
		t.Error("Next handler should not be called for pprof index path")
	}

	// Response should contain pprof content
	body := w.Body.String()
	if !strings.Contains(body, "profile") || !strings.Contains(body, "pprof") {
		t.Error("Expected pprof index content in response")
	}

	t.Logf("Pprof index served successfully")
}

// TestPprofHandlersNonMatchingPath tests that handlers pass through non-matching paths
func TestPprofHandlersNonMatchingPath(t *testing.T) {
	handlers := map[string]routeradapter.MiddlewareFunc{
		"index":        PprofIndexHandler("/debug/pprof/"),
		"cmdline":      PprofCmdlineHandler("/debug/pprof/cmdline"),
		"profile":      PprofProfileHandler("/debug/pprof/profile"),
		"symbol":       PprofSymbolHandler("/debug/pprof/symbol"),
		"trace":        PprofTraceHandler("/debug/pprof/trace"),
		"allocs":       PprofAllocsHandler("/debug/pprof/allocs"),
		"block":        PprofBlockHandler("/debug/pprof/block"),
		"goroutine":    PprofGoroutineHandler("/debug/pprof/goroutine"),
		"heap":         PprofHeapHandler("/debug/pprof/heap"),
		"mutex":        PprofMutexHandler("/debug/pprof/mutex"),
		"threadcreate": PprofThreadCreateHandler("/debug/pprof/threadcreate"),
	}

	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/other", nil)
			w := httptest.NewRecorder()
			ctx := routeradapter.NewRouterContext(w, req)

			nextCalled := false
			err := handler(ctx, func() error {
				nextCalled = true
				return nil
			})

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !nextCalled {
				t.Errorf("%s handler should call next for non-matching path", name)
			}
		})
	}
}

// TestPprofCmdlineHandler tests the cmdline handler
func TestPprofCmdlineHandler(t *testing.T) {
	middleware := PprofCmdlineHandler("/debug/pprof/cmdline")

	req := httptest.NewRequest("GET", "/debug/pprof/cmdline", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	err := middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have some content (command line arguments)
	if w.Body.Len() == 0 {
		t.Error("Expected cmdline content in response")
	}

	t.Logf("Cmdline handler served successfully")
}

// TestPprofHeapHandler tests the heap profile handler
func TestPprofHeapHandler(t *testing.T) {
	middleware := PprofHeapHandler("/debug/pprof/heap")

	req := httptest.NewRequest("GET", "/debug/pprof/heap", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	err := middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Heap profile should return some content
	if w.Body.Len() == 0 {
		t.Error("Expected heap profile content in response")
	}

	t.Logf("Heap profile served successfully (%d bytes)", w.Body.Len())
}

// TestPprofGoroutineHandler tests the goroutine handler
func TestPprofGoroutineHandler(t *testing.T) {
	middleware := PprofGoroutineHandler("/debug/pprof/goroutine")

	req := httptest.NewRequest("GET", "/debug/pprof/goroutine", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	err := middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have goroutine profile content
	if w.Body.Len() == 0 {
		t.Error("Expected goroutine profile content")
	}

	t.Logf("Goroutine profile served successfully")
}
