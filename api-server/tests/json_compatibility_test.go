package tests

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	goccyjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// GOCCY/GO-JSON COMPATIBILITY TESTS
// ============================================================================
// These tests verify that goccy/go-json handles errors correctly and
// maintains compatibility with encoding/json error behavior.
// ============================================================================

type TestStruct struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// TestGoccyJSON_Marshal_ValidData tests that valid data is marshaled correctly
func TestGoccyJSON_Marshal_ValidData(t *testing.T) {
	data := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	result, err := goccyjson.Marshal(data)

	assert.NoError(t, err)
	assert.Contains(t, string(result), `"name":"John Doe"`)
	assert.Contains(t, string(result), `"email":"john@example.com"`)
	assert.Contains(t, string(result), `"age":30`)
}

// TestGoccyJSON_Unmarshal_ValidJSON tests that valid JSON is unmarshaled correctly
func TestGoccyJSON_Unmarshal_ValidJSON(t *testing.T) {
	validJSON := `{"name":"John Doe","email":"john@example.com","age":30}`

	var result TestStruct
	err := goccyjson.Unmarshal([]byte(validJSON), &result)

	assert.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

// TestGoccyJSON_Unmarshal_InvalidJSON tests that malformed JSON returns an error
func TestGoccyJSON_Unmarshal_InvalidJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "malformed_json",
			input:       `{"name":"John","email":"john@example.com","age":30`,
			description: "missing closing brace",
		},
		{
			name:        "invalid_syntax",
			input:       `{"name":"John","email":"john@example.com","age":}`,
			description: "invalid value",
		},
		{
			name:        "unexpected_token",
			input:       `{"name":"John","email":"john@example.com",,}`,
			description: "double comma",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := goccyjson.Unmarshal([]byte(tt.input), &result)

			assert.Error(t, err, "Expected error for %s", tt.description)
		})
	}
}

// TestGoccyJSON_Unmarshal_TypeMismatch tests type mismatch errors
func TestGoccyJSON_Unmarshal_TypeMismatch(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "string_to_int",
			input: `{"name":"John","email":"john@example.com","age":"thirty"}`,
		},
		{
			name:  "bool_to_string",
			input: `{"name":true,"email":"john@example.com","age":30}`,
		},
		{
			name:  "array_to_struct",
			input: `["John","john@example.com",30]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := goccyjson.Unmarshal([]byte(tt.input), &result)

			assert.Error(t, err, "Expected type mismatch error")
		})
	}
}

// TestGoccyJSON_Unmarshal_EmptyBody tests empty body handling
func TestGoccyJSON_Unmarshal_EmptyBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "empty_string",
			body: "",
		},
		{
			name: "whitespace_only",
			body: "   \n\t  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := goccyjson.Unmarshal([]byte(tt.body), &result)

			assert.Error(t, err, "Expected error for empty body")
		})
	}
}

// TestGoccyJSON_Unmarshal_SpecialCharacters tests special characters handling
func TestGoccyJSON_Unmarshal_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unicode",
			input:    `{"name":"José","email":"jose@example.com","age":30}`,
			expected: "José",
		},
		{
			name:     "emoji",
			input:    `{"name":"John 🚀","email":"john@example.com","age":30}`,
			expected: "John 🚀",
		},
		{
			name:     "escaped_quotes",
			input:    `{"name":"John \"The Rock\" Doe","email":"john@example.com","age":30}`,
			expected: `John "The Rock" Doe`,
		},
		{
			name:     "newlines_and_tabs",
			input:    `{"name":"John\nDoe\tJr","email":"john@example.com","age":30}`,
			expected: "John\nDoe\tJr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := goccyjson.Unmarshal([]byte(tt.input), &result)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Name)
		})
	}
}

// TestGoccyJSON_Unmarshal_LargePayload tests large payload handling
func TestGoccyJSON_Unmarshal_LargePayload(t *testing.T) {
	// Create a large valid JSON payload
	largeString := strings.Repeat("A", 10000)
	input := `{"name":"` + largeString + `","email":"john@example.com","age":30}`

	var result TestStruct
	err := goccyjson.Unmarshal([]byte(input), &result)

	assert.NoError(t, err)
	assert.Equal(t, largeString, result.Name)
}

// TestGoccyJSON_Unmarshal_NestedStructures tests nested JSON structures
func TestGoccyJSON_Unmarshal_NestedStructures(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	input := `{"name":"John","address":{"street":"123 Main St","city":"NYC"}}`

	var result Person
	err := goccyjson.Unmarshal([]byte(input), &result)

	assert.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "123 Main St", result.Address.Street)
	assert.Equal(t, "NYC", result.Address.City)
}

// TestGoccyJSON_Unmarshal_Arrays tests array handling
func TestGoccyJSON_Unmarshal_Arrays(t *testing.T) {
	type PersonList struct {
		People []TestStruct `json:"people"`
	}

	input := `{"people":[{"name":"John","email":"john@example.com","age":30},{"name":"Jane","email":"jane@example.com","age":25}]}`

	var result PersonList
	err := goccyjson.Unmarshal([]byte(input), &result)

	assert.NoError(t, err)
	assert.Len(t, result.People, 2)
	assert.Equal(t, "John", result.People[0].Name)
	assert.Equal(t, "Jane", result.People[1].Name)
}

// TestGoccyJSON_Unmarshal_NullValues tests null value handling
func TestGoccyJSON_Unmarshal_NullValues(t *testing.T) {
	type NullableStruct struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
		Age   *int    `json:"age"`
	}

	input := `{"name":"John","email":null,"age":null}`

	var result NullableStruct
	err := goccyjson.Unmarshal([]byte(input), &result)

	assert.NoError(t, err)
	require.NotNil(t, result.Name)
	assert.Equal(t, "John", *result.Name)
	assert.Nil(t, result.Email)
	assert.Nil(t, result.Age)
}

// TestGoccyJSON_Unmarshal_UnknownFields tests handling of unknown fields
func TestGoccyJSON_Unmarshal_UnknownFields(t *testing.T) {
	// JSON with extra fields that don't exist in struct
	input := `{"name":"John","email":"john@example.com","age":30,"extra_field":"ignored","another_field":123}`

	var result TestStruct
	err := goccyjson.Unmarshal([]byte(input), &result)

	// goccy/go-json should ignore unknown fields by default (same as encoding/json)
	assert.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

// TestGoccyJSON_Decoder tests that Decoder works correctly
func TestGoccyJSON_Decoder(t *testing.T) {
	input := `{"name":"John Doe","email":"john@example.com","age":30}`
	reader := bytes.NewReader([]byte(input))

	decoder := goccyjson.NewDecoder(reader)
	var result TestStruct
	err := decoder.Decode(&result)

	assert.NoError(t, err)
	assert.Equal(t, "John Doe", result.Name)
}

// TestGoccyJSON_Encoder tests that Encoder works correctly
func TestGoccyJSON_Encoder(t *testing.T) {
	data := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	var buf bytes.Buffer
	encoder := goccyjson.NewEncoder(&buf)
	err := encoder.Encode(data)

	assert.NoError(t, err)
	result := buf.String()
	assert.Contains(t, result, `"name":"John Doe"`)
	assert.Contains(t, result, `"email":"john@example.com"`)
	assert.Contains(t, result, `"age":30`)
}

// TestGoccyJSON_vs_EncodingJSON_Compatibility compares goccy vs encoding/json
func TestGoccyJSON_vs_EncodingJSON_Compatibility(t *testing.T) {
	data := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	// Marshal with both libraries
	goccyResult, goccyErr := goccyjson.Marshal(data)
	stdResult, stdErr := json.Marshal(data)

	// Both should succeed
	assert.NoError(t, goccyErr)
	assert.NoError(t, stdErr)

	// Results should be compatible (may have different formatting but same data)
	var goccyParsed, stdParsed TestStruct
	assert.NoError(t, goccyjson.Unmarshal(goccyResult, &goccyParsed))
	assert.NoError(t, goccyjson.Unmarshal(stdResult, &stdParsed))

	assert.Equal(t, goccyParsed, stdParsed)
}

// TestGoccyJSON_ErrorMessages tests that error messages are helpful
func TestGoccyJSON_ErrorMessages(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unexpected_eof",
			input: `{"name":"John"`,
		},
		{
			name:  "invalid_character",
			input: `{"name":"John","age":30,}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := goccyjson.Unmarshal([]byte(tt.input), &result)

			assert.Error(t, err)
			assert.NotEmpty(t, err.Error(), "Error message should not be empty")
		})
	}
}
