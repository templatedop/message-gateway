package router

import (
	"reflect"
	"testing"

	"MgApplication/api-server/handler"
	"MgApplication/api-server/route"
	"MgApplication/api-server/swagger"

	"github.com/gin-gonic/gin"
)

// Mock handler for testing
type mockHandler struct {
	prefix string
	name   string
	routes []route.Route
	mws    []gin.HandlerFunc
}

func (m *mockHandler) Prefix() string {
	return m.prefix
}

func (m *mockHandler) Name() string {
	return m.name
}

func (m *mockHandler) Routes() []route.Route {
	return m.routes
}

func (m *mockHandler) Middlewares() []gin.HandlerFunc {
	return m.mws
}

// Mock route for testing
type mockRoute struct {
	meta route.Meta
}

func (m *mockRoute) Meta() route.Meta {
	return m.meta
}

func (m *mockRoute) Desc(s string) route.Route {
	m.meta.Description = s
	return m
}

func (m *mockRoute) Name(s string) route.Route {
	m.meta.Name = s
	return m
}

func (m *mockRoute) AddMiddlewares(mws ...gin.HandlerFunc) route.Route {
	m.meta.Middlewares = append(m.meta.Middlewares, mws...)
	return m
}

func TestNewRegistry(t *testing.T) {
	h := &mockHandler{
		prefix: "/api/v1",
		name:   "test",
		routes: []route.Route{},
		mws:    []gin.HandlerFunc{},
	}

	reg := newRegistry(h)

	if reg.base != "/api/v1" {
		t.Errorf("base = %v, want %v", reg.base, "/api/v1")
	}

	if reg.name != "test" {
		t.Errorf("name = %v, want %v", reg.name, "test")
	}
}

func TestRegistry_GetMetas_Caching(t *testing.T) {
	// Create mock routes
	routes := []route.Route{
		&mockRoute{meta: route.Meta{
			Method: "GET",
			Path:   "/users",
			Func:   func(*gin.Context) {},
		}},
		&mockRoute{meta: route.Meta{
			Method: "POST",
			Path:   "/users",
			Func:   func(*gin.Context) {},
		}},
	}

	h := &mockHandler{
		prefix: "/api",
		name:   "users",
		routes: routes,
	}

	reg := newRegistry(h)

	// First call should compute metas
	metas1 := reg.getMetas()
	if len(metas1) != 2 {
		t.Errorf("Expected 2 metas, got %d", len(metas1))
	}

	// Verify paths are prefixed
	for _, meta := range metas1 {
		if meta.Path == "/users" {
			t.Error("Path should be prefixed with base path")
		}
	}

	// Second call should return cached metas (same slice)
	metas2 := reg.getMetas()
	if len(metas2) != 2 {
		t.Errorf("Expected 2 cached metas, got %d", len(metas2))
	}

	// Verify it's the same cached slice (pointer equality)
	if &metas1[0] != &metas2[0] {
		t.Error("Second call should return cached metas (same slice)")
	}
}

func TestRegistry_GetMetas_PathPrefixing(t *testing.T) {
	tests := []struct {
		name       string
		base       string
		routePath  string
		expectPath string
	}{
		{
			name:       "Simple prefix",
			base:       "/api",
			routePath:  "/users",
			expectPath: "/api/users",
		},
		{
			name:       "Prefix with trailing slash",
			base:       "/api/",
			routePath:  "/users",
			expectPath: "/api/users",
		},
		{
			name:       "Empty base",
			base:       "",
			routePath:  "/users",
			expectPath: "/users",
		},
		{
			name:       "Nested path",
			base:       "/api/v1",
			routePath:  "/users/:id",
			expectPath: "/api/v1/users/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routes := []route.Route{
				&mockRoute{meta: route.Meta{
					Method: "GET",
					Path:   tt.routePath,
					Func:   func(*gin.Context) {},
				}},
			}

			h := &mockHandler{
				prefix: tt.base,
				name:   "test",
				routes: routes,
			}

			reg := newRegistry(h)
			metas := reg.getMetas()

			if len(metas) != 1 {
				t.Fatalf("Expected 1 meta, got %d", len(metas))
			}

			if metas[0].Path != tt.expectPath {
				t.Errorf("Path = %v, want %v", metas[0].Path, tt.expectPath)
			}
		})
	}
}

func TestRegistry_SwaggerDefs_UsesCachedMetas(t *testing.T) {
	// Create mock routes with request/response types
	type TestReq struct {
		Name string `json:"name"`
	}
	type TestResp struct {
		ID int `json:"id"`
	}

	routes := []route.Route{
		&mockRoute{meta: route.Meta{
			Method: "POST",
			Path:   "/users",
			Func:   func(*gin.Context) {},
			Req:    reflect.TypeOf(TestReq{}),
			Res:    reflect.TypeOf(TestResp{}),
		}},
	}

	h := &mockHandler{
		prefix: "/api",
		name:   "users",
		routes: routes,
	}

	reg := newRegistry(h)

	// Call getMetas first to populate cache
	metas1 := reg.getMetas()

	// Call SwaggerDefs - should use cached metas
	defs := reg.SwaggerDefs()

	if len(defs) != 1 {
		t.Errorf("Expected 1 swagger def, got %d", len(defs))
	}

	// Verify it used cached metas by checking the path is prefixed
	if defs[0].Endpoint != metas1[0].Path {
		t.Error("SwaggerDefs should use cached metas with prefixed paths")
	}

	// Verify swagger def fields
	if defs[0].Group != "users" {
		t.Errorf("Group = %v, want %v", defs[0].Group, "users")
	}

	if defs[0].Method != "POST" {
		t.Errorf("Method = %v, want %v", defs[0].Method, "POST")
	}

	if defs[0].RequestType != reflect.TypeOf(TestReq{}) {
		t.Error("RequestType should match")
	}

	if defs[0].ResponseType != reflect.TypeOf(TestResp{}) {
		t.Error("ResponseType should match")
	}
}

func TestGetSwaggerDefs(t *testing.T) {
	type TestReq struct{ Name string }
	type TestResp struct{ ID int }

	controllers := []handler.Handler{
		&mockHandler{
			prefix: "/api/users",
			name:   "users",
			routes: []route.Route{
				&mockRoute{meta: route.Meta{
					Method: "GET",
					Path:   "/",
					Req:    reflect.TypeOf(TestReq{}),
					Res:    reflect.TypeOf(TestResp{}),
				}},
			},
		},
		&mockHandler{
			prefix: "/api/products",
			name:   "products",
			routes: []route.Route{
				&mockRoute{meta: route.Meta{
					Method: "GET",
					Path:   "/",
					Req:    reflect.TypeOf(TestReq{}),
					Res:    reflect.TypeOf(TestResp{}),
				}},
			},
		},
	}

	registries := ParseControllers(controllers...)
	defs := GetSwaggerDefs(registries)

	// Should have 2 endpoint definitions
	if len(defs) != 2 {
		t.Errorf("Expected 2 swagger defs, got %d", len(defs))
	}

	// Verify endpoints have different groups
	groups := make(map[string]bool)
	for _, def := range defs {
		groups[def.Group] = true
	}

	if !groups["users"] || !groups["products"] {
		t.Error("Should have defs for both users and products groups")
	}
}

func TestParseControllers(t *testing.T) {
	controllers := []handler.Handler{
		&mockHandler{prefix: "/api/v1", name: "test1"},
		&mockHandler{prefix: "/api/v2", name: "test2"},
	}

	registries := ParseControllers(controllers...)

	if len(registries) != 2 {
		t.Errorf("Expected 2 registries, got %d", len(registries))
	}

	// Verify each controller was converted to registry
	names := make(map[string]bool)
	for _, reg := range registries {
		names[reg.name] = true
	}

	if !names["test1"] || !names["test2"] {
		t.Error("All controllers should be converted to registries")
	}
}

func TestParseGroupedControllers_Empty(t *testing.T) {
	p := struct {
		Controllers []handler.Handler `group:"servercontrollers"`
	}{
		Controllers: []handler.Handler{},
	}

	registries := ParseGroupedControllers(p)

	if registries != nil {
		t.Error("Empty controllers should return nil (not empty slice)")
	}
}

func TestParseGroupedControllers_WithControllers(t *testing.T) {
	p := struct {
		Controllers []handler.Handler `group:"servercontrollers"`
	}{
		Controllers: []handler.Handler{
			&mockHandler{prefix: "/api", name: "test"},
		},
	}

	registries := ParseGroupedControllers(p)

	if len(registries) != 1 {
		t.Errorf("Expected 1 registry, got %d", len(registries))
	}
}

// Benchmark tests

func BenchmarkRegistry_GetMetas_FirstCall(b *testing.B) {
	routes := []route.Route{
		&mockRoute{meta: route.Meta{Method: "GET", Path: "/users"}},
		&mockRoute{meta: route.Meta{Method: "POST", Path: "/users"}},
		&mockRoute{meta: route.Meta{Method: "GET", Path: "/users/:id"}},
		&mockRoute{meta: route.Meta{Method: "PUT", Path: "/users/:id"}},
		&mockRoute{meta: route.Meta{Method: "DELETE", Path: "/users/:id"}},
	}

	h := &mockHandler{
		prefix: "/api",
		name:   "users",
		routes: routes,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg := newRegistry(h)
		reg.getMetas()
	}
}

func BenchmarkRegistry_GetMetas_CachedCall(b *testing.B) {
	routes := []route.Route{
		&mockRoute{meta: route.Meta{Method: "GET", Path: "/users"}},
		&mockRoute{meta: route.Meta{Method: "POST", Path: "/users"}},
		&mockRoute{meta: route.Meta{Method: "GET", Path: "/users/:id"}},
		&mockRoute{meta: route.Meta{Method: "PUT", Path: "/users/:id"}},
		&mockRoute{meta: route.Meta{Method: "DELETE", Path: "/users/:id"}},
	}

	h := &mockHandler{
		prefix: "/api",
		name:   "users",
		routes: routes,
	}

	reg := newRegistry(h)
	reg.getMetas() // Prime the cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg.getMetas()
	}
}

func BenchmarkGetSwaggerDefs(b *testing.B) {
	type TestReq struct{ Name string }
	type TestResp struct{ ID int }

	controllers := []handler.Handler{
		&mockHandler{
			prefix: "/api/users",
			name:   "users",
			routes: []route.Route{
				&mockRoute{meta: route.Meta{Method: "GET", Path: "/", Req: reflect.TypeOf(TestReq{}), Res: reflect.TypeOf(TestResp{})}},
				&mockRoute{meta: route.Meta{Method: "POST", Path: "/", Req: reflect.TypeOf(TestReq{}), Res: reflect.TypeOf(TestResp{})}},
			},
		},
		&mockHandler{
			prefix: "/api/products",
			name:   "products",
			routes: []route.Route{
				&mockRoute{meta: route.Meta{Method: "GET", Path: "/", Req: reflect.TypeOf(TestReq{}), Res: reflect.TypeOf(TestResp{})}},
			},
		},
	}

	registries := ParseControllers(controllers...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetSwaggerDefs(registries)
	}
}
