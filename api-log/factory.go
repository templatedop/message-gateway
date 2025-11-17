package log

import (
	"context"
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

// GetCtxLogger returns the logger embedded in the gin context
func getCtxLogger(ctx context.Context) *Logger {
	//check null
	var value any
	if ctx == nil {
		baseLogger := GetBaseLoggerInstance()
		Debug(nil, "Context is nil. Returning Base Logger.")
		return baseLogger
	} else if ginCtx, ok := ctx.(*gin.Context); ok {
		// ctx is of type *gin.Context
		if value = ginCtx.Request.Context().Value(ctxLoggerKey); value == nil {
			baseLogger := GetBaseLoggerInstance()
			// Warn(nil, "Could not find logger inside context. Returning Base Logger.")
			return baseLogger
		}
	} else {
		// ctx is of type context.Context
		if value = ctx.Value(ctxLoggerKey); value == nil {
			baseLogger := GetBaseLoggerInstance()
			// Warn(nil, "Could not find logger inside context. Returning Base Logger.")
			return baseLogger
		}
	}
	// check if the value is of type *Logger
	contextLogger, ok := value.(*Logger)
	if !ok {
		baseLogger := GetBaseLoggerInstance()
		// Warn(nil, "Logger found in context is not of type *Logger. Returning Base Logger.")
		return baseLogger
	}
	return contextLogger
}

func RequestResponseLoggerMiddleware(c *gin.Context) {
	// Start timer
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	method := c.Request.Method

	// Process request
	c.Next()
	if c.Request.Method == "GET" && c.Request.URL.Path == "/healthz" {
		return
	}
	// Stop timer
	timeStamp := time.Now()

	if raw != "" {
		path = path + "?" + raw
	}
	fullPath := path

	getCtxLogger(c).ToZerolog().Info().Str("user-agent", c.Request.UserAgent()).Str("client-ip", c.ClientIP()).Str("method", method).Str("path", fullPath).Str("protocol", c.Request.Proto).Str("http-referer", c.Request.Referer()).Int("status", c.Writer.Status()).Dur("latency-ms", timeStamp.Sub(start).Round(time.Millisecond)).Int64("bytes-received", c.Request.ContentLength).
		Int("bytes-sent", c.Writer.Size()).Msg("request")
}

// SetCtxLogger sets the request metadata in the logger
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

	officeId := c.Request.Header.Get("x-office-id")
	if officeId == "" {
		if ctxOfficeId, ok := c.Request.Context().Value("x-office-id").(string); ok && ctxOfficeId != "" {
			officeId = ctxOfficeId
		} else {
			officeId = ""
		}
	}

	userId := c.Request.Header.Get("x-user-id")
	if userId == "" {
		if ctxUserId, ok := c.Request.Context().Value("x-user-id").(string); ok && ctxUserId != "" {
			userId = ctxUserId
		} else {
			userId = ""
		}
	}

	return l.logger.With().Str(string(requestIDContextKey), requestID).Str(string(traceIDKey), traceID).Str(string(officeIDKey), officeId).Str(string(userIDKey), userId).Logger()

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
