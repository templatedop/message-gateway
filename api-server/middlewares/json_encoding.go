package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

// CustomJSON sets up goccy/go-json encoding/decoding functions in the gin context
// This allows handlers to use high-performance JSON operations via c.Get("json") and c.Get("jsonUnmarshal")
// goccy/go-json provides 2-8x better performance than encoding/json in this environment.
func CustomJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("json", json.Marshal)
		c.Set("jsonUnmarshal", json.Unmarshal)
		c.Next()
	}
}
