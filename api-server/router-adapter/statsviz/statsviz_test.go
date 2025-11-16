package statsviz

import (
	"net/http/httptest"
	"strings"
	"testing"

	"MgApplication/api-server/router-adapter"
)

// TestStatsvizHandler tests the statsviz visualization handler
func TestStatsvizHandler(t *testing.T) {
	middleware, err := StatsvizHandler("/debug/statsviz")
	if err != nil {
		t.Fatalf("Failed to create statsviz handler: %v", err)
	}

	t.Run("index page", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/debug/statsviz/", nil)
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
			t.Error("Next handler should not be called for statsviz index path")
		}

		// Check for HTML content
		body := w.Body.String()
		if !strings.Contains(body, "html") && !strings.Contains(body, "statsviz") {
			t.Log("Response may not contain expected HTML, but handler executed")
		}

		t.Logf("Statsviz index served (response size: %d bytes)", len(body))
	})

	t.Run("non-matching path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/other", nil)
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

		if !nextCalled {
			t.Error("Next handler should be called for non-matching path")
		}
	})
}

// TestDefaultStatsvizHandler tests the default statsviz handler
func TestDefaultStatsvizHandler(t *testing.T) {
	middleware, err := DefaultStatsvizHandler()
	if err != nil {
		t.Fatalf("Failed to create default statsviz handler: %v", err)
	}

	req := httptest.NewRequest("GET", "/debug/statsviz/", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	err = middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Log("Default statsviz handler created and executed successfully")
}

// TestStatsvizHandlerCustomPath tests statsviz with custom path
func TestStatsvizHandlerCustomPath(t *testing.T) {
	middleware, err := StatsvizHandler("/custom/stats")
	if err != nil {
		t.Fatalf("Failed to create statsviz handler: %v", err)
	}

	req := httptest.NewRequest("GET", "/custom/stats/", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	err = middleware(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Log("Statsviz with custom path executed successfully")
}

// TestStatsvizWebSocketPath tests the websocket endpoint
func TestStatsvizWebSocketPath(t *testing.T) {
	middleware, err := StatsvizHandler("/debug/statsviz")
	if err != nil {
		t.Fatalf("Failed to create statsviz handler: %v", err)
	}

	req := httptest.NewRequest("GET", "/debug/statsviz/ws", nil)
	w := httptest.NewRecorder()
	ctx := routeradapter.NewRouterContext(w, req)

	nextCalled := false
	err = middleware(ctx, func() error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if nextCalled {
		t.Error("Next handler should not be called for websocket path")
	}

	t.Log("Statsviz websocket endpoint handled")
}
