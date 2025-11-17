package benchmarks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	goccyjson "github.com/goccy/go-json"
)

// ============================================================================
// API SERVER JSON BENCHMARKS - BEFORE vs AFTER
// ============================================================================
// These benchmarks compare API server performance using:
// - BEFORE: Gin's default binding with encoding/json
// - AFTER: CustomJSONBinding with goccy/go-json
//
// This measures real-world performance impact on the actual API server.
// ============================================================================

// Test request structures matching common API patterns
type SmallRequest struct {
	ID     string `json:"id" binding:"required"`
	Action string `json:"action" binding:"required"`
}

type MediumRequest struct {
	ID       string            `json:"id" binding:"required"`
	Name     string            `json:"name" binding:"required"`
	Email    string            `json:"email" binding:"required,email"`
	Age      int               `json:"age" binding:"required,gte=0,lte=150"`
	Country  string            `json:"country"`
	City     string            `json:"city"`
	Address  string            `json:"address"`
	Phone    string            `json:"phone"`
	Metadata map[string]string `json:"metadata"`
	Tags     []string          `json:"tags"`
}

type LargeRequest struct {
	ID          string                 `json:"id" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Email       string                 `json:"email" binding:"required,email"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Tags        []string               `json:"tags"`
	Nested      NestedData             `json:"nested"`
	Items       []Item                 `json:"items"`
	Attributes  map[string]string      `json:"attributes"`
	Timestamp   int64                  `json:"timestamp"`
}

type NestedData struct {
	Level1 string                 `json:"level1"`
	Level2 map[string]interface{} `json:"level2"`
}

type Item struct {
	ItemID   string  `json:"item_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// Response structures
type SmallResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type MediumResponse struct {
	Status    string            `json:"status"`
	Message   string            `json:"message"`
	Data      interface{}       `json:"data"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp int64             `json:"timestamp"`
}

type LargeResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Items     []Item                 `json:"items"`
	Metadata  map[string]string      `json:"metadata"`
	Nested    NestedData             `json:"nested"`
	Timestamp int64                  `json:"timestamp"`
}

// ============================================================================
// CustomJSONBinding - goccy/go-json implementation (AFTER)
// ============================================================================

type CustomJSONBinding struct{}

func (CustomJSONBinding) Name() string {
	return "json"
}

func (CustomJSONBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return io.EOF
	}
	decoder := goccyjson.NewDecoder(req.Body)
	return decoder.Decode(obj)
}

func (CustomJSONBinding) BindBody(body []byte, obj interface{}) error {
	return goccyjson.Unmarshal(body, obj)
}

// ============================================================================
// BENCHMARK 1: Small Request Binding
// ============================================================================

func BenchmarkAPIServer_Before_SmallRequestBinding(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	// Use Gin's default JSON binding (encoding/json)
	binding.JSON = binding.Default("POST", "application/json").(binding.BindingBody)

	reqData := SmallRequest{
		ID:     "req-12345",
		Action: "process",
	}
	jsonBody, _ := json.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/v1/action", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		var result SmallRequest
		_ = binding.JSON.BindBody(jsonBody, &result)
	}
}

func BenchmarkAPIServer_After_SmallRequestBinding(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	// Use CustomJSONBinding with goccy/go-json
	customBinding := CustomJSONBinding{}

	reqData := SmallRequest{
		ID:     "req-12345",
		Action: "process",
	}
	jsonBody, _ := goccyjson.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/v1/action", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		var result SmallRequest
		_ = customBinding.BindBody(jsonBody, &result)
	}
}

// ============================================================================
// BENCHMARK 2: Medium Request Binding
// ============================================================================

func BenchmarkAPIServer_Before_MediumRequestBinding(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	binding.JSON = binding.Default("POST", "application/json").(binding.BindingBody)

	reqData := MediumRequest{
		ID:      "req-67890",
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Age:     30,
		Country: "USA",
		City:    "New York",
		Address: "123 Main St",
		Phone:   "+1-555-0100",
		Metadata: map[string]string{
			"source":  "api",
			"version": "v1",
			"client":  "web",
		},
		Tags: []string{"urgent", "customer", "verified"},
	}
	jsonBody, _ := json.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result MediumRequest
		_ = binding.JSON.BindBody(jsonBody, &result)
	}
}

func BenchmarkAPIServer_After_MediumRequestBinding(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	customBinding := CustomJSONBinding{}

	reqData := MediumRequest{
		ID:      "req-67890",
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Age:     30,
		Country: "USA",
		City:    "New York",
		Address: "123 Main St",
		Phone:   "+1-555-0100",
		Metadata: map[string]string{
			"source":  "api",
			"version": "v1",
			"client":  "web",
		},
		Tags: []string{"urgent", "customer", "verified"},
	}
	jsonBody, _ := goccyjson.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result MediumRequest
		_ = customBinding.BindBody(jsonBody, &result)
	}
}

// ============================================================================
// BENCHMARK 3: Large Request Binding
// ============================================================================

func BenchmarkAPIServer_Before_LargeRequestBinding(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	binding.JSON = binding.Default("POST", "application/json").(binding.BindingBody)

	reqData := LargeRequest{
		ID:          "req-99999",
		Name:        "Complex Transaction",
		Email:       "transaction@example.com",
		Description: "This is a complex transaction with nested data and multiple items",
		Metadata: map[string]interface{}{
			"type":     "transaction",
			"priority": 1,
			"verified": true,
		},
		Tags: []string{"important", "urgent", "verified", "customer"},
		Nested: NestedData{
			Level1: "nested-value",
			Level2: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
		},
		Items: []Item{
			{ItemID: "item-1", Quantity: 10, Price: 99.99},
			{ItemID: "item-2", Quantity: 5, Price: 149.99},
			{ItemID: "item-3", Quantity: 2, Price: 299.99},
		},
		Attributes: map[string]string{
			"attr1": "value1",
			"attr2": "value2",
			"attr3": "value3",
		},
		Timestamp: 1699564800,
	}
	jsonBody, _ := json.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result LargeRequest
		_ = binding.JSON.BindBody(jsonBody, &result)
	}
}

func BenchmarkAPIServer_After_LargeRequestBinding(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	customBinding := CustomJSONBinding{}

	reqData := LargeRequest{
		ID:          "req-99999",
		Name:        "Complex Transaction",
		Email:       "transaction@example.com",
		Description: "This is a complex transaction with nested data and multiple items",
		Metadata: map[string]interface{}{
			"type":     "transaction",
			"priority": 1,
			"verified": true,
		},
		Tags: []string{"important", "urgent", "verified", "customer"},
		Nested: NestedData{
			Level1: "nested-value",
			Level2: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
		},
		Items: []Item{
			{ItemID: "item-1", Quantity: 10, Price: 99.99},
			{ItemID: "item-2", Quantity: 5, Price: 149.99},
			{ItemID: "item-3", Quantity: 2, Price: 299.99},
		},
		Attributes: map[string]string{
			"attr1": "value1",
			"attr2": "value2",
			"attr3": "value3",
		},
		Timestamp: 1699564800,
	}
	jsonBody, _ := goccyjson.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result LargeRequest
		_ = customBinding.BindBody(jsonBody, &result)
	}
}

// ============================================================================
// BENCHMARK 4: Response Marshaling
// ============================================================================

func BenchmarkAPIServer_Before_SmallResponseMarshal(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	resp := SmallResponse{
		Status:  "success",
		Message: "Operation completed successfully",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}

func BenchmarkAPIServer_After_SmallResponseMarshal(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	resp := SmallResponse{
		Status:  "success",
		Message: "Operation completed successfully",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = goccyjson.Marshal(resp)
	}
}

func BenchmarkAPIServer_Before_MediumResponseMarshal(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	resp := MediumResponse{
		Status:  "success",
		Message: "Request processed",
		Data: map[string]interface{}{
			"id":   "resp-123",
			"name": "Result Data",
		},
		Metadata: map[string]string{
			"version": "v1",
			"region":  "us-east-1",
		},
		Timestamp: 1699564800,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}

func BenchmarkAPIServer_After_MediumResponseMarshal(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	resp := MediumResponse{
		Status:  "success",
		Message: "Request processed",
		Data: map[string]interface{}{
			"id":   "resp-123",
			"name": "Result Data",
		},
		Metadata: map[string]string{
			"version": "v1",
			"region":  "us-east-1",
		},
		Timestamp: 1699564800,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = goccyjson.Marshal(resp)
	}
}

func BenchmarkAPIServer_Before_LargeResponseMarshal(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	resp := LargeResponse{
		Status:  "success",
		Message: "Complex operation completed",
		Data: map[string]interface{}{
			"id":          "resp-999",
			"transaction": "complete",
			"verified":    true,
		},
		Items: []Item{
			{ItemID: "item-1", Quantity: 10, Price: 99.99},
			{ItemID: "item-2", Quantity: 5, Price: 149.99},
			{ItemID: "item-3", Quantity: 2, Price: 299.99},
		},
		Metadata: map[string]string{
			"version": "v1",
			"region":  "us-east-1",
			"zone":    "1a",
		},
		Nested: NestedData{
			Level1: "response-nested",
			Level2: map[string]interface{}{
				"key1": "value1",
				"key2": 456,
			},
		},
		Timestamp: 1699564800,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}

func BenchmarkAPIServer_After_LargeResponseMarshal(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	resp := LargeResponse{
		Status:  "success",
		Message: "Complex operation completed",
		Data: map[string]interface{}{
			"id":          "resp-999",
			"transaction": "complete",
			"verified":    true,
		},
		Items: []Item{
			{ItemID: "item-1", Quantity: 10, Price: 99.99},
			{ItemID: "item-2", Quantity: 5, Price: 149.99},
			{ItemID: "item-3", Quantity: 2, Price: 299.99},
		},
		Metadata: map[string]string{
			"version": "v1",
			"region":  "us-east-1",
			"zone":    "1a",
		},
		Nested: NestedData{
			Level1: "response-nested",
			Level2: map[string]interface{}{
				"key1": "value1",
				"key2": 456,
			},
		},
		Timestamp: 1699564800,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = goccyjson.Marshal(resp)
	}
}

// ============================================================================
// BENCHMARK 5: Full Request/Response Cycle
// ============================================================================

func BenchmarkAPIServer_Before_FullCycle_Small(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	binding.JSON = binding.Default("POST", "application/json").(binding.BindingBody)

	reqData := SmallRequest{
		ID:     "req-12345",
		Action: "process",
	}
	reqJSON, _ := json.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Unmarshal request
		var req SmallRequest
		_ = binding.JSON.BindBody(reqJSON, &req)

		// Process (simulated)
		resp := SmallResponse{
			Status:  "success",
			Message: "Processed: " + req.Action,
		}

		// Marshal response
		_, _ = json.Marshal(resp)
	}
}

func BenchmarkAPIServer_After_FullCycle_Small(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	customBinding := CustomJSONBinding{}

	reqData := SmallRequest{
		ID:     "req-12345",
		Action: "process",
	}
	reqJSON, _ := goccyjson.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Unmarshal request
		var req SmallRequest
		_ = customBinding.BindBody(reqJSON, &req)

		// Process (simulated)
		resp := SmallResponse{
			Status:  "success",
			Message: "Processed: " + req.Action,
		}

		// Marshal response
		_, _ = goccyjson.Marshal(resp)
	}
}

func BenchmarkAPIServer_Before_FullCycle_Medium(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	binding.JSON = binding.Default("POST", "application/json").(binding.BindingBody)

	reqData := MediumRequest{
		ID:      "req-67890",
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Age:     30,
		Country: "USA",
		City:    "New York",
		Address: "123 Main St",
		Phone:   "+1-555-0100",
		Metadata: map[string]string{
			"source": "api",
		},
		Tags: []string{"urgent"},
	}
	reqJSON, _ := json.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var req MediumRequest
		_ = binding.JSON.BindBody(reqJSON, &req)

		resp := MediumResponse{
			Status:  "success",
			Message: "User " + req.Name + " processed",
			Data:    map[string]interface{}{"id": req.ID},
			Metadata: map[string]string{
				"version": "v1",
			},
			Timestamp: 1699564800,
		}

		_, _ = json.Marshal(resp)
	}
}

func BenchmarkAPIServer_After_FullCycle_Medium(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	customBinding := CustomJSONBinding{}

	reqData := MediumRequest{
		ID:      "req-67890",
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Age:     30,
		Country: "USA",
		City:    "New York",
		Address: "123 Main St",
		Phone:   "+1-555-0100",
		Metadata: map[string]string{
			"source": "api",
		},
		Tags: []string{"urgent"},
	}
	reqJSON, _ := goccyjson.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var req MediumRequest
		_ = customBinding.BindBody(reqJSON, &req)

		resp := MediumResponse{
			Status:  "success",
			Message: "User " + req.Name + " processed",
			Data:    map[string]interface{}{"id": req.ID},
			Metadata: map[string]string{
				"version": "v1",
			},
			Timestamp: 1699564800,
		}

		_, _ = goccyjson.Marshal(resp)
	}
}

func BenchmarkAPIServer_Before_FullCycle_Large(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	binding.JSON = binding.Default("POST", "application/json").(binding.BindingBody)

	reqData := LargeRequest{
		ID:          "req-99999",
		Name:        "Complex Transaction",
		Email:       "transaction@example.com",
		Description: "Complex transaction",
		Metadata: map[string]interface{}{
			"type": "transaction",
		},
		Tags: []string{"important"},
		Nested: NestedData{
			Level1: "value",
			Level2: map[string]interface{}{
				"key": "val",
			},
		},
		Items: []Item{
			{ItemID: "item-1", Quantity: 10, Price: 99.99},
		},
		Attributes: map[string]string{
			"attr": "value",
		},
		Timestamp: 1699564800,
	}
	reqJSON, _ := json.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var req LargeRequest
		_ = binding.JSON.BindBody(reqJSON, &req)

		resp := LargeResponse{
			Status:  "success",
			Message: "Transaction processed",
			Data: map[string]interface{}{
				"id": req.ID,
			},
			Items:     req.Items,
			Metadata:  map[string]string{"version": "v1"},
			Nested:    req.Nested,
			Timestamp: req.Timestamp,
		}

		_, _ = json.Marshal(resp)
	}
}

func BenchmarkAPIServer_After_FullCycle_Large(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	customBinding := CustomJSONBinding{}

	reqData := LargeRequest{
		ID:          "req-99999",
		Name:        "Complex Transaction",
		Email:       "transaction@example.com",
		Description: "Complex transaction",
		Metadata: map[string]interface{}{
			"type": "transaction",
		},
		Tags: []string{"important"},
		Nested: NestedData{
			Level1: "value",
			Level2: map[string]interface{}{
				"key": "val",
			},
		},
		Items: []Item{
			{ItemID: "item-1", Quantity: 10, Price: 99.99},
		},
		Attributes: map[string]string{
			"attr": "value",
		},
		Timestamp: 1699564800,
	}
	reqJSON, _ := goccyjson.Marshal(reqData)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var req LargeRequest
		_ = customBinding.BindBody(reqJSON, &req)

		resp := LargeResponse{
			Status:  "success",
			Message: "Transaction processed",
			Data: map[string]interface{}{
				"id": req.ID,
			},
			Items:     req.Items,
			Metadata:  map[string]string{"version": "v1"},
			Nested:    req.Nested,
			Timestamp: req.Timestamp,
		}

		_, _ = goccyjson.Marshal(resp)
	}
}
