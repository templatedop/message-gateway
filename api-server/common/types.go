package common

import (
	"MgApplication/api-server/util/wrapper"

	"github.com/gin-gonic/gin"
)

type (
	MiddlewareGroup = []gin.HandlerFunc
	GinAppWrapper   = wrapper.Wrapper[*gin.Engine]
)
