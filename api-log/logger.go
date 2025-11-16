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

// Converts the logger to a zerolog logger
func (l *Logger) ToZerolog() *zerolog.Logger {
	return l.logger
}

// converts a zerolog logger to a logger
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

// Fatal -.
func Fatal(ctx context.Context, message interface{}, args ...interface{}) {
	if ctx != nil {
		getCtxLogger(ctx).msg(zerolog.FatalLevel, message, args...)
	} else if baseLogger != nil {
		baseLogger.msg(zerolog.FatalLevel, message, args...)
	} else {
		getDefaultLogger().msg(zerolog.FatalLevel, message, args...)
	}

	// os.Exit(1)
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
