package cors

import (
	"time"

	"github.com/gin-contrib/cors"
)

type Option func(*cors.Config)

func WithAllowMethods(ms []string) Option {
	return func(c *cors.Config) {
		if len(ms) > 0 {
			c.AllowMethods = ms
		}
	}
}

func WithAllowHeaders(hs []string) Option {
	return func(c *cors.Config) {
		if len(hs) > 0 {
			c.AllowHeaders = hs
		}
	}
}

func WithAllowOrigin(os []string) Option {
	return func(c *cors.Config) {
		if len(os) > 0 {
			c.AllowOrigins = os
		}
	}
}

func WithAllowWildcard(aw bool) Option {
	return func(c *cors.Config) {
		c.AllowWildcard = aw
	}
}

func WithAllowCredentials(ac bool) Option {
	return func(c *cors.Config) {
		c.AllowCredentials = ac
	}
}

func WithMaxAge(ma time.Duration) Option {
	return func(c *cors.Config) {
		c.MaxAge = ma
	}
}
