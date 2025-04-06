package log

import (
	"github.com/gin-gonic/gin"
)

func Middleware(l Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if l != nil {
			SetLoggerToContext(c, l)
		}

		c.Next()
	}
}
