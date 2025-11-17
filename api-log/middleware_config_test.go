package log

import "testing"

func TestDefaultMiddlewareConfig(t *testing.T) {
	config := DefaultMiddlewareConfig()

	if config == nil {
		t.Fatal("DefaultMiddlewareConfig returned nil")
	}

	// Verify default skip paths
	expectedPaths := []string{"/healthz", "/health"}
	if len(config.SkipPaths) != len(expectedPaths) {
		t.Errorf("Expected %d skip paths, got %d", len(expectedPaths), len(config.SkipPaths))
	}

	for i, path := range expectedPaths {
		if config.SkipPaths[i] != path {
			t.Errorf("Expected skip path %s, got %s", path, config.SkipPaths[i])
		}
	}

	// Verify empty prefixes
	if len(config.SkipPathPrefixes) != 0 {
		t.Errorf("Expected no skip path prefixes, got %d", len(config.SkipPathPrefixes))
	}

	// Verify empty method paths
	if config.SkipMethodPaths == nil {
		t.Error("SkipMethodPaths should be initialized, not nil")
	}
	if len(config.SkipMethodPaths) != 0 {
		t.Errorf("Expected no skip method paths, got %d", len(config.SkipMethodPaths))
	}
}

func TestMiddlewareConfig_ShouldSkip_NilConfig(t *testing.T) {
	var config *MiddlewareConfig = nil

	// Nil config should never skip
	if config.ShouldSkip("GET", "/healthz") {
		t.Error("Nil config should not skip any paths")
	}
}

func TestMiddlewareConfig_ShouldSkip_ExactPaths(t *testing.T) {
	config := &MiddlewareConfig{
		SkipPaths:        []string{"/healthz", "/metrics", "/ready"},
		SkipPathPrefixes: []string{},
		SkipMethodPaths:  make(map[string][]string),
	}

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{"skip /healthz", "GET", "/healthz", true},
		{"skip /metrics", "POST", "/metrics", true},
		{"skip /ready", "GET", "/ready", true},
		{"don't skip /health", "GET", "/health", false},
		{"don't skip /api/users", "GET", "/api/users", false},
		{"don't skip /healthz/status", "GET", "/healthz/status", false}, // Not exact match
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ShouldSkip(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("ShouldSkip(%s, %s) = %v, expected %v", tt.method, tt.path, result, tt.expected)
			}
		})
	}
}

func TestMiddlewareConfig_ShouldSkip_PathPrefixes(t *testing.T) {
	config := &MiddlewareConfig{
		SkipPaths:        []string{},
		SkipPathPrefixes: []string{"/internal/", "/debug/", "/_"},
		SkipMethodPaths:  make(map[string][]string),
	}

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{"skip /internal/health", "GET", "/internal/health", true},
		{"skip /internal/metrics", "POST", "/internal/metrics", true},
		{"skip /debug/pprof", "GET", "/debug/pprof", true},
		{"skip /_status", "GET", "/_status", true},
		{"don't skip /api/internal", "GET", "/api/internal", false}, // Prefix doesn't match
		{"don't skip /api/users", "GET", "/api/users", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ShouldSkip(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("ShouldSkip(%s, %s) = %v, expected %v", tt.method, tt.path, result, tt.expected)
			}
		})
	}
}

func TestMiddlewareConfig_ShouldSkip_MethodPaths(t *testing.T) {
	config := &MiddlewareConfig{
		SkipPaths:        []string{},
		SkipPathPrefixes: []string{},
		SkipMethodPaths: map[string][]string{
			"GET":  {"/status", "/ping"},
			"POST": {"/webhook"},
		},
	}

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{"skip GET /status", "GET", "/status", true},
		{"skip GET /ping", "GET", "/ping", true},
		{"skip POST /webhook", "POST", "/webhook", true},
		{"don't skip POST /status", "POST", "/status", false}, // Different method
		{"don't skip GET /webhook", "GET", "/webhook", false}, // Different method
		{"don't skip PUT /status", "PUT", "/status", false},   // Method not configured
		{"don't skip GET /api/users", "GET", "/api/users", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ShouldSkip(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("ShouldSkip(%s, %s) = %v, expected %v", tt.method, tt.path, result, tt.expected)
			}
		})
	}
}

func TestMiddlewareConfig_ShouldSkip_Combined(t *testing.T) {
	config := &MiddlewareConfig{
		SkipPaths:        []string{"/healthz"},
		SkipPathPrefixes: []string{"/internal/"},
		SkipMethodPaths: map[string][]string{
			"GET": {"/metrics"},
		},
	}

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{"skip via exact path", "POST", "/healthz", true},
		{"skip via prefix", "GET", "/internal/debug", true},
		{"skip via method+path", "GET", "/metrics", true},
		{"don't skip regular path", "GET", "/api/users", false},
		{"don't skip POST /metrics", "POST", "/metrics", false}, // Different method
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ShouldSkip(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("ShouldSkip(%s, %s) = %v, expected %v", tt.method, tt.path, result, tt.expected)
			}
		})
	}
}

func TestMiddlewareConfig_ShouldSkip_EmptyConfig(t *testing.T) {
	config := &MiddlewareConfig{
		SkipPaths:        []string{},
		SkipPathPrefixes: []string{},
		SkipMethodPaths:  make(map[string][]string),
	}

	// Empty config should never skip
	if config.ShouldSkip("GET", "/healthz") {
		t.Error("Empty config should not skip any paths")
	}
	if config.ShouldSkip("POST", "/api/users") {
		t.Error("Empty config should not skip any paths")
	}
}

func TestMiddlewareConfig_ShouldSkip_CaseSensitive(t *testing.T) {
	config := &MiddlewareConfig{
		SkipPaths:        []string{"/healthz"},
		SkipPathPrefixes: []string{},
		SkipMethodPaths: map[string][]string{
			"GET": {"/metrics"},
		},
	}

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{"exact case match", "GET", "/healthz", true},
		{"path case mismatch", "GET", "/HEALTHZ", false}, // Paths are case-sensitive
		{"path case mismatch 2", "GET", "/Healthz", false},
		{"method case match", "GET", "/metrics", true},
		{"method case mismatch", "get", "/metrics", false}, // Methods are case-sensitive in map
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ShouldSkip(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("ShouldSkip(%s, %s) = %v, expected %v", tt.method, tt.path, result, tt.expected)
			}
		})
	}
}
