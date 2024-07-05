package api

import (
	"github.com/gin-gonic/gin"
)

// Error implementation - found in Effective Go
type rcError string

func (err rcError) Error() string {
	return string(err)
}

func HandleF(hc interface{ HandleHTTP(c *gin.Context) }) gin.HandlerFunc {
	return func(c *gin.Context) {
		hc.HandleHTTP(c)
	}
}
