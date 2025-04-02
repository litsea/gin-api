package cors

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func New(opts ...Option) gin.HandlerFunc {
	c := &cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowOrigins:     []string{"*"},
		AllowWildcard:    true,
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	for _, opt := range opts {
		opt(c)
	}

	return cors.New(*c)
}
