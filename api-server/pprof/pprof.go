package handler

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func PprofSymbolHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Symbol(c.Writer, c.Request)
	}
}

func PprofAllocsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
	}
}

func PprofBlockHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
	}
}

func PprofCmdlineHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Cmdline(c.Writer, c.Request)
	}
}

func PprofGoroutineHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
	}
}

func PprofHeapHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
	}
}

func PprofIndexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Index(c.Writer, c.Request)
	}
}

func PprofMutexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
	}
}

func PprofProfileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Profile(c.Writer, c.Request)
	}
}

func PprofThreadCreateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("threadcreate").ServeHTTP(c.Writer, c.Request)
	}
}

func PprofTraceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Trace(c.Writer, c.Request)
	}
}
