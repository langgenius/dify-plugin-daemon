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
