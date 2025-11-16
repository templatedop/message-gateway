package benchmarks

import (
	"bytes"
	"testing"

	goccyjson "github.com/goccy/go-json"
	"github.com/bytedance/sonic"
)

// ============================================================================
// JSON LIBRARY COMPARISON BENCHMARKS
// ============================================================================
// This file compares performance between:
// 1. goccy/go-json (CURRENT implementation)
// 2. bytedance/sonic (PROPOSED replacement)
//
// Goal: Measure actual performance improvement from migrating to sonic
// ============================================================================

// Test data structures representing typical API payloads

type SmallPayload struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Active  bool   `json:"active"`
}

type MediumPayload struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	Age       int               `json:"age"`
	Country   string            `json:"country"`
	City      string            `json:"city"`
	Address   string            `json:"address"`
	Phone     string            `json:"phone"`
	Active    bool              `json:"active"`
	Metadata  map[string]string `json:"metadata"`
	Tags      []string          `json:"tags"`
	Timestamp int64             `json:"timestamp"`
}

type LargePayload struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Email        string              `json:"email"`
	Age          int                 `json:"age"`
	Country      string              `json:"country"`
	City         string              `json:"city"`
	Address      string              `json:"address"`
	Phone        string              `json:"phone"`
	Active       bool                `json:"active"`
	Metadata     map[string]string   `json:"metadata"`
	Tags         []string            `json:"tags"`
	Timestamp    int64               `json:"timestamp"`
	Description  string              `json:"description"`
	Preferences  map[string]any `json:"preferences"`
	History      []map[string]string `json:"history"`
	RelatedUsers []SmallPayload      `json:"related_users"`
}

// Sample data
var (
	smallData = SmallPayload{
		ID:     "user-12345",
		Name:   "John Doe",
		Email:  "john@example.com",
		Active: true,
	}

	mediumData = MediumPayload{
		ID:      "user-12345",
		Name:    "John Doe",
		Email:   "john@example.com",
		Age:     30,
		Country: "USA",
		City:    "New York",
		Address: "123 Main St, Apt 4B",
		Phone:   "+1-555-0123",
		Active:  true,
		Metadata: map[string]string{
			"created_by": "admin",
			"source":     "api",
			"version":    "2.0",
		},
		Tags:      []string{"premium", "verified", "active"},
		Timestamp: 1705429200,
	}

	largeData = LargePayload{
		ID:      "user-12345",
		Name:    "John Doe",
		Email:   "john@example.com",
		Age:     30,
		Country: "USA",
		City:    "New York",
		Address: "123 Main St, Apt 4B, New York, NY 10001",
		Phone:   "+1-555-0123",
		Active:  true,
		Metadata: map[string]string{
			"created_by":   "admin",
			"source":       "api",
			"version":      "2.0",
			"team":         "engineering",
			"department":   "technology",
			"cost_center":  "CC-12345",
			"project_code": "PROJ-789",
		},
		Tags:        []string{"premium", "verified", "active", "enterprise", "priority"},
		Timestamp:   1705429200,
		Description: "This is a comprehensive user profile with extended metadata and historical tracking capabilities. The user has been active for several years and maintains a premium subscription with enterprise-level access.",
		Preferences: map[string]any{
			"theme":            "dark",
			"language":         "en-US",
			"notifications":    true,
			"email_frequency":  "daily",
			"timezone":         "America/New_York",
			"two_factor":       true,
			"analytics_opt_in": false,
		},
		History: []map[string]string{
			{"action": "login", "timestamp": "2025-01-16T10:00:00Z", "ip": "192.168.1.1"},
			{"action": "update_profile", "timestamp": "2025-01-16T11:30:00Z", "ip": "192.168.1.1"},
			{"action": "purchase", "timestamp": "2025-01-16T14:45:00Z", "ip": "192.168.1.1"},
			{"action": "logout", "timestamp": "2025-01-16T18:00:00Z", "ip": "192.168.1.1"},
		},
		RelatedUsers: []SmallPayload{
			{ID: "user-11111", Name: "Alice Smith", Email: "alice@example.com", Active: true},
			{ID: "user-22222", Name: "Bob Johnson", Email: "bob@example.com", Active: true},
			{ID: "user-33333", Name: "Carol White", Email: "carol@example.com", Active: false},
		},
	}
)

// ============================================================================
// MARSHAL BENCHMARKS (Encoding/Serialization)
// ============================================================================

// Small Payload Marshaling
func BenchmarkGoccyJSON_Marshal_Small(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := goccyjson.Marshal(smallData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSonic_Marshal_Small(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(smallData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Medium Payload Marshaling
func BenchmarkGoccyJSON_Marshal_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := goccyjson.Marshal(mediumData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSonic_Marshal_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(mediumData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Large Payload Marshaling
func BenchmarkGoccyJSON_Marshal_Large(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := goccyjson.Marshal(largeData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSonic_Marshal_Large(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(largeData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// UNMARSHAL BENCHMARKS (Decoding/Deserialization)
// ============================================================================

var (
	smallJSON  []byte
	mediumJSON []byte
	largeJSON  []byte
)

func init() {
	var err error
	smallJSON, err = goccyjson.Marshal(smallData)
	if err != nil {
		panic(err)
	}
	mediumJSON, err = goccyjson.Marshal(mediumData)
	if err != nil {
		panic(err)
	}
	largeJSON, err = goccyjson.Marshal(largeData)
	if err != nil {
		panic(err)
	}
}

// Small Payload Unmarshaling
func BenchmarkGoccyJSON_Unmarshal_Small(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result SmallPayload
		err := goccyjson.Unmarshal(smallJSON, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSonic_Unmarshal_Small(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result SmallPayload
		err := sonic.Unmarshal(smallJSON, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Medium Payload Unmarshaling
func BenchmarkGoccyJSON_Unmarshal_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result MediumPayload
		err := goccyjson.Unmarshal(mediumJSON, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSonic_Unmarshal_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result MediumPayload
		err := sonic.Unmarshal(mediumJSON, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Large Payload Unmarshaling
func BenchmarkGoccyJSON_Unmarshal_Large(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result LargePayload
		err := goccyjson.Unmarshal(largeJSON, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSonic_Unmarshal_Large(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result LargePayload
		err := sonic.Unmarshal(largeJSON, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// DECODER BENCHMARKS (Stream-based deserialization)
// ============================================================================

func BenchmarkGoccyJSON_Decoder_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result MediumPayload
		decoder := goccyjson.NewDecoder(bytes.NewReader(mediumJSON))
		err := decoder.Decode(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSonic_Decoder_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result MediumPayload
		decoder := sonic.ConfigDefault.NewDecoder(bytes.NewReader(mediumJSON))
		err := decoder.Decode(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// ENCODER BENCHMARKS (Stream-based serialization)
// ============================================================================

func BenchmarkGoccyJSON_Encoder_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		encoder := goccyjson.NewEncoder(&buf)
		err := encoder.Encode(mediumData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSonic_Encoder_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		encoder := sonic.ConfigDefault.NewEncoder(&buf)
		err := encoder.Encode(mediumData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// PARALLEL BENCHMARKS (High Concurrency)
// ============================================================================

func BenchmarkGoccyJSON_Marshal_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := goccyjson.Marshal(mediumData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkSonic_Marshal_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := sonic.Marshal(mediumData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkGoccyJSON_Unmarshal_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var result MediumPayload
			err := goccyjson.Unmarshal(mediumJSON, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkSonic_Unmarshal_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var result MediumPayload
			err := sonic.Unmarshal(mediumJSON, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
