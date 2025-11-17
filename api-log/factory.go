package log

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type (
	key string
)

const (
	ctxLoggerKey        key = "logger"
	requestIDContextKey key = "request-id"
	userIDKey           key = "user-id"
	methodKey           key = "method"
	pathKey             key = "path"
	paramsKey           key = "params"
	traceIDKey          key = "trace_id"
	officeIDKey         key = "office-id"
)

var (
	once      sync.Once
	createErr error
)

type LoggerFactory interface {
	Create(options ...loggerOption) error
}

type DefaultLoggerFactory struct{}

func NewDefaultLoggerFactory() LoggerFactory {
	return &DefaultLoggerFactory{}
}

// Create creates a new logger instance with the provided options.
// This function is thread-safe and will only initialize the logger once.
// Subsequent calls will return the error from the first initialization attempt (if any).
func (f *DefaultLoggerFactory) Create(options ...loggerOption) error {
	once.Do(func() {
		appliedOpts := defaultLoggerOptions()
		for _, applyOpt := range options {
			applyOpt(&appliedOpts)
		}

		logger := zerolog.
			New(appliedOpts.OutputWriter).
			With().
			Timestamp().
			Str(service, appliedOpts.ServiceName).
			Str(version, appliedOpts.Version).
			Logger().
			Level(appliedOpts.Level)

		zerolog.DefaultContextLogger = &logger
		baseLogger = &Logger{
			logger: &logger,
		}

		// Set global sampling config if provided
		samplingConfig = appliedOpts.SamplingConfig

		// createErr remains nil if initialization succeeds
		// In future, if we add validation or other operations that can fail,
		// we can set createErr here
	})

	return createErr
}

// SetCtxLoggerMiddleware creates a sublogger with request metadata and embed this inside ginCtx
func SetCtxLoggerMiddleware(c *gin.Context) {
	if baseLogger != nil {
		ctxLogger := baseLogger.setRequestMetadata(c)

		// set the above child logger in the context so others can use it.
		ctx := context.WithValue(c.Request.Context(), ctxLoggerKey, &Logger{logger: &ctxLogger})
		c.Request = c.Request.WithContext(ctx)
	}
	c.Next()
}

// getCtxLogger returns the logger embedded in the context.
// If no logger is found, it returns the base logger instance.
// This function handles both gin.Context and context.Context types.
func getCtxLogger(ctx context.Context) *Logger {
	var value any
	if ctx == nil {
		baseLogger := GetBaseLoggerInstance()
		// This is expected in some cases (e.g., startup), so we log at debug level
		if baseLogger != nil && baseLogger.logger != nil && baseLogger.logger.GetLevel() <= zerolog.DebugLevel {
			baseLogger.logger.Debug().Msg("Context is nil, returning base logger")
		}
		return baseLogger
	} else if ginCtx, ok := ctx.(*gin.Context); ok {
		// ctx is of type *gin.Context
		if value = ginCtx.Request.Context().Value(ctxLoggerKey); value == nil {
			baseLogger := GetBaseLoggerInstance()
			// Logger not in context is common before middleware runs
			if baseLogger != nil && baseLogger.logger != nil && baseLogger.logger.GetLevel() <= zerolog.DebugLevel {
				baseLogger.logger.Debug().Msg("Logger not found in gin.Context, returning base logger")
			}
			return baseLogger
		}
	} else {
		// ctx is of type context.Context
		if value = ctx.Value(ctxLoggerKey); value == nil {
			baseLogger := GetBaseLoggerInstance()
			// Logger not in context is common in non-HTTP contexts
			if baseLogger != nil && baseLogger.logger != nil && baseLogger.logger.GetLevel() <= zerolog.DebugLevel {
				baseLogger.logger.Debug().Msg("Logger not found in context.Context, returning base logger")
			}
			return baseLogger
		}
	}
	// check if the value is of type *Logger
	contextLogger, ok := value.(*Logger)
	if !ok {
		baseLogger := GetBaseLoggerInstance()
		// This is unexpected and should be warned about
		if baseLogger != nil && baseLogger.logger != nil {
			baseLogger.logger.Warn().Msgf("Logger in context has wrong type: %T, returning base logger", value)
		}
		return baseLogger
	}
	return contextLogger
}

// RequestResponseLoggerMiddleware logs HTTP request/response details using default configuration.
// This function uses DefaultMiddlewareConfig which skips logging for /healthz and /health endpoints.
// For custom skip paths configuration, use RequestResponseLoggerMiddlewareWithConfig instead.
func RequestResponseLoggerMiddleware(c *gin.Context) {
	handler := RequestResponseLoggerMiddlewareWithConfig(nil)
	handler(c)
}

// RequestResponseLoggerMiddlewareWithConfig logs HTTP request/response details with configurable skip paths.
// If config is nil, DefaultMiddlewareConfig() is used.
//
// Example usage:
//   config := &log.MiddlewareConfig{
//       SkipPaths: []string{"/healthz", "/metrics", "/ready"},
//       SkipPathPrefixes: []string{"/internal/", "/debug/"},
//       SkipMethodPaths: map[string][]string{
//           "GET": {"/status", "/ping"},
//       },
//   }
//   router.Use(log.RequestResponseLoggerMiddlewareWithConfig(config))
func RequestResponseLoggerMiddlewareWithConfig(config *MiddlewareConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method

		// Process request
		c.Next()

		// Check if path should be skipped
		if config.ShouldSkip(method, path) {
			return
		}

		// Stop timer
		timeStamp := time.Now()

		// Use strings.Builder for efficient path concatenation
		var pathBuilder strings.Builder
		pathBuilder.Grow(len(path) + len(raw) + 1) // Pre-allocate: path + "?" + raw
		pathBuilder.WriteString(path)
		if raw != "" {
			pathBuilder.WriteByte('?')
			pathBuilder.WriteString(raw)
		}
		fullPath := pathBuilder.String()

		getCtxLogger(c).ToZerolog().Info().Str("user-agent", c.Request.UserAgent()).Str("client-ip", c.ClientIP()).Str("method", method).Str("path", fullPath).Str("protocol", c.Request.Proto).Str("http-referer", c.Request.Referer()).Int("status", c.Writer.Status()).Dur("latency-ms", timeStamp.Sub(start).Round(time.Millisecond)).Int64("bytes-received", c.Request.ContentLength).
			Int("bytes-sent", c.Writer.Size()).Msg("request")
	}
}

// setRequestMetadata sets the request metadata in the logger
func (l *Logger) setRequestMetadata(c *gin.Context) zerolog.Logger {
	requestID := c.Request.Header.Get("X-Request-ID")
	if requestID == "" {
		// generate a random request ID
		requestID = uuid.New().String()
		c.Writer.Header().Set("X-Request-ID", requestID)
	}

	traceID := c.Request.Header.Get("X-Trace-ID")
	if traceID == "" {
		if ctxTraceID, ok := c.Request.Context().Value(traceIDKey).(string); ok && ctxTraceID != "" {
			traceID = ctxTraceID
		} else {
			traceID = ""
		}
	}

	officeID := c.Request.Header.Get("x-office-id")
	if officeID == "" {
		if ctxOfficeID, ok := c.Request.Context().Value("x-office-id").(string); ok && ctxOfficeID != "" {
			officeID = ctxOfficeID
		} else {
			officeID = ""
		}
	}

	userID := c.Request.Header.Get("x-user-id")
	if userID == "" {
		if ctxUserID, ok := c.Request.Context().Value("x-user-id").(string); ok && ctxUserID != "" {
			userID = ctxUserID
		} else {
			userID = ""
		}
	}

	return l.logger.With().Str(string(requestIDContextKey), requestID).Str(string(traceIDKey), traceID).Str(string(officeIDKey), officeID).Str(string(userIDKey), userID).Logger()

}

// return default logger instance with default options
func getDefaultLogger() *Logger {

	appliedOpts := defaultLoggerOptions()
	logger := zerolog.
		New(appliedOpts.OutputWriter).
		With().
		Timestamp().
		Str(service, appliedOpts.ServiceName).
		Logger().
		Level(appliedOpts.Level)

	zerolog.DefaultContextLogger = &logger

	return &Logger{logger: &logger}
}

// Original default zerolog logger
func GetBaseLoggerInstance() *Logger {
	if baseLogger == nil {
		// Info(nil, "Base logger instance is not initialized. Returning dafult logger")
		return getDefaultLogger()
	}
	return baseLogger
}
