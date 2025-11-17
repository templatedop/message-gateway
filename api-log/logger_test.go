package log

import (
	"bytes"
	"context"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func setupTestLogger() *bytes.Buffer {
	// Reset global state
	once = sync.Once{}
	createErr = nil
	baseLogger = nil

	var buf bytes.Buffer
	factory := NewDefaultLoggerFactory()
	factory.Create(
		WithServiceName("test-service"),
		WithLevel(zerolog.DebugLevel),
		WithOutputWriter(&buf),
	)

	return &buf
}

func TestDebug(t *testing.T) {
	buf := setupTestLogger()

	Debug(nil, "debug message")

	if buf.Len() == 0 {
		t.Error("Debug message should be logged")
	}

	output := buf.String()
	if !contains(output, "debug message") {
		t.Errorf("Expected 'debug message' in output, got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	buf := setupTestLogger()

	Info(nil, "info message")

	if buf.Len() == 0 {
		t.Error("Info message should be logged")
	}

	output := buf.String()
	if !contains(output, "info message") {
		t.Errorf("Expected 'info message' in output, got: %s", output)
	}
}

func TestWarn(t *testing.T) {
	buf := setupTestLogger()

	Warn(nil, "warning message")

	if buf.Len() == 0 {
		t.Error("Warn message should be logged")
	}

	output := buf.String()
	if !contains(output, "warning message") {
		t.Errorf("Expected 'warning message' in output, got: %s", output)
	}
}

func TestError(t *testing.T) {
	buf := setupTestLogger()

	Error(nil, "error message")

	if buf.Len() == 0 {
		t.Error("Error message should be logged")
	}

	output := buf.String()
	if !contains(output, "error message") {
		t.Errorf("Expected 'error message' in output, got: %s", output)
	}
}

func TestCritical(t *testing.T) {
	buf := setupTestLogger()

	// Critical should log but not exit/panic
	Critical(nil, "critical message")

	if buf.Len() == 0 {
		t.Error("Critical message should be logged")
	}

	output := buf.String()
	if !contains(output, "critical message") {
		t.Errorf("Expected 'critical message' in output, got: %s", output)
	}
}

func TestFatal_Panics(t *testing.T) {
	buf := setupTestLogger()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Fatal should panic")
		} else {
			// Verify the message was logged before panic
			output := buf.String()
			if !contains(output, "fatal message") {
				t.Errorf("Expected 'fatal message' in output before panic, got: %s", output)
			}
		}
	}()

	Fatal(nil, "fatal message")
}

func TestLogLevels_WithContext(t *testing.T) {
	buf := setupTestLogger()
	ctx := context.Background()

	Debug(ctx, "debug with context")
	Info(ctx, "info with context")
	Warn(ctx, "warn with context")
	Error(ctx, "error with context")
	Critical(ctx, "critical with context")

	output := buf.String()

	if !contains(output, "debug with context") {
		t.Error("Debug with context should be logged")
	}
	if !contains(output, "info with context") {
		t.Error("Info with context should be logged")
	}
	if !contains(output, "warn with context") {
		t.Error("Warn with context should be logged")
	}
	if !contains(output, "error with context") {
		t.Error("Error with context should be logged")
	}
	if !contains(output, "critical with context") {
		t.Error("Critical with context should be logged")
	}
}

func TestLogWithFormatting(t *testing.T) {
	buf := setupTestLogger()

	Info(nil, "user %s logged in", "john")

	output := buf.String()
	if !contains(output, "user john logged in") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

func TestLogWithError(t *testing.T) {
	buf := setupTestLogger()

	testErr := &testError{msg: "test error"}
	Error(nil, testErr)

	output := buf.String()
	if !contains(output, "test error") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestGetCtxLogger_NilContext(t *testing.T) {
	setupTestLogger()

	logger := getCtxLogger(nil)

	if logger == nil {
		t.Error("getCtxLogger should return base logger for nil context")
	}
}

func TestGetCtxLogger_WithGinContext(t *testing.T) {
	setupTestLogger()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	logger := getCtxLogger(c)

	if logger == nil {
		t.Error("getCtxLogger should return a logger for gin.Context")
	}
}

func TestGetCtxLogger_WithStandardContext(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	logger := getCtxLogger(ctx)

	if logger == nil {
		t.Error("getCtxLogger should return base logger for standard context without logger")
	}
}

func TestGetCtxLogger_WithLoggerInContext(t *testing.T) {
	buf := setupTestLogger()

	customLogger := zerolog.New(buf).With().Str("custom", "true").Logger()
	logger := FromZerolog(customLogger)

	ctx := context.WithValue(context.Background(), ctxLoggerKey, logger)

	retrievedLogger := getCtxLogger(ctx)

	if retrievedLogger == nil {
		t.Error("getCtxLogger should return the logger from context")
	}

	if retrievedLogger != logger {
		t.Error("getCtxLogger should return the same logger instance")
	}
}

func TestGetCtxLogger_WithWrongTypeInContext(t *testing.T) {
	setupTestLogger()

	// Put wrong type in context
	ctx := context.WithValue(context.Background(), ctxLoggerKey, "not a logger")

	logger := getCtxLogger(ctx)

	// Should fall back to base logger
	if logger == nil {
		t.Error("getCtxLogger should return base logger when context has wrong type")
	}
}

func TestToZerolog(t *testing.T) {
	setupTestLogger()

	if baseLogger == nil {
		t.Fatal("baseLogger should be initialized")
	}

	zl := baseLogger.ToZerolog()

	if zl == nil {
		t.Error("ToZerolog should return a non-nil zerolog.Logger")
	}
}

func TestFromZerolog(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf)

	logger := FromZerolog(zl)

	if logger == nil {
		t.Error("FromZerolog should return a non-nil Logger")
	}

	if logger.logger == nil {
		t.Error("FromZerolog should set the internal logger")
	}

	// Verify it works by logging via the Event API
	logger.logger.Info().Msg("test")
	if buf.Len() == 0 {
		t.Error("Logger should be able to log messages")
	}
}

func TestLogWithEvent_String(t *testing.T) {
	buf := setupTestLogger()

	// Test logWithEvent with string message
	event := InfoEvent(nil)
	logWithEvent(event, "test message")

	output := buf.String()
	if !contains(output, "test message") {
		t.Errorf("Expected 'test message' in output, got: %s", output)
	}
}

func TestLogWithEvent_Error(t *testing.T) {
	buf := setupTestLogger()

	// Test logWithEvent with error
	testErr := &testError{msg: "error occurred"}
	event := ErrorEvent(nil)
	logWithEvent(event, testErr)

	output := buf.String()
	if !contains(output, "error occurred") {
		t.Errorf("Expected 'error occurred' in output, got: %s", output)
	}
}

func TestLogWithEvent_WithFormatting(t *testing.T) {
	buf := setupTestLogger()

	// Test logWithEvent with format string
	event := InfoEvent(nil)
	logWithEvent(event, "user %s completed %d tasks", "john", 5)

	output := buf.String()
	if !contains(output, "john") || !contains(output, "5") {
		t.Errorf("Expected formatted message in output, got: %s", output)
	}
}

// Tests for Event-based API

func TestDebugEvent(t *testing.T) {
	buf := setupTestLogger()

	DebugEvent(nil).Str("key", "value").Int("count", 42).Msg("debug event test")

	output := buf.String()
	if !contains(output, "debug event test") {
		t.Error("Debug event message should be logged")
	}
	if !contains(output, "key") || !contains(output, "value") {
		t.Error("Debug event should include structured fields")
	}
	if !contains(output, "42") {
		t.Error("Debug event should include integer field")
	}
}

func TestInfoEvent(t *testing.T) {
	buf := setupTestLogger()

	InfoEvent(nil).Str("operation", "login").Str("user", "john").Msg("user logged in")

	output := buf.String()
	if !contains(output, "user logged in") {
		t.Error("Info event message should be logged")
	}
	if !contains(output, "operation") || !contains(output, "login") {
		t.Error("Info event should include operation field")
	}
	if !contains(output, "john") {
		t.Error("Info event should include user field")
	}
}

func TestWarnEvent(t *testing.T) {
	buf := setupTestLogger()

	WarnEvent(nil).Str("reason", "rate_limit").Int("attempts", 5).Msg("approaching rate limit")

	output := buf.String()
	if !contains(output, "approaching rate limit") {
		t.Error("Warn event message should be logged")
	}
	if !contains(output, "rate_limit") {
		t.Error("Warn event should include reason field")
	}
	if !contains(output, "5") {
		t.Error("Warn event should include attempts field")
	}
}

func TestErrorEvent(t *testing.T) {
	buf := setupTestLogger()

	testErr := &testError{msg: "connection timeout"}
	ErrorEvent(nil).Err(testErr).Str("host", "db.example.com").Msg("database connection failed")

	output := buf.String()
	if !contains(output, "database connection failed") {
		t.Error("Error event message should be logged")
	}
	if !contains(output, "connection timeout") {
		t.Error("Error event should include error")
	}
	if !contains(output, "db.example.com") {
		t.Error("Error event should include host field")
	}
}

func TestCriticalEvent(t *testing.T) {
	buf := setupTestLogger()

	CriticalEvent(nil).Str("service", "payment").Bool("available", false).Msg("service unavailable")

	output := buf.String()
	if !contains(output, "service unavailable") {
		t.Error("Critical event message should be logged")
	}
	if !contains(output, "payment") {
		t.Error("Critical event should include service field")
	}
	if !contains(output, "false") {
		t.Error("Critical event should include available field")
	}
}

func TestDebugEvent_WithContext(t *testing.T) {
	buf := setupTestLogger()
	ctx := context.Background()

	DebugEvent(ctx).Str("context_key", "context_value").Msg("debug with context")

	output := buf.String()
	if !contains(output, "debug with context") {
		t.Error("Debug event with context should be logged")
	}
	if !contains(output, "context_value") {
		t.Error("Debug event should include fields when using context")
	}
}

func TestInfoEvent_WithContext(t *testing.T) {
	buf := setupTestLogger()
	ctx := context.Background()

	InfoEvent(ctx).Int("status", 200).Msg("request processed")

	output := buf.String()
	if !contains(output, "request processed") {
		t.Error("Info event with context should be logged")
	}
	if !contains(output, "200") {
		t.Error("Info event should include status field")
	}
}

func TestErrorEvent_WithContext(t *testing.T) {
	buf := setupTestLogger()
	ctx := context.Background()

	testErr := &testError{msg: "validation failed"}
	ErrorEvent(ctx).Err(testErr).Str("field", "email").Msg("input validation error")

	output := buf.String()
	if !contains(output, "input validation error") {
		t.Error("Error event with context should be logged")
	}
	if !contains(output, "validation failed") {
		t.Error("Error event should include error details")
	}
	if !contains(output, "email") {
		t.Error("Error event should include field information")
	}
}

func TestEventAPI_MultipleFields(t *testing.T) {
	buf := setupTestLogger()

	InfoEvent(nil).
		Str("user_id", "user123").
		Str("action", "purchase").
		Float64("amount", 99.99).
		Int("quantity", 3).
		Bool("success", true).
		Msg("transaction completed")

	output := buf.String()
	if !contains(output, "transaction completed") {
		t.Error("Message should be logged")
	}
	if !contains(output, "user123") {
		t.Error("Should include user_id field")
	}
	if !contains(output, "purchase") {
		t.Error("Should include action field")
	}
	if !contains(output, "99.99") {
		t.Error("Should include amount field")
	}
	if !contains(output, "3") {
		t.Error("Should include quantity field")
	}
}

func TestEventAPI_EmptyMessage(t *testing.T) {
	buf := setupTestLogger()

	InfoEvent(nil).Str("key", "value").Msg("")

	output := buf.String()
	if !contains(output, "key") || !contains(output, "value") {
		t.Error("Should log structured fields even with empty message")
	}
}

func TestEventAPI_WithGinContext(t *testing.T) {
	setupTestLogger()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Should not panic and should work with gin.Context
	InfoEvent(c).Str("endpoint", "/test").Msg("handling request")
}

// Tests for WithFields API

func TestDebugWithFields(t *testing.T) {
	buf := setupTestLogger()

	DebugWithFields(nil, "processing user", map[string]interface{}{
		"user_id": "123",
		"action":  "login",
		"count":   42,
	})

	output := buf.String()
	if !contains(output, "processing user") {
		t.Error("Message should be logged")
	}
	if !contains(output, "123") {
		t.Error("user_id field should be logged")
	}
	if !contains(output, "login") {
		t.Error("action field should be logged")
	}
	if !contains(output, "42") {
		t.Error("count field should be logged")
	}
}

func TestInfoWithFields(t *testing.T) {
	buf := setupTestLogger()

	InfoWithFields(nil, "user logged in", map[string]interface{}{
		"user_id": "456",
		"ip":      "192.168.1.1",
		"success": true,
	})

	output := buf.String()
	if !contains(output, "user logged in") {
		t.Error("Message should be logged")
	}
	if !contains(output, "456") {
		t.Error("user_id field should be logged")
	}
	if !contains(output, "192.168.1.1") {
		t.Error("ip field should be logged")
	}
	if !contains(output, "true") {
		t.Error("success field should be logged")
	}
}

func TestWarnWithFields(t *testing.T) {
	buf := setupTestLogger()

	WarnWithFields(nil, "rate limit approaching", map[string]interface{}{
		"attempts": 4,
		"limit":    5,
		"rate":     0.8,
	})

	output := buf.String()
	if !contains(output, "rate limit approaching") {
		t.Error("Message should be logged")
	}
	if !contains(output, "4") {
		t.Error("attempts field should be logged")
	}
	if !contains(output, "5") {
		t.Error("limit field should be logged")
	}
}

func TestErrorWithFields(t *testing.T) {
	buf := setupTestLogger()

	testErr := &testError{msg: "connection timeout"}
	ErrorWithFields(nil, "database query failed", map[string]interface{}{
		"error":    testErr,
		"query":    "SELECT * FROM users",
		"duration": 5000,
	})

	output := buf.String()
	if !contains(output, "database query failed") {
		t.Error("Message should be logged")
	}
	if !contains(output, "connection timeout") {
		t.Error("error field should be logged")
	}
	if !contains(output, "SELECT") {
		t.Error("query field should be logged")
	}
}

func TestCriticalWithFields(t *testing.T) {
	buf := setupTestLogger()

	CriticalWithFields(nil, "service unavailable", map[string]interface{}{
		"service":   "payment-gateway",
		"available": false,
		"attempts":  3,
	})

	output := buf.String()
	if !contains(output, "service unavailable") {
		t.Error("Message should be logged")
	}
	if !contains(output, "payment-gateway") {
		t.Error("service field should be logged")
	}
	if !contains(output, "false") {
		t.Error("available field should be logged")
	}
}

func TestWithFields_MultipleTypes(t *testing.T) {
	buf := setupTestLogger()

	InfoWithFields(nil, "complex data", map[string]interface{}{
		"string_field":  "value",
		"int_field":     42,
		"int64_field":   int64(9876543210),
		"float_field":   3.14159,
		"bool_field":    true,
		"strings_field": []string{"a", "b", "c"},
	})

	output := buf.String()
	if !contains(output, "complex data") {
		t.Error("Message should be logged")
	}
	if !contains(output, "value") {
		t.Error("string field should be logged")
	}
	if !contains(output, "42") {
		t.Error("int field should be logged")
	}
	if !contains(output, "3.14159") {
		t.Error("float field should be logged")
	}
}

// Tests for Tags support

func TestWithTags(t *testing.T) {
	buf := setupTestLogger()

	ctx := WithTags(context.Background(), "database", "payment")

	InfoEvent(ctx).Msg("processing transaction")

	output := buf.String()
	if !contains(output, "database") {
		t.Error("database tag should be logged")
	}
	if !contains(output, "payment") {
		t.Error("payment tag should be logged")
	}
}

func TestWithTags_Multiple(t *testing.T) {
	buf := setupTestLogger()

	ctx := WithTags(context.Background(), "tag1", "tag2")
	ctx = WithTags(ctx, "tag3")

	InfoEvent(ctx).Msg("test message")

	output := buf.String()
	if !contains(output, "tag1") {
		t.Error("tag1 should be logged")
	}
	if !contains(output, "tag2") {
		t.Error("tag2 should be logged")
	}
	if !contains(output, "tag3") {
		t.Error("tag3 should be logged")
	}
}

func TestWithTags_SimpleAPI(t *testing.T) {
	buf := setupTestLogger()

	ctx := WithTags(context.Background(), "api", "auth")

	Info(ctx, "user authenticated")

	output := buf.String()
	if !contains(output, "api") {
		t.Error("api tag should be logged")
	}
	if !contains(output, "auth") {
		t.Error("auth tag should be logged")
	}
}

func TestWithTags_WithFields(t *testing.T) {
	buf := setupTestLogger()

	ctx := WithTags(context.Background(), "database")

	InfoWithFields(ctx, "query executed", map[string]interface{}{
		"table": "users",
		"rows":  10,
	})

	output := buf.String()
	if !contains(output, "database") {
		t.Error("database tag should be logged")
	}
	if !contains(output, "users") {
		t.Error("table field should be logged")
	}
	if !contains(output, "10") {
		t.Error("rows field should be logged")
	}
}

func TestGetTags_NilContext(t *testing.T) {
	tags := GetTags(nil)

	if tags != nil {
		t.Error("GetTags should return nil for nil context")
	}
}

func TestGetTags_NoTags(t *testing.T) {
	ctx := context.Background()
	tags := GetTags(ctx)

	if tags != nil {
		t.Error("GetTags should return nil when no tags are present")
	}
}

func TestGetTags_WithTags(t *testing.T) {
	ctx := WithTags(context.Background(), "tag1", "tag2")
	tags := GetTags(ctx)

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "tag1" || tags[1] != "tag2" {
		t.Errorf("Expected [tag1, tag2], got %v", tags)
	}
}

func TestWithTags_NilContext(t *testing.T) {
	ctx := WithTags(nil, "tag1")

	if ctx == nil {
		t.Error("WithTags should create a context when given nil")
	}

	tags := GetTags(ctx)
	if len(tags) != 1 || tags[0] != "tag1" {
		t.Error("Tag should be added to new context")
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
