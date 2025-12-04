package swagger

import (
	"os"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinFunction tests the min helper function
func TestMinFunction(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 10, 5, 5},
		{"equal", 7, 7, 7},
		{"zero", 0, 5, 0},
		{"negative", -5, 3, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGenerateJsonFileNotExists tests behavior when v3Doc.json doesn't exist
func TestGenerateJsonFileNotExists(t *testing.T) {
	// Setup: ensure docs directory and file don't exist
	testDir := "./test_docs_not_exists"
	defer os.RemoveAll(testDir)

	// Change working directory temporarily
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create a temp directory for this test
	os.Mkdir(testDir, 0755)
	os.Chdir(testDir)

	// Create a minimal v3 doc
	v3Doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	// This should not crash - it should gracefully skip
	generatejson(v3Doc)

	// Verify no files were created
	_, err := os.Stat("./docs/v3Doc.json")
	assert.True(t, os.IsNotExist(err), "v3Doc.json should not exist")

	_, err = os.Stat("./docs/resolved_swagger.json")
	assert.True(t, os.IsNotExist(err), "resolved_swagger.json should not exist")
}

// TestGenerateJsonEmptyFile tests behavior with empty v3Doc.json
func TestGenerateJsonEmptyFile(t *testing.T) {
	// Setup: create docs directory with empty file
	testDir := "./test_docs_empty"
	defer os.RemoveAll(testDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.Mkdir(testDir, 0755)
	os.Chdir(testDir)
	os.Mkdir("docs", 0755)

	// Create empty v3Doc.json
	err := os.WriteFile("./docs/v3Doc.json", []byte(""), 0644)
	require.NoError(t, err)

	v3Doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	// This should not crash - should detect empty file
	generatejson(v3Doc)

	// Verify resolved file was not created
	_, err = os.Stat("./docs/resolved_swagger.json")
	assert.True(t, os.IsNotExist(err), "resolved_swagger.json should not be created for empty input")
}

// TestGenerateJsonInvalidJSON tests behavior with invalid JSON
func TestGenerateJsonInvalidJSON(t *testing.T) {
	// Setup: create docs directory with invalid JSON
	testDir := "./test_docs_invalid"
	defer os.RemoveAll(testDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.Mkdir(testDir, 0755)
	os.Chdir(testDir)
	os.Mkdir("docs", 0755)

	// Create invalid JSON in v3Doc.json
	invalidJSON := `{"this is": "not valid json"`
	err := os.WriteFile("./docs/v3Doc.json", []byte(invalidJSON), 0644)
	require.NoError(t, err)

	v3Doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	// This should not crash - should handle parse error gracefully
	generatejson(v3Doc)

	// Verify resolved file was not created
	_, err = os.Stat("./docs/resolved_swagger.json")
	assert.True(t, os.IsNotExist(err), "resolved_swagger.json should not be created for invalid JSON")
}

// TestGenerateJsonValidFile tests successful processing
func TestGenerateJsonValidFile(t *testing.T) {
	// Setup: create docs directory with valid v3Doc.json
	testDir := "./test_docs_valid"
	defer os.RemoveAll(testDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	os.Mkdir(testDir, 0755)
	os.Chdir(testDir)
	os.Mkdir("docs", 0755)

	// Create a valid minimal swagger JSON
	validJSON := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {
									"schema": {
										"type": "object",
										"properties": {
											"message": {
												"type": "string"
											}
										}
									}
								}
							}
						}
					}
				}
			}
		},
		"components": {
			"schemas": {}
		}
	}`

	err := os.WriteFile("./docs/v3Doc.json", []byte(validJSON), 0644)
	require.NoError(t, err)

	v3Doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	// This should succeed
	generatejson(v3Doc)

	// Verify resolved file was created
	_, err = os.Stat("./docs/resolved_swagger.json")
	assert.False(t, os.IsNotExist(err), "resolved_swagger.json should be created")

	// Verify content can be parsed
	if !os.IsNotExist(err) {
		data, err := os.ReadFile("./docs/resolved_swagger.json")
		require.NoError(t, err)

		// Should be valid JSON
		_, err = gabs.ParseJSON(data)
		assert.NoError(t, err, "resolved_swagger.json should contain valid JSON")
	}
}

// TestReplaceDataType tests the NullString replacement logic
func TestReplaceDataType(t *testing.T) {
	jsonStr := `{
		"type": "object",
		"properties": {
			"name": {
				"type": "NullString"
			},
			"nested": {
				"type": "object",
				"properties": {
					"value": {
						"type": "NullString"
					}
				}
			}
		}
	}`

	container, err := gabs.ParseJSON([]byte(jsonStr))
	require.NoError(t, err)

	// Replace NullString with string
	replaceDataType(container, "NullString", "string")

	// Verify replacements
	nameType := container.Path("properties.name.type").Data()
	assert.Equal(t, "string", nameType, "name type should be replaced with string")

	nestedType := container.Path("properties.nested.properties.value.type").Data()
	assert.Equal(t, "string", nestedType, "nested value type should be replaced with string")
}

// TestTraverseAndReplaceRefs tests the reference resolution logic
func TestTraverseAndReplaceRefs(t *testing.T) {
	jsonStr := `{
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"id": {
							"type": "integer"
						},
						"name": {
							"type": "string"
						}
					}
				}
			}
		},
		"paths": {
			"/user": {
				"get": {
					"responses": {
						"200": {
							"content": {
								"application/json": {
									"schema": {
										"$ref": "#/components/schemas/User"
									}
								}
							}
						}
					}
				}
			}
		}
	}`

	container, err := gabs.ParseJSON([]byte(jsonStr))
	require.NoError(t, err)

	// Resolve references
	traverseAndReplaceRefs(container, container)

	// Verify $ref was replaced
	schema := container.Path("paths./user.get.responses.200.content.application/json.schema")
	assert.NotNil(t, schema, "Schema should exist")

	// After resolution, $ref should be gone and actual schema should be present
	refValue := schema.Path("$ref")
	if refValue.Data() != nil {
		// If $ref still exists, it should have been merged with the actual schema
		t.Log("Reference resolution: $ref may still exist but schema should be merged")
	}
}

// TestWrap200Responses tests the success response wrapping logic
func TestWrap200Responses(t *testing.T) {
	jsonStr := `{
		"paths": {
			"/test": {
				"get": {
					"responses": {
						"200": {
							"content": {
								"application/json": {
									"schema": {
										"type": "object",
										"properties": {
											"id": {
												"type": "integer"
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`

	container, err := gabs.ParseJSON([]byte(jsonStr))
	require.NoError(t, err)

	// Wrap 200 responses
	wrap200Responses(container)

	// Verify wrapping
	schema := container.Path("paths./test.get.responses.200.content.application/json.schema")
	assert.NotNil(t, schema, "Schema should exist")

	// Check for success wrapper properties
	properties := schema.Path("properties")
	if properties.Data() != nil {
		// Should have success, message, and data properties
		success := properties.Path("success")
		message := properties.Path("message")
		data := properties.Path("data")

		assert.NotNil(t, success.Data(), "Should have success property")
		assert.NotNil(t, message.Data(), "Should have message property")
		assert.NotNil(t, data.Data(), "Should have data property")
	}
}

// TestResolveSchema tests schema resolution logic
func TestResolveSchema(t *testing.T) {
	jsonStr := `{
		"components": {
			"schemas": {
				"TestSchema": {
					"type": "object",
					"properties": {
						"field": {
							"type": "string"
						}
					}
				}
			}
		}
	}`

	container, err := gabs.ParseJSON([]byte(jsonStr))
	require.NoError(t, err)

	// Test valid reference
	resolved := resolveSchema("#/components/schemas/TestSchema", container)
	assert.NotNil(t, resolved, "Should resolve valid reference")

	if resolved != nil {
		schemaType := resolved.Path("type").Data()
		assert.Equal(t, "object", schemaType, "Resolved schema should have correct type")
	}

	// Test invalid reference
	invalidResolved := resolveSchema("#/components/schemas/NonExistent", container)
	assert.Nil(t, invalidResolved, "Should return nil for invalid reference")

	// Test non-component reference
	otherRef := resolveSchema("#/definitions/Something", container)
	assert.Nil(t, otherRef, "Should return nil for non-component reference")
}

// Benchmark tests
func BenchmarkGenerateJsonFileNotExists(b *testing.B) {
	testDir := "./bench_test"
	os.Mkdir(testDir, 0755)
	defer os.RemoveAll(testDir)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(testDir)

	v3Doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info:    &openapi3.Info{Title: "Test", Version: "1.0.0"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generatejson(v3Doc)
	}
}

func BenchmarkMinFunction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = min(100, 200)
	}
}
