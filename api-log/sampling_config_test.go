package log

import (
	"bytes"
	"context"
	"math/rand"
	"strings"
	"sync"
	"testing"

	"github.com/rs/zerolog"
)

func TestDefaultSamplingConfig(t *testing.T) {
	config := DefaultSamplingConfig()

	if config == nil {
		t.Fatal("DefaultSamplingConfig returned nil")
	}

	if config.GlobalRate != 1.0 {
		t.Errorf("Expected GlobalRate 1.0, got %f", config.GlobalRate)
	}

	if config.LevelRates == nil {
		t.Error("LevelRates should not be nil")
	}

	if config.TagRates == nil {
		t.Error("TagRates should not be nil")
	}

	if config.DisabledLevels == nil {
		t.Error("DisabledLevels should not be nil")
	}

	if config.Rand == nil {
		t.Error("Rand should not be nil")
	}
}

func TestSamplingConfig_ShouldLog_NilConfig(t *testing.T) {
	var config *SamplingConfig = nil

	// Nil config should always return true
	if !config.ShouldLog(zerolog.InfoLevel, []string{}) {
		t.Error("Nil config should always allow logging")
	}

	if !config.ShouldLog(zerolog.DebugLevel, []string{"tag"}) {
		t.Error("Nil config should always allow logging")
	}
}

func TestSamplingConfig_ShouldLog_DisabledLevels(t *testing.T) {
	config := &SamplingConfig{
		GlobalRate: 1.0,
		DisabledLevels: []zerolog.Level{
			zerolog.DebugLevel,
			zerolog.TraceLevel,
		},
		Rand: rand.New(rand.NewSource(12345)),
	}

	// Debug should be disabled
	if config.ShouldLog(zerolog.DebugLevel, []string{}) {
		t.Error("Debug level should be disabled")
	}

	// Trace should be disabled
	if config.ShouldLog(zerolog.TraceLevel, []string{}) {
		t.Error("Trace level should be disabled")
	}

	// Info should be allowed
	if !config.ShouldLog(zerolog.InfoLevel, []string{}) {
		t.Error("Info level should be allowed")
	}

	// Error should be allowed
	if !config.ShouldLog(zerolog.ErrorLevel, []string{}) {
		t.Error("Error level should be allowed")
	}
}

func TestSamplingConfig_ShouldLog_GlobalRate(t *testing.T) {
	// Test with 0% rate (log nothing)
	config := &SamplingConfig{
		GlobalRate: 0.0,
		Rand:       rand.New(rand.NewSource(12345)),
	}

	allowed := 0
	for i := 0; i < 100; i++ {
		if config.ShouldLog(zerolog.InfoLevel, []string{}) {
			allowed++
		}
	}

	if allowed != 0 {
		t.Errorf("With GlobalRate 0.0, expected 0 logs allowed, got %d", allowed)
	}

	// Test with 100% rate (log everything)
	config = &SamplingConfig{
		GlobalRate: 1.0,
		Rand:       rand.New(rand.NewSource(12345)),
	}

	allowed = 0
	for i := 0; i < 100; i++ {
		if config.ShouldLog(zerolog.InfoLevel, []string{}) {
			allowed++
		}
	}

	if allowed != 100 {
		t.Errorf("With GlobalRate 1.0, expected 100 logs allowed, got %d", allowed)
	}

	// Test with 50% rate (approximately half)
	config = &SamplingConfig{
		GlobalRate: 0.5,
		Rand:       rand.New(rand.NewSource(12345)),
	}

	allowed = 0
	for i := 0; i < 1000; i++ {
		if config.ShouldLog(zerolog.InfoLevel, []string{}) {
			allowed++
		}
	}

	// Should be approximately 500, allow ±100 for randomness
	if allowed < 400 || allowed > 600 {
		t.Errorf("With GlobalRate 0.5, expected ~500 logs allowed, got %d", allowed)
	}
}

func TestSamplingConfig_ShouldLog_LevelRates(t *testing.T) {
	config := &SamplingConfig{
		GlobalRate: 1.0, // No global sampling
		LevelRates: map[zerolog.Level]float64{
			zerolog.DebugLevel: 0.1, // 10% of debug logs
			zerolog.InfoLevel:  1.0, // 100% of info logs
		},
		Rand: rand.New(rand.NewSource(12345)),
	}

	// Test debug sampling (10%)
	debugAllowed := 0
	for i := 0; i < 1000; i++ {
		if config.ShouldLog(zerolog.DebugLevel, []string{}) {
			debugAllowed++
		}
	}

	// Should be approximately 100, allow ±50 for randomness
	if debugAllowed < 50 || debugAllowed > 150 {
		t.Errorf("With DebugLevel rate 0.1, expected ~100 logs allowed, got %d", debugAllowed)
	}

	// Test info logging (100%)
	infoAllowed := 0
	for i := 0; i < 100; i++ {
		if config.ShouldLog(zerolog.InfoLevel, []string{}) {
			infoAllowed++
		}
	}

	if infoAllowed != 100 {
		t.Errorf("With InfoLevel rate 1.0, expected 100 logs allowed, got %d", infoAllowed)
	}

	// Test level without specific rate (should use global rate of 1.0)
	errorAllowed := 0
	for i := 0; i < 100; i++ {
		if config.ShouldLog(zerolog.ErrorLevel, []string{}) {
			errorAllowed++
		}
	}

	if errorAllowed != 100 {
		t.Errorf("Error level without specific rate should use global rate, expected 100, got %d", errorAllowed)
	}
}

func TestSamplingConfig_ShouldLog_TagRates(t *testing.T) {
	config := &SamplingConfig{
		GlobalRate: 1.0, // No global sampling
		TagRates: map[string]float64{
			"database": 0.5, // 50% of database logs
			"cache":    0.1, // 10% of cache logs
		},
		Rand: rand.New(rand.NewSource(12345)),
	}

	// Test database tag sampling (50%)
	dbAllowed := 0
	for i := 0; i < 1000; i++ {
		if config.ShouldLog(zerolog.InfoLevel, []string{"database"}) {
			dbAllowed++
		}
	}

	// Should be approximately 500, allow ±100 for randomness
	if dbAllowed < 400 || dbAllowed > 600 {
		t.Errorf("With database tag rate 0.5, expected ~500 logs allowed, got %d", dbAllowed)
	}

	// Test cache tag sampling (10%)
	cacheAllowed := 0
	for i := 0; i < 1000; i++ {
		if config.ShouldLog(zerolog.InfoLevel, []string{"cache"}) {
			cacheAllowed++
		}
	}

	// Should be approximately 100, allow ±50 for randomness
	if cacheAllowed < 50 || cacheAllowed > 150 {
		t.Errorf("With cache tag rate 0.1, expected ~100 logs allowed, got %d", cacheAllowed)
	}

	// Test tag without specific rate (should use global rate of 1.0)
	apiAllowed := 0
	for i := 0; i < 100; i++ {
		if config.ShouldLog(zerolog.InfoLevel, []string{"api"}) {
			apiAllowed++
		}
	}

	if apiAllowed != 100 {
		t.Errorf("Tag without specific rate should use global rate, expected 100, got %d", apiAllowed)
	}
}

func TestSamplingConfig_ShouldLog_CombinedRules(t *testing.T) {
	config := &SamplingConfig{
		GlobalRate: 0.8, // 80% global
		LevelRates: map[zerolog.Level]float64{
			zerolog.DebugLevel: 0.5, // 50% of debug
		},
		TagRates: map[string]float64{
			"expensive": 0.2, // 20% of expensive operations
		},
		DisabledLevels: []zerolog.Level{zerolog.TraceLevel},
		Rand:           rand.New(rand.NewSource(12345)),
	}

	// Trace should be completely disabled
	if config.ShouldLog(zerolog.TraceLevel, []string{}) {
		t.Error("Trace level should be disabled")
	}

	// Debug with expensive tag: 80% global * 50% level * 20% tag ≈ 8%
	expensiveDebugAllowed := 0
	for i := 0; i < 1000; i++ {
		if config.ShouldLog(zerolog.DebugLevel, []string{"expensive"}) {
			expensiveDebugAllowed++
		}
	}

	// Very rough check - should be low
	if expensiveDebugAllowed > 200 {
		t.Errorf("Debug with expensive tag should be heavily sampled, got %d/1000", expensiveDebugAllowed)
	}
}

func TestSamplingConfig_Integration(t *testing.T) {
	// Reset global state
	once = sync.Once{}
	createErr = nil
	baseLogger = nil
	samplingConfig = nil

	var buf bytes.Buffer
	factory := NewDefaultLoggerFactory()

	// Create logger with sampling config
	config := &SamplingConfig{
		GlobalRate: 1.0,
		LevelRates: map[zerolog.Level]float64{
			zerolog.DebugLevel: 0.0, // Sample 0% of debug (disable)
			zerolog.InfoLevel:  1.0, // Sample 100% of info
		},
		Rand: rand.New(rand.NewSource(12345)),
	}

	factory.Create(
		WithServiceName("test-sampling"),
		WithLevel(zerolog.InfoLevel), // Use Info level to avoid internal debug logs
		WithOutputWriter(&buf),
		WithSampling(config),
	)

	ctx := context.Background()

	// Debug logs should be sampled out (0%)
	buf.Reset()
	Debug(ctx, "debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug logs should be sampled out")
	}

	// Info logs should go through (100%)
	buf.Reset()
	Info(ctx, "info message")
	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Error("Info logs should not be sampled")
	}

	// Error logs should go through (no specific rate, uses global 1.0)
	buf.Reset()
	Error(ctx, "error message")
	output = buf.String()
	if !strings.Contains(output, "error message") {
		t.Error("Error logs should not be sampled")
	}
}

func TestSamplingConfig_WithTags_Integration(t *testing.T) {
	// Reset global state
	once = sync.Once{}
	createErr = nil
	baseLogger = nil
	samplingConfig = nil

	var buf bytes.Buffer
	factory := NewDefaultLoggerFactory()

	// Create logger with tag-based sampling
	config := &SamplingConfig{
		GlobalRate: 1.0,
		TagRates: map[string]float64{
			"skip": 0.0, // Sample 0% of logs with "skip" tag
		},
		Rand: rand.New(rand.NewSource(12345)),
	}

	factory.Create(
		WithServiceName("test-tag-sampling"),
		WithLevel(zerolog.InfoLevel),
		WithOutputWriter(&buf),
		WithSampling(config),
	)

	ctx := context.Background()

	// Logs without tag should go through
	buf.Reset()
	Info(ctx, "normal message")
	if !strings.Contains(buf.String(), "normal message") {
		t.Error("Normal logs should not be sampled")
	}

	// Logs with "skip" tag should be sampled out
	buf.Reset()
	ctxWithTag := WithTags(ctx, "skip")
	Info(ctxWithTag, "skipped message")
	if buf.Len() > 0 {
		t.Error("Logs with 'skip' tag should be sampled out")
	}

	// Logs with different tag should go through
	buf.Reset()
	ctxWithOtherTag := WithTags(ctx, "important")
	Info(ctxWithOtherTag, "important message")
	if !strings.Contains(buf.String(), "important message") {
		t.Error("Logs with 'important' tag should not be sampled")
	}
}

func TestSamplingConfig_EventAPI_Integration(t *testing.T) {
	// Reset global state
	once = sync.Once{}
	createErr = nil
	baseLogger = nil
	samplingConfig = nil

	var buf bytes.Buffer
	factory := NewDefaultLoggerFactory()

	config := &SamplingConfig{
		GlobalRate: 1.0,
		LevelRates: map[zerolog.Level]float64{
			zerolog.WarnLevel: 0.0, // Disable warnings
		},
		Rand: rand.New(rand.NewSource(12345)),
	}

	factory.Create(
		WithServiceName("test-event-sampling"),
		WithLevel(zerolog.InfoLevel), // Use Info level to avoid internal debug logs
		WithOutputWriter(&buf),
		WithSampling(config),
	)

	ctx := context.Background()

	// Warn event should be sampled out
	buf.Reset()
	WarnEvent(ctx).Str("key", "value").Msg("warning message")
	if strings.Contains(buf.String(), "warning message") {
		t.Error("Warn events should be sampled out")
	}

	// Info event should go through
	buf.Reset()
	InfoEvent(ctx).Str("key", "value").Msg("info message")
	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Error("Info events should not be sampled")
	}
	if !strings.Contains(output, "\"key\":\"value\"") {
		t.Error("Event fields should be preserved")
	}
}

func TestSamplingConfig_WithFields_Integration(t *testing.T) {
	// Reset global state
	once = sync.Once{}
	createErr = nil
	baseLogger = nil
	samplingConfig = nil

	var buf bytes.Buffer
	factory := NewDefaultLoggerFactory()

	config := &SamplingConfig{
		GlobalRate:     1.0, // Allow all logs by default
		DisabledLevels: []zerolog.Level{zerolog.DebugLevel},
		Rand:           rand.New(rand.NewSource(12345)),
	}

	factory.Create(
		WithServiceName("test-fields-sampling"),
		WithLevel(zerolog.InfoLevel), // Use Info level to avoid internal debug logs
		WithOutputWriter(&buf),
		WithSampling(config),
	)

	ctx := context.Background()

	// Debug with fields should be sampled out
	buf.Reset()
	DebugWithFields(ctx, "debug message", map[string]interface{}{
		"key": "value",
	})
	if strings.Contains(buf.String(), "debug message") {
		t.Error("DebugWithFields should be sampled out")
	}

	// Info with fields should go through
	buf.Reset()
	InfoWithFields(ctx, "info message", map[string]interface{}{
		"key": "value",
	})
	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Error("InfoWithFields should not be sampled")
	}
	if !strings.Contains(output, "\"key\":\"value\"") {
		t.Error("Fields should be preserved")
	}
}
