package log

import (
	"bytes"
	"sync"
	"testing"

	"github.com/rs/zerolog"
)

func TestDefaultLoggerFactory_Create_ThreadSafety(t *testing.T) {
	// Reset global state for test
	once = sync.Once{}
	createErr = nil
	baseLogger = nil

	factory := NewDefaultLoggerFactory()

	// Create a buffer to capture log output
	var buf bytes.Buffer

	// First call should succeed
	err := factory.Create(
		WithServiceName("test-service"),
		WithLevel(zerolog.InfoLevel),
		WithOutputWriter(&buf),
	)

	if err != nil {
		t.Errorf("First Create() call should succeed, got error: %v", err)
	}

	if baseLogger == nil {
		t.Error("baseLogger should be initialized after Create()")
	}

	// Subsequent calls should return the same error state (nil in this case)
	err2 := factory.Create(
		WithServiceName("different-service"), // Different options should be ignored
	)

	if err2 != nil {
		t.Errorf("Second Create() call should return nil (same as first), got: %v", err2)
	}

	// Verify logger was only initialized once (service name should still be "test-service")
	if baseLogger.logger.GetLevel() != zerolog.InfoLevel {
		t.Error("Logger should retain original configuration from first Create() call")
	}
}

func TestDefaultLoggerFactory_Create_Concurrent(t *testing.T) {
	// Reset global state for test
	once = sync.Once{}
	createErr = nil
	baseLogger = nil

	factory := NewDefaultLoggerFactory()

	// Launch multiple goroutines calling Create() concurrently
	var wg sync.WaitGroup
	numGoroutines := 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := factory.Create(
				WithServiceName("concurrent-test"),
				WithLevel(zerolog.DebugLevel),
			)
			errors <- err
		}()
	}

	wg.Wait()
	close(errors)

	// All calls should return nil (no error)
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent Create() call failed: %v", err)
		}
	}

	// Logger should be initialized
	if baseLogger == nil {
		t.Error("baseLogger should be initialized after concurrent Create() calls")
	}
}

func TestDefaultLoggerFactory_Create_WithOptions(t *testing.T) {
	// Reset global state for test
	once = sync.Once{}
	createErr = nil
	baseLogger = nil

	factory := NewDefaultLoggerFactory()

	var buf bytes.Buffer

	err := factory.Create(
		WithServiceName("custom-service"),
		WithLevel(zerolog.WarnLevel),
		WithOutputWriter(&buf),
		WithVersion("v1.2.3"),
	)

	if err != nil {
		t.Errorf("Create() with options failed: %v", err)
	}

	if baseLogger == nil {
		t.Fatal("baseLogger should be initialized")
	}

	// Verify options were applied
	if baseLogger.logger.GetLevel() != zerolog.WarnLevel {
		t.Errorf("Expected log level WarnLevel, got: %v", baseLogger.logger.GetLevel())
	}
}

func TestGetBaseLoggerInstance_WhenNotInitialized(t *testing.T) {
	// Reset global state
	once = sync.Once{}
	createErr = nil
	baseLogger = nil

	// Should return a default logger, not nil
	logger := GetBaseLoggerInstance()

	if logger == nil {
		t.Error("GetBaseLoggerInstance() should return a default logger, not nil")
	}

	if logger.logger == nil {
		t.Error("Returned logger should have a valid zerolog instance")
	}
}

func TestGetBaseLoggerInstance_WhenInitialized(t *testing.T) {
	// Reset and initialize
	once = sync.Once{}
	createErr = nil
	baseLogger = nil

	factory := NewDefaultLoggerFactory()
	factory.Create(WithServiceName("initialized-test"))

	logger := GetBaseLoggerInstance()

	if logger == nil {
		t.Error("GetBaseLoggerInstance() should return the initialized logger")
	}

	if logger != baseLogger {
		t.Error("GetBaseLoggerInstance() should return the same instance as baseLogger")
	}
}
