package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

func CustomJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("json", json.Marshal)
		c.Set("jsonUnmarshal", json.Unmarshal)
		c.Next()
	}
}
