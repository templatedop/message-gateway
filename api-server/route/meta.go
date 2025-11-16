package route

import (
	"reflect"


	 "github.com/gin-gonic/gin"
)

type Meta struct {
	Method        string
	Path          string
	Name          string
	Desc          string
	Func          gin.HandlerFunc
	Req           reflect.Type
	Res           reflect.Type
	Middlewares   []gin.HandlerFunc
	DefaultStatus int
}
