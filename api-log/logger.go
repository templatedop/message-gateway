package log

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

const (
	level   = "level"
	message = "message"
	service = "service"
	version = "version"
	stdout  = "stdout"
	noop    = "noop"
	test    = "test"
	console = "console"
)

var (
	baseLogger *Logger
)

type Logger struct {
	logger *zerolog.Logger
}

// ToZerolog exposes the internal zerolog logger.
// Deprecated: Use the Event-based API instead (DebugEvent, InfoEvent, etc.)
// for structured logging with fields. This method will be removed in a future version.
//
// Instead of:
//   log.GetBaseLoggerInstance().ToZerolog().Info().Str("key", "val").Msg("message")
// Use:
//   log.InfoEvent(ctx).Str("key", "val").Msg("message")
func (l *Logger) ToZerolog() *zerolog.Logger {
	return l.logger
}

// FromZerolog converts a zerolog logger to a logger.
// Deprecated: This is mainly for internal use. Use the standard API instead.
func FromZerolog(logger zerolog.Logger) *Logger {
	return &Logger{&logger}
}

// Debug -.
func Debug(ctx context.Context, message interface{}, args ...interface{}) {
	if ctx != nil {
		getCtxLogger(ctx).msg(zerolog.DebugLevel, message, args...)
	} else if baseLogger != nil {
		baseLogger.msg(zerolog.DebugLevel, message, args...)
	} else {
		getDefaultLogger().msg(zerolog.DebugLevel, message, args...)
	}
}

// Info -.
func Info(ctx context.Context, message interface{}, args ...interface{}) {
	// l.msg("info", message, args...)
	if ctx != nil {
		getCtxLogger(ctx).msg(zerolog.InfoLevel, message, args...)
	} else if baseLogger != nil {
		baseLogger.msg(zerolog.InfoLevel, message, args...)
	} else {
		getDefaultLogger().msg(zerolog.InfoLevel, message, args...)
	}
}

// Warn -.
func Warn(ctx context.Context, message interface{}, args ...interface{}) {
	if ctx != nil {
		getCtxLogger(ctx).msg(zerolog.WarnLevel, message, args...)
	} else if baseLogger != nil {
		baseLogger.msg(zerolog.WarnLevel, message, args...)
	} else {
		getDefaultLogger().msg(zerolog.WarnLevel, message, args...)
	}
}

// Error -.
func Error(ctx context.Context, message interface{}, args ...interface{}) {
	/*if l.logger.GetLevel() <= zerolog.DebugLevel {
		l.Debug(message, args...)
	}*/
	if ctx != nil {
		getCtxLogger(ctx).msg(zerolog.ErrorLevel, message, args...)
	} else if baseLogger != nil {
		baseLogger.msg(zerolog.ErrorLevel, message, args...)
	} else {
		getDefaultLogger().msg(zerolog.ErrorLevel, message, args...)
	}
}

// Critical logs a message at FatalLevel but does not exit the application.
// Use this for critical errors that should be logged at the highest severity
// but where you want to handle cleanup or continue execution.
func Critical(ctx context.Context, message interface{}, args ...interface{}) {
	if ctx != nil {
		getCtxLogger(ctx).msg(zerolog.FatalLevel, message, args...)
	} else if baseLogger != nil {
		baseLogger.msg(zerolog.FatalLevel, message, args...)
	} else {
		getDefaultLogger().msg(zerolog.FatalLevel, message, args...)
	}
}

// Fatal logs a message at FatalLevel and exits the application with status code 1.
// This follows standard logging library conventions where Fatal terminates the program.
// For critical errors that don't require immediate exit, use Critical instead.
func Fatal(ctx context.Context, message interface{}, args ...interface{}) {
	Critical(ctx, message, args...)
	panic("fatal error occurred") // Using panic instead of os.Exit for better testability
}

// DebugEvent returns a zerolog.Event for structured logging at Debug level.
// This allows adding fields before calling Msg() to log the event.
//
// Example:
//   log.DebugEvent(ctx).Str("user_id", "123").Int("count", 10).Msg("processing items")
func DebugEvent(ctx context.Context) *zerolog.Event {
	return getEventLogger(ctx, zerolog.DebugLevel)
}

// InfoEvent returns a zerolog.Event for structured logging at Info level.
// This allows adding fields before calling Msg() to log the event.
//
// Example:
//   log.InfoEvent(ctx).Str("operation", "login").Dur("latency", duration).Msg("user logged in")
func InfoEvent(ctx context.Context) *zerolog.Event {
	return getEventLogger(ctx, zerolog.InfoLevel)
}

// WarnEvent returns a zerolog.Event for structured logging at Warn level.
// This allows adding fields before calling Msg() to log the event.
//
// Example:
//   log.WarnEvent(ctx).Str("reason", "rate_limit").Int("attempts", 5).Msg("rate limit approaching")
func WarnEvent(ctx context.Context) *zerolog.Event {
	return getEventLogger(ctx, zerolog.WarnLevel)
}

// ErrorEvent returns a zerolog.Event for structured logging at Error level.
// This allows adding fields before calling Msg() to log the event.
//
// Example:
//   log.ErrorEvent(ctx).Err(err).Str("query", sql).Msg("database query failed")
func ErrorEvent(ctx context.Context) *zerolog.Event {
	return getEventLogger(ctx, zerolog.ErrorLevel)
}

// CriticalEvent returns a zerolog.Event for structured logging at Fatal level (without exiting).
// This allows adding fields before calling Msg() to log the event.
//
// Example:
//   log.CriticalEvent(ctx).Err(err).Str("service", "payment").Msg("payment service unavailable")
func CriticalEvent(ctx context.Context) *zerolog.Event {
	return getEventLogger(ctx, zerolog.FatalLevel)
}

// getEventLogger returns a zerolog.Event with caller information for the given context and level.
// It handles context-aware logger retrieval and sets up proper caller skip frames.
func getEventLogger(ctx context.Context, level zerolog.Level) *zerolog.Event {
	var logger *Logger
	if ctx != nil {
		logger = getCtxLogger(ctx)
	} else if baseLogger != nil {
		logger = baseLogger
	} else {
		logger = getDefaultLogger()
	}

	// Add caller information with correct skip frame count
	// Skip frames: getEventLogger -> InfoEvent/DebugEvent/etc -> actual caller
	frame := 2
	lw := logger.logger.With().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + frame).Logger()
	return lw.WithLevel(level)
}

func (l *Logger) log(level zerolog.Level, message string, args ...interface{}) {
	frame := 3

	lw := l.logger.With().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + frame).Logger()
	logEvent := lw.WithLevel(level)
	if len(args) == 0 {
		logEvent.Msg(message)
	} else {
		logEvent.Msgf(message, args...)
	}
}

func (l *Logger) msg(level zerolog.Level, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(level, msg.Error(), args...)
	case string:
		l.log(level, msg, args...)
	default:
		l.log(zerolog.InfoLevel, fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}
