package ginadapter

import (
	"MgApplication/api-server/router-adapter"

	"github.com/gin-gonic/gin"
)

// WrapGinMiddleware converts a Gin middleware to framework-agnostic middleware
// This allows using existing Gin middlewares with the adapter pattern
func WrapGinMiddleware(ginMiddleware gin.HandlerFunc) routeradapter.MiddlewareFunc {
	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Get native Gin context
		ginCtx, ok := ctx.GetNativeContext().(*gin.Context)
		if !ok {
			// If not a Gin context, skip this middleware
			return next()
		}

		// Call Gin middleware
		ginMiddleware(ginCtx)

		// Check if middleware aborted the request
		if ginCtx.IsAborted() {
			return nil
		}

		// Check for errors
		if len(ginCtx.Errors) > 0 {
			return ginCtx.Errors.Last().Err
		}

		// Call next
		return next()
	}
}

// ConvertGinHandler converts a Gin HandlerFunc to framework-agnostic handler
// This allows using existing Gin handlers with the adapter
func ConvertGinHandler(ginHandler gin.HandlerFunc) routeradapter.HandlerFunc {
	return func(ctx *routeradapter.RouterContext) error {
		// Get native Gin context
		ginCtx, ok := ctx.GetNativeContext().(*gin.Context)
		if !ok {
			// If not a Gin context, return error
			return &routeradapter.RouterError{
				Code:    "invalid_context",
				Message: "context is not a Gin context",
			}
		}

		// Call Gin handler
		ginHandler(ginCtx)

		// Check for errors
		if len(ginCtx.Errors) > 0 {
			return ginCtx.Errors.Last().Err
		}

		return nil
	}
}

// GetGinContext extracts the Gin context from RouterContext
// Returns nil if the context is not from Gin adapter
func GetGinContext(ctx *routeradapter.RouterContext) *gin.Context {
	if ctx == nil {
		return nil
	}

	ginCtx, _ := ctx.GetNativeContext().(*gin.Context)
	return ginCtx
}

// MustGetGinContext is like GetGinContext but panics if context is not Gin
// Use this when you're certain the context is from Gin adapter
func MustGetGinContext(ctx *routeradapter.RouterContext) *gin.Context {
	ginCtx := GetGinContext(ctx)
	if ginCtx == nil {
		panic("context is not from Gin adapter")
	}
	return ginCtx
}
