package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BodyLimiter(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

func BodyLimitErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, err := range c.Errors {
			if err.Err != nil && err.Err.Error() == "http: request body too large" {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
					"error": "request body too large",
				})
				return
			}
		}
	}
}
