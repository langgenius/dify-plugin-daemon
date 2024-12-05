package controllers

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func PprofIndex(c *gin.Context) {
	pprof.Index(c.Writer, c.Request)
}

func PprofCmdline(c *gin.Context) {
	pprof.Cmdline(c.Writer, c.Request)
}

func PprofProfile(c *gin.Context) {
	pprof.Profile(c.Writer, c.Request)
}

func PprofSymbol(c *gin.Context) {
	pprof.Symbol(c.Writer, c.Request)
}

func PprofTrace(c *gin.Context) {
	pprof.Trace(c.Writer, c.Request)
}

func PprofGoroutine(c *gin.Context) {
	pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
}

func PprofHeap(c *gin.Context) {
	pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
}

func PprofAllocs(c *gin.Context) {
	pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
}

func PprofBlock(c *gin.Context) {
	pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
}

func PprofMutex(c *gin.Context) {
	pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
}

func PprofThreadcreate(c *gin.Context) {
	pprof.Handler("threadcreate").ServeHTTP(c.Writer, c.Request)
}
