package log_test

import (
	"bytes"
	"context"
	"errors"
	"time"

	log "MgApplication/api-log"

	"github.com/rs/zerolog"
)

// Example demonstrating the new Event-based API for structured logging
func ExampleInfoEvent() {
	// Setup logger for this example
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()

	// Old way (deprecated):
	// log.GetBaseLoggerInstance().ToZerolog().Info().Str("user", "john").Msg("user logged in")

	// New way (recommended):
	log.InfoEvent(ctx).
		Str("user", "john").
		Str("action", "login").
		Msg("user logged in")

	// Simple logging (still supported)
	log.Info(ctx, "processing request")
}

// Example showing error logging with structured fields
func ExampleErrorEvent() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.DebugLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()
	err := errors.New("connection timeout")

	// Old way:
	// log.GetBaseLoggerInstance().ToZerolog().Error().Err(err).Str("host", "db.example.com").Msg("connection failed")

	// New way:
	log.ErrorEvent(ctx).
		Err(err).
		Str("host", "db.example.com").
		Int("port", 5432).
		Dur("timeout", 30*time.Second).
		Msg("database connection failed")
}

// Example showing multiple field types
func ExampleDebugEvent_multipleFields() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.DebugLevel),
		log.WithOutputWriter(&buf),
	)

	log.DebugEvent(nil).
		Str("operation", "cache_lookup").
		Str("key", "user:123").
		Bool("hit", true).
		Dur("latency", 2*time.Millisecond).
		Int("size", 1024).
		Msg("cache operation completed")
}

// Example showing migration from old pattern
func Example_migration() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("migration-example"),
		log.WithLevel(zerolog.InfoLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()

	// === BEFORE (Deprecated) ===
	// log.GetBaseLoggerInstance().ToZerolog().Info().Str("module", "auth").Msg("starting module")

	// === AFTER (Recommended) ===
	log.InfoEvent(ctx).Str("module", "auth").Msg("starting module")

	// === Simple logging (unchanged) ===
	log.Info(ctx, "server started on port 8080")
}

// Example showing critical event logging
func ExampleCriticalEvent() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.DebugLevel),
		log.WithOutputWriter(&buf),
	)

	ctx := context.Background()

	log.CriticalEvent(ctx).
		Str("service", "payment-gateway").
		Bool("available", false).
		Int("retry_attempts", 5).
		Msg("critical service unavailable")
}

// Example showing warning event with structured data
func ExampleWarnEvent() {
	var buf bytes.Buffer
	factory := log.NewDefaultLoggerFactory()
	factory.Create(
		log.WithServiceName("example-service"),
		log.WithLevel(zerolog.DebugLevel),
		log.WithOutputWriter(&buf),
	)

	log.WarnEvent(nil).
		Str("ip", "192.168.1.100").
		Int("attempts", 4).
		Int("limit", 5).
		Float64("rate", 0.8).
		Msg("rate limit threshold approaching")
}
