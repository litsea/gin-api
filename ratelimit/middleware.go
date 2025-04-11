package ratelimit

import (
	"github.com/gin-gonic/gin"

	api "github.com/litsea/gin-api"
	"github.com/litsea/gin-api/errcode"
)

func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := l.LimitByRequest(c.Writer, c.Request)
		if err != nil {
			api.Error(c, errcode.ErrTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}
