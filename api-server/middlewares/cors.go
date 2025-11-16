package middlewares

import (
	"strings"

	config "MgApplication/api-config"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {

	allowOrigins := cfg.GetStringSlice("alloworigins")
	allowMethods := cfg.GetStringSlice("allowmethods")
	allowHeaders := cfg.GetStringSlice("allowheaders")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if the request origin is in the allowed list
		originAllowed := false
		for _, allowed := range allowOrigins {
			if allowed == "*" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
				originAllowed = true
				break
			}
			if origin == allowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				originAllowed = true
				break
			}
		}

		// Only set other CORS headers if origin is allowed
		if originAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ", "))
			c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
			c.Writer.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours
		}

		if c.Request.Method == "OPTIONS" {
			if originAllowed {
				c.AbortWithStatus(204)
			} else {
				c.AbortWithStatus(403)
			}
			return
		}

		c.Next()
	}
}
