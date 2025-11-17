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

	// Verify it works by logging a message
	logger.msg(zerolog.InfoLevel, "test")
	if buf.Len() == 0 {
		t.Error("Logger should be able to log messages")
	}
}

func TestLogger_Msg_WithString(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf)
	logger := &Logger{logger: &zl}

	logger.msg(zerolog.InfoLevel, "test message")

	output := buf.String()
	if !contains(output, "test message") {
		t.Errorf("Expected 'test message' in output, got: %s", output)
	}
}

func TestLogger_Msg_WithError(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf)
	logger := &Logger{logger: &zl}

	testErr := &testError{msg: "error occurred"}
	logger.msg(zerolog.ErrorLevel, testErr)

	output := buf.String()
	if !contains(output, "error occurred") {
		t.Errorf("Expected 'error occurred' in output, got: %s", output)
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
