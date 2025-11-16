package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
	//	config "MgApplication/api-config"
)

type timeoutkey string

const ServerTimeOutKey timeoutkey = "timeout"

// func Timeout(cfg *config.Config) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx := c.Request.Context()
// 		// t:=cfg.GetInt("timeout")
// 		// t := cfg.GetDuration("server.timeout")

// 		// ctx, cancel := context.WithTimeout(ctx, t*time.Second)
// 		// defer cancel()

// 		c.Request = c.Request.WithContext(ctx)
// 		c.Next()
// 	}
// }

func TimeoutMiddleware1(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"error": "request timeout",
			})
			return
		}
	}
}

// func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
// 		defer cancel()

// 		c.Request = c.Request.WithContext(ctx)
// 		done := make(chan struct{})
// 		var handlerError error

// 		go func() {
//             defer close(done)
// 		defer func() {
// 			if err := recover(); err != nil {
// 				log.Error(ctx, "PANIC recovered: %v\n", err)
// 				if !c.Writer.Written() {
// 					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
// 						"error": "Internal server error",
// 					})
// 				}
// 			}
// 		}()

// 		done := make(chan struct{})

// 		go func() {
// 			defer close(done)
// 			defer func() {
// 				if err := recover(); err != nil {
// 					log.Error(ctx, "Handler PANIC recovered: %v\n", err)
// 				}
// 			}()

// 			c.Next()
// 		}()

// 		select {
// 		case <-done:
// 			// Normal completion
// 			return
// 		case <-ctx.Done():
// 			if ctx.Err() == context.DeadlineExceeded {
// 				// Timeout occurred
// 				c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
// 					"error": "Request timeout",
// 				})
// 			}
// 			return
// 		}
// 	}
// }

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		var handlerError error

		go func() {
			defer close(done)
			defer func() {
				if err := recover(); err != nil {
					handlerError = fmt.Errorf("handler panic: %v", err)
					log.Error(ctx, "Handler PANIC recovered: %v\n", err)
				}
			}()

			c.Next()
		}()

		select {
		case <-done:
			if handlerError != nil && !c.Writer.Written() {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
			return
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
					"error": "Request timeout",
				})
			}
			return
		}
	}
}
