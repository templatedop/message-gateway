package swagger

import (
	"reflect"
	"testing"

	errors "MgApplication/api-errors"
	"MgApplication/api-server/util/diutil/typlect"
)

// Test structures
type TestUser struct {
	ID       int    `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email"`
	Age      int    `json:"age,omitempty"`
	IsActive bool   `json:"is_active"`
}

type TestProduct struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type TestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type NestedStruct struct {
	User    TestUser    `json:"user"`
	Product TestProduct `json:"product"`
}

func TestBuildDefinitions(t *testing.T) {
	tests := []struct {
		name          string
		endpoints     []EndpointDef
		expectTypes   int // minimum expected type definitions
		expectCaching bool
	}{
		{
			name: "Single endpoint",
			endpoints: []EndpointDef{
				{
					RequestType:  reflect.TypeOf(TestUser{}),
					ResponseType: reflect.TypeOf(TestResponse{}),
					Group:        "users",
					Name:         "CreateUser",
					Endpoint:     "/api/users",
					Method:       "POST",
				},
			},
			expectTypes:   3, // TestUser, TestResponse, APIErrorResponse
			expectCaching: true,
		},
		{
			name: "Multiple endpoints with duplicate types",
			endpoints: []EndpointDef{
				{
					RequestType:  reflect.TypeOf(TestUser{}),
					ResponseType: reflect.TypeOf(TestResponse{}),
					Group:        "users",
					Name:         "CreateUser",
					Endpoint:     "/api/users",
					Method:       "POST",
				},
				{
					RequestType:  reflect.TypeOf(TestUser{}),
					ResponseType: reflect.TypeOf(TestResponse{}),
					Group:        "users",
					Name:         "UpdateUser",
					Endpoint:     "/api/users/:id",
					Method:       "PUT",
				},
			},
			expectTypes:   3, // Should not duplicate: TestUser, TestResponse, APIErrorResponse
			expectCaching: true,
		},
		{
			name: "Nested structures",
			endpoints: []EndpointDef{
				{
					RequestType:  reflect.TypeOf(NestedStruct{}),
					ResponseType: reflect.TypeOf(TestResponse{}),
					Group:        "complex",
					Name:         "ComplexOperation",
					Endpoint:     "/api/complex",
					Method:       "POST",
				},
			},
			expectTypes:   5, // NestedStruct, TestUser, TestProduct, TestResponse, APIErrorResponse
			expectCaching: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset type cache before each test
			processedTypes.Range(func(key, value interface{}) bool {
				processedTypes.Delete(key)
				return true
			})

			defs := buildDefinitions(tt.endpoints)

			// Check minimum number of definitions
			if len(defs) < tt.expectTypes {
				t.Errorf("Expected at least %d type definitions, got %d", tt.expectTypes, len(defs))
			}

			// Verify APIErrorResponse is always included
			errType := reflect.TypeOf(errors.APIErrorResponse{})
			errName := getNameFromType(errType)
			if _, ok := defs[errName]; !ok {
				t.Errorf("APIErrorResponse should always be included in definitions")
			}

			// Verify type cache is populated
			if tt.expectCaching {
				cacheSize := 0
				processedTypes.Range(func(key, value interface{}) bool {
					cacheSize++
					return true
				})
				if cacheSize == 0 {
					t.Error("Type cache should be populated after buildDefinitions")
				}
			}
		})
	}
}

func TestBuildModelDefinition_TypeCaching(t *testing.T) {
	// Reset cache
	processedTypes.Range(func(key, value interface{}) bool {
		processedTypes.Delete(key)
		return true
	})

	defs := make(m)
	typ := reflect.TypeOf(TestUser{})

	// First call should process the type
	buildModelDefinition(defs, typ, true)
	if len(defs) == 0 {
		t.Error("First call should create definition")
	}

	// Verify type is cached
	if _, exists := processedTypes.Load(typ); !exists {
		t.Error("Type should be cached after first call")
	}

	// Second call should skip processing (cached)
	initialLen := len(defs)
	buildModelDefinition(defs, typ, true)
	if len(defs) != initialLen {
		t.Error("Second call should not modify definitions (should use cache)")
	}
}

func TestBuildModelDefinition_MapPreSizing(t *testing.T) {
	defs := make(m)
	typ := reflect.TypeOf(TestUser{})

	// Reset cache
	processedTypes.Range(func(key, value interface{}) bool {
		processedTypes.Delete(key)
		return true
	})

	buildModelDefinition(defs, typ, true)

	// Verify definition was created
	typeName := getNameFromType(typ)
	def, ok := defs[typeName]
	if !ok {
		t.Fatalf("Definition for %s should exist", typeName)
	}

	// Check that properties exist
	defMap, ok := def.(m)
	if !ok {
		t.Fatal("Definition should be a map")
	}

	props, ok := defMap["properties"].(m)
	if !ok {
		t.Fatal("Properties should exist in definition")
	}

	// TestUser has 5 fields (id, name, email, age, is_active)
	expectedFields := 5
	if len(props) != expectedFields {
		t.Errorf("Expected %d properties, got %d", expectedFields, len(props))
	}
}

func TestGetFieldName(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{
			name: "Field with json tag",
			field: reflect.StructField{
				Name: "UserID",
				Tag:  `json:"user_id"`,
			},
			expected: "user_id",
		},
		{
			name: "Field with json tag and omitempty",
			field: reflect.StructField{
				Name: "Email",
				Tag:  `json:"email,omitempty"`,
			},
			expected: "email",
		},
		{
			name: "Field without json tag",
			field: reflect.StructField{
				Name: "InternalField",
				Tag:  `validate:"required"`,
			},
			expected: "InternalField",
		},
		{
			name: "Field with empty json tag",
			field: reflect.StructField{
				Name: "Field",
				Tag:  `json:""`,
			},
			expected: "Field", // Falls back to field name when tag is empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFieldName(tt.field)
			if result != tt.expected {
				t.Errorf("getFieldName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildDefinitions_NoParam(t *testing.T) {
	endpoints := []EndpointDef{
		{
			RequestType:  typlect.TypeNoParam,
			ResponseType: reflect.TypeOf(TestResponse{}),
			Group:        "test",
			Name:         "NoParamEndpoint",
			Endpoint:     "/api/test",
			Method:       "GET",
		},
	}

	defs := buildDefinitions(endpoints)

	// NoParam should not create a definition
	noParamName := getNameFromType(typlect.TypeNoParam)
	if _, ok := defs[noParamName]; ok {
		t.Error("NoParam type should not create a definition")
	}

	// But TestResponse should exist
	respName := getNameFromType(reflect.TypeOf(TestResponse{}))
	if _, ok := defs[respName]; !ok {
		t.Error("TestResponse definition should exist")
	}
}

func TestBuildDefinitions_PointerTypes(t *testing.T) {
	endpoints := []EndpointDef{
		{
			RequestType:  reflect.TypeOf(&TestUser{}),
			ResponseType: reflect.TypeOf(&TestResponse{}),
			Group:        "test",
			Name:         "PointerTest",
			Endpoint:     "/api/test",
			Method:       "POST",
		},
	}

	defs := buildDefinitions(endpoints)

	// Should dereference pointers and create definitions for base types
	userName := getNameFromType(reflect.TypeOf(TestUser{}))
	if _, ok := defs[userName]; !ok {
		t.Error("TestUser definition should exist (dereferenced from pointer)")
	}

	respName := getNameFromType(reflect.TypeOf(TestResponse{}))
	if _, ok := defs[respName]; !ok {
		t.Error("TestResponse definition should exist (dereferenced from pointer)")
	}
}

func TestBuildDefinitions_SliceTypes(t *testing.T) {
	endpoints := []EndpointDef{
		{
			RequestType:  reflect.TypeOf([]TestUser{}),
			ResponseType: reflect.TypeOf(TestResponse{}),
			Group:        "test",
			Name:         "SliceTest",
			Endpoint:     "/api/test",
			Method:       "POST",
		},
	}

	defs := buildDefinitions(endpoints)

	// Should extract element type from slice
	userName := getNameFromType(reflect.TypeOf(TestUser{}))
	if _, ok := defs[userName]; !ok {
		t.Error("TestUser definition should exist (extracted from slice)")
	}
}

// Benchmark tests to verify performance improvements

func BenchmarkBuildDefinitions_WithCache(b *testing.B) {
	endpoints := []EndpointDef{
		{RequestType: reflect.TypeOf(TestUser{}), ResponseType: reflect.TypeOf(TestResponse{}), Group: "users", Name: "Create", Endpoint: "/users", Method: "POST"},
		{RequestType: reflect.TypeOf(TestUser{}), ResponseType: reflect.TypeOf(TestResponse{}), Group: "users", Name: "Update", Endpoint: "/users/:id", Method: "PUT"},
		{RequestType: reflect.TypeOf(TestUser{}), ResponseType: reflect.TypeOf(TestResponse{}), Group: "users", Name: "Delete", Endpoint: "/users/:id", Method: "DELETE"},
		{RequestType: reflect.TypeOf(TestProduct{}), ResponseType: reflect.TypeOf(TestResponse{}), Group: "products", Name: "Create", Endpoint: "/products", Method: "POST"},
		{RequestType: reflect.TypeOf(TestProduct{}), ResponseType: reflect.TypeOf(TestResponse{}), Group: "products", Name: "Update", Endpoint: "/products/:id", Method: "PUT"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processedTypes.Range(func(key, value interface{}) bool {
			processedTypes.Delete(key)
			return true
		})
		buildDefinitions(endpoints)
	}
}

func BenchmarkGetFieldName(b *testing.B) {
	field := reflect.StructField{
		Name: "UserID",
		Tag:  `json:"user_id,omitempty"`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getFieldName(field)
	}
}

func BenchmarkBuildModelDefinition_SingleType(b *testing.B) {
	typ := reflect.TypeOf(TestUser{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processedTypes.Range(func(key, value interface{}) bool {
			processedTypes.Delete(key)
			return true
		})
		defs := make(m)
		buildModelDefinition(defs, typ, true)
	}
}

func BenchmarkBuildModelDefinition_NestedType(b *testing.B) {
	typ := reflect.TypeOf(NestedStruct{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processedTypes.Range(func(key, value interface{}) bool {
			processedTypes.Delete(key)
			return true
		})
		defs := make(m)
		buildModelDefinition(defs, typ, true)
	}
}
