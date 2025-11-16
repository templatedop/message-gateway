package middlewares

import (
	"net/http"

	"MgApplication/api-server/ratelimiter"
	"MgApplication/api-server/router-adapter"
)

// RateLimiter returns a middleware that implements rate limiting using LeakyBucket algorithm
// If the rate limit is exceeded, it returns 429 Too Many Requests
func RateLimiter(bucket *ratelimiter.LeakyBucket) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Check if request is allowed by rate limiter
		if bucket.Allow() {
			// Request allowed, proceed to next middleware/handler
			return next()
		}

		// Rate limit exceeded, return 429
		return ctx.JSON(http.StatusTooManyRequests, map[string]string{
			"error": "rate limit exceeded",
			"message": "too many requests, please try again later",
		})
	}
}
