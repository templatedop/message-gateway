package log_test

import (
	"bytes"
	"context"
	"errors"
	"time"

	log "MgApplication/api-log"

	"github.com/rs/zerolog"
)

// Example demonstrating WithFields API for simple structured logging
func ExampleInfoWithFields() {
	// Setup logger
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()

	// Log with structured fields using a map
	log.InfoWithFields(ctx, "user logged in", map[string]interface{}{
		"user_id":  "12345",
		"username": "john_doe",
		"ip":       "192.168.1.100",
		"success":  true,
	})
}

// Example showing ErrorWithFields with error object
func ExampleErrorWithFields() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.DebugLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()
	err := errors.New("connection refused")

	// Log error with additional context
	log.ErrorWithFields(ctx, "database connection failed", map[string]interface{}{
		"error":    err,
		"host":     "db.example.com",
		"port":     5432,
		"database": "users_db",
		"retries":  3,
		"duration": 5 * time.Second,
	})
}

// Example demonstrating WithFields with multiple data types
func ExampleDebugWithFields_multipleTypes() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.DebugLevel),
		log.WithOutputWriter(&buf),
	)

	// WithFields supports many types
	log.DebugWithFields(nil, "processing request", map[string]interface{}{
		"string_field":  "value",
		"int_field":     42,
		"int64_field":   int64(9876543210),
		"float_field":   3.14159,
		"bool_field":    true,
		"strings_field": []string{"tag1", "tag2", "tag3"},
		"ints_field":    []int{1, 2, 3},
	})
}

// Example demonstrating tags support
func ExampleWithTags() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()

	// Add tags to context - these will be included in ALL logs with this context
	ctx = log.WithTags(ctx, "database", "payment")

	// All logs with this context will include tags: ["database", "payment"]
	log.Info(ctx, "processing payment transaction")
	log.InfoEvent(ctx).Str("transaction_id", "tx-123").Msg("transaction completed")
}

// Example showing tags with multiple additions
func ExampleWithTags_multiple() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()

	// Tags accumulate
	ctx = log.WithTags(ctx, "api", "v1")
	ctx = log.WithTags(ctx, "auth")

	// This log will have tags: ["api", "v1", "auth"]
	log.Info(ctx, "user authenticated")
}

// Example showing tags with WithFields
func ExampleWithTags_withFields() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()
	ctx = log.WithTags(ctx, "database", "query")

	// Combine tags with structured fields
	log.InfoWithFields(ctx, "query executed successfully", map[string]interface{}{
		"table":    "users",
		"rows":     150,
		"duration": 45 * time.Millisecond,
	})
	// Output will include: tags: ["database", "query"], table: "users", rows: 150, duration: 45ms
}

// Example showing complete workflow with tags and fields
func Example_tagsAndFieldsWorkflow() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("payment-service"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	// Create context with tags for the entire request
	ctx := context.Background()
	ctx = log.WithTags(ctx, "payment", "stripe")

	// Simple logging - includes tags automatically
	log.Info(ctx, "starting payment processing")

	// WithFields API - includes tags + fields
	log.InfoWithFields(ctx, "payment validated", map[string]interface{}{
		"amount":   99.99,
		"currency": "USD",
		"user_id":  "user-123",
	})

	// Event API - includes tags + manual fields
	log.InfoEvent(ctx).
		Str("payment_id", "pay-456").
		Float64("amount", 99.99).
		Msg("payment captured")

	// Error with context
	err := errors.New("insufficient funds")
	log.ErrorWithFields(ctx, "payment failed", map[string]interface{}{
		"error":      err,
		"payment_id": "pay-456",
		"balance":    50.00,
	})
}

// Example comparing all three logging APIs
func Example_apiComparison() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()
	ctx = log.WithTags(ctx, "example")

	// 1. Simple API - for basic messages
	log.Info(ctx, "server started on port %d", 8080)

	// 2. WithFields API - for simple structured logging
	log.InfoWithFields(ctx, "user created", map[string]interface{}{
		"user_id": "123",
		"email":   "user@example.com",
	})

	// 3. Event API - for complex structured logging
	log.InfoEvent(ctx).
		Str("user_id", "123").
		Str("email", "user@example.com").
		Int("age", 30).
		Bool("verified", true).
		Dur("registration_time", 150*time.Millisecond).
		Msg("user created with full details")
}

// Example showing GetTags usage
func ExampleGetTags() {
	ctx := context.Background()
	ctx = log.WithTags(ctx, "api", "v2", "production")

	// Retrieve tags from context
	tags := log.GetTags(ctx)

	// tags = []string{"api", "v2", "production"}
	_ = tags

	// Can be used for conditional logic
	if len(tags) > 0 {
		// Tags are present
		log.Info(ctx, "request has tags")
	}
}

// Example showing nil context handling
func ExampleWithTags_nilContext() {
	// WithTags creates a new context if given nil
	ctx := log.WithTags(nil, "background-job")

	// Can use the created context
	log.Info(ctx, "background job started")
}

// Example showing real-world usage in HTTP handler
func Example_httpHandlerUsage() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("api-server"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	// Simulating HTTP handler
	handleRequest := func(ctx context.Context, userID string) {
		// Add request-specific tags
		ctx = log.WithTags(ctx, "http", "api-v1")

		// Log request start
		log.InfoWithFields(ctx, "handling request", map[string]interface{}{
			"user_id": userID,
			"path":    "/api/v1/users",
			"method":  "GET",
		})

		// Do some work...
		// ...

		// Log success
		log.InfoWithFields(ctx, "request completed", map[string]interface{}{
			"user_id":  userID,
			"status":   200,
			"duration": 45 * time.Millisecond,
		})
	}

	ctx := context.Background()
	handleRequest(ctx, "user-123")
}
