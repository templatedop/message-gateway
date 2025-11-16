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
		c.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(allowOrigins, ","))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ","))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ","))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
