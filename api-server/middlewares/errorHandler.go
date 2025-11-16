package middlewares

import (
	apierrors "MgApplication/api-errors"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		err := c.Errors.Last()
		if err == nil {
			return
		}
		apierrors.HandleCommonError(c, err.Err)
	}
}
