package graceful

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/litsea/gin-api/log"
)

const (
	defaultReadTimeout         = 15 * time.Second
	defaultWriteTimeout        = 15 * time.Second
	defaultMaxShutdownDuration = 30 * time.Second
)

type Graceful struct {
	router              *gin.Engine
	server              *http.Server
	l                   log.Logger
	addr                string
	readTimeout         time.Duration
	writeTimeout        time.Duration
	maxShutdownDuration time.Duration
	cleanup             []cleanup
}

type cleanup func()

func New(r *gin.Engine, opts ...Option) *Graceful {
	g := &Graceful{
		router:              r,
		l:                   log.NewDisabled(), // default disabled
		addr:                ":8080",
		readTimeout:         defaultReadTimeout,
		writeTimeout:        defaultWriteTimeout,
		maxShutdownDuration: defaultMaxShutdownDuration,
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.server == nil {
		g.server = &http.Server{
			Addr:              g.addr,
			Handler:           g.router,
			ReadTimeout:       g.readTimeout,
			WriteTimeout:      g.writeTimeout,
			ReadHeaderTimeout: time.Second * 5, // Set a reasonable ReadHeaderTimeout value
		}
	}

	return g
}

func (g *Graceful) Run() {
	// kill -INT <pid>
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		// serve connections
		g.l.Info("graceful.Run: server start running...")
		if err := g.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			g.l.Error("graceful.Run ListenAndServe failed", "err", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	g.l.Info("graceful.Run: server shutting down gracefully, press Ctrl+C again to force")
	stop()

	// Wait for requests currently being handling
	ctx, cancel := context.WithTimeout(context.Background(), g.maxShutdownDuration)
	defer cancel()

	if err := g.server.Shutdown(ctx); err != nil {
		g.l.Warn("graceful.Run: server forced to shutdown", "err", err)
	}

	for _, c := range g.cleanup {
		c()
	}

	g.l.Info("graceful.Run: server exited")
}
