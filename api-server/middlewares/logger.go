package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	config "MgApplication/api-config"
	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Recover(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var err apierrors.AppError
				if e, ok := r.(error); ok {

					err = apierrors.NewAppError("500", "500", e)
				} else {
					err = apierrors.NewAppError("500", "500", fmt.Errorf("%v", r))
				}
				// Log a concise panic header
				zl := log.GetBaseLoggerInstance().ToZerolog()
				zl.Error().Str("code", err.Code).Msgf("Panic: %s", err.Error())

				if cfg.GetString("log.level") == "debug" {

					// Emit stack trace one line per log entry to avoid JSON-escaped newlines
					if err.Stack != nil {
						for _, line := range strings.Split(err.Stack.String(), "\n") {
							l := strings.TrimSpace(line)
							if l == "" {
								continue
							}
							zl.Error().Msg(l)
						}
					}
				} else {
					zl.Error().Msg(err.Stack.String())
				}
				// --- OpenTelemetry integration ---
				span := trace.SpanFromContext(c.Request.Context())
				if span != nil && span.IsRecording() {
					span.RecordError(&err)
					span.SetStatus(codes.Error, fmt.Sprintf("panic: %v", err.Stack))
				}

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func SetCtxLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.SetCtxLoggerMiddleware(c)
	}
}

func RequestResponseLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.RequestResponseLoggerMiddleware(c)
	}
}
