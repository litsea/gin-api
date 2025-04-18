package graceful

import (
	"net/http"
	"time"

	"github.com/litsea/gin-api/log"
)

type Option func(*Graceful)

func WithServer(srv *http.Server) Option {
	return func(g *Graceful) {
		g.server = srv
	}
}

func WithAddr(addr string) Option {
	return func(c *Graceful) {
		c.addr = addr
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Graceful) {
		c.readTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Graceful) {
		c.writeTimeout = timeout
	}
}

func WithStopTimeout(timeout time.Duration) Option {
	return func(c *Graceful) {
		c.stopTimeout = timeout
	}
}

func WithCleanup(cleanup ...cleanup) Option {
	return func(c *Graceful) {
		if len(cleanup) > 0 {
			c.cleanup = append(c.cleanup, cleanup...)
		}
	}
}

func WithLogger(l log.Logger) Option {
	return func(c *Graceful) {
		if l != nil {
			c.l = l
		}
	}
}
