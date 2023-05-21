package middleware

import (
	"example.com/m/v2/ioc"
	"github.com/gin-gonic/gin"
)

func ErrorHandler(ioc *ioc.IOC) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		ioc.Logger.Trace("ErrorHandler() middleware called with # of errors:", len(c.Errors))
		for _, ginErr := range c.Errors {
			// Log error on the server-side
			ioc.Logger.Error(ginErr.Error())
		}
	}
}
