package middlewares

import (
	"net/http"
	//r "testencrypt/rate-gin/ratelimiter"

	rate "MgApplication/api-server/ratelimiter"

	"github.com/gin-gonic/gin"
)

func RateMiddleware(globalBucket *rate.LeakyBucket) gin.HandlerFunc {
	return func(c *gin.Context) {

		if globalBucket.Allow() {
			c.Next()
		} else {

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Traffic shaping limit exceeded",
			})
		}

	}
}
