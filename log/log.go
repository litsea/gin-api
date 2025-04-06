package log

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

const (
	customLoggerContextKey = "litsea.gin-api.logger"
)

var _ Logger = (*DefaultLogger)(nil)

// AddAttributesFunc functions for add log attrs to gin context.
// When this function is specified,
// the error log will be passed to the middleware for processing.
type AddAttributesFunc func(c *gin.Context, key string, value any)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	GetAddAttributesFunc() AddAttributesFunc
}

type DefaultLogger struct {
	enable bool
	log    *slog.Logger
	fn     AddAttributesFunc
}

func New(sl *slog.Logger, fn ...AddAttributesFunc) *DefaultLogger {
	l := &DefaultLogger{
		enable: true,
		log:    sl,
	}
	if len(fn) > 0 {
		l.fn = fn[0]
	}

	return l
}

// GetLoggerFromContext get logger context set by SetLoggerToContext()
//
//nolint:ireturn
func GetLoggerFromContext(ctx *gin.Context) Logger {
	v, exists := ctx.Get(customLoggerContextKey)
	if !exists {
		return &DefaultLogger{}
	}

	l, ok := v.(Logger)
	if !ok {
		return &DefaultLogger{}
	}

	return l
}

func SetLoggerToContext(ctx *gin.Context, l Logger) {
	ctx.Set(customLoggerContextKey, l)
}

func (l *DefaultLogger) Debug(msg string, args ...any) {
	if !l.enable {
		return
	}

	l.log.Debug(msg, args...)
}

func (l *DefaultLogger) Info(msg string, args ...any) {
	if !l.enable {
		return
	}

	l.log.Info(msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...any) {
	if !l.enable {
		return
	}

	l.log.Warn(msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...any) {
	if !l.enable {
		return
	}

	l.log.Error(msg, args...)
}

func (l *DefaultLogger) GetAddAttributesFunc() AddAttributesFunc {
	return l.fn
}
