package log

import (
	"context"
	"log/slog"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	RequestIDKey = "id"

	RequestBodyMaxSize = 64 * 1024 // 64KB

	HiddenRequestHeaders = map[string]struct{}{
		"authorization": {},
		"cookie":        {},
		"set-cookie":    {},
		"x-auth-token":  {},
		"x-csrf-token":  {},
		"x-xsrf-token":  {},
	}

	// RequestIDHeaderKey Formatted with http.CanonicalHeaderKey

	defaultLogExtraAttrs = map[string]any{
		"log.handler": "litsea.gin-api.log",
	}
)

var _ Logger = (*DefaultLogger)(nil)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	DebugRequest(ctx *gin.Context, msg string, attrs map[string]any)
	InfoRequest(ctx *gin.Context, msg string, attrs map[string]any)
	WarnRequest(ctx *gin.Context, msg string, attrs map[string]any)
	ErrorRequest(ctx *gin.Context, msg string, attrs map[string]any)
	Config() *Config
}

type Config struct {
	withUserAgent      bool
	withRequestBody    bool
	withRequestHeader  bool
	withStackTrace     bool
	requestIDHeaderKey string
	extraAttrs         map[string]any
}

type DefaultLogger struct {
	enable bool
	sl     *slog.Logger
	cfg    *Config
}

func New(sl *slog.Logger, opts ...Option) *DefaultLogger {
	if sl == nil {
		return NewDisabled()
	}

	l := &DefaultLogger{
		enable: true,
		sl:     sl,
	}

	cfg := &Config{
		withUserAgent:      false,
		withRequestBody:    false,
		withRequestHeader:  false,
		withStackTrace:     false,
		requestIDHeaderKey: defaultRequestIDHeaderKey,
		extraAttrs:         defaultLogExtraAttrs,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	l.cfg = cfg

	return l
}

func NewDisabled() *DefaultLogger {
	return &DefaultLogger{
		cfg: &Config{},
	}
}

func (l *DefaultLogger) Config() *Config {
	return l.cfg
}

func (l *DefaultLogger) Debug(msg string, args ...any) {
	l.logContext(context.Background(), slog.LevelDebug, msg, args...)
}

func (l *DefaultLogger) Info(msg string, args ...any) {
	l.logContext(context.Background(), slog.LevelInfo, msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...any) {
	l.logContext(context.Background(), slog.LevelWarn, msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...any) {
	l.logContext(context.Background(), slog.LevelError, msg, args...)
}

func (l *DefaultLogger) logContext(ctx context.Context, lv slog.Level, msg string, args ...any) {
	if !l.enable || !l.sl.Enabled(ctx, lv) {
		return
	}

	// skip [runtime.Callers, this function, this function's caller]
	skip := 3
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])

	r := slog.NewRecord(time.Now(), lv, msg, pcs[0])
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.sl.Handler().Handle(ctx, r)
}

func (l *DefaultLogger) DebugRequest(ctx *gin.Context, msg string, attrs map[string]any) {
	l.logRequest(ctx, slog.LevelDebug, msg, attrs)
}

func (l *DefaultLogger) InfoRequest(ctx *gin.Context, msg string, attrs map[string]any) {
	l.logRequest(ctx, slog.LevelInfo, msg, attrs)
}

func (l *DefaultLogger) WarnRequest(ctx *gin.Context, msg string, attrs map[string]any) {
	l.logRequest(ctx, slog.LevelWarn, msg, attrs)
}

func (l *DefaultLogger) ErrorRequest(ctx *gin.Context, msg string, attrs map[string]any) {
	l.logRequest(ctx, slog.LevelError, msg, attrs)
}

func (l *DefaultLogger) logRequest(ctx *gin.Context, lv slog.Level, msg string, args map[string]any) {
	if !l.enable {
		return
	}

	if !l.sl.Enabled(ctx, lv) {
		return
	}

	method := ctx.Request.Method
	host := ctx.Request.Host
	path := ctx.Request.URL.Path

	var status int
	st, ok := args["status"]
	if ok {
		if sts, ok := st.(int); ok {
			status = sts
		}
	}

	params := map[string]string{}
	for _, p := range ctx.Params {
		params[p.Key] = p.Value
	}

	requestAttributes := []slog.Attr{
		slog.String("method", method),
		slog.String("host", host),
		slog.String("path", path),
		slog.String("query", ctx.Request.URL.RawQuery),
		slog.Any("params", params),
		slog.String("route", ctx.FullPath()),
		slog.String("ip", ctx.ClientIP()),
		slog.String("referer", ctx.Request.Referer()),
	}

	if l.cfg.withUserAgent {
		requestAttributes = append(requestAttributes, slog.String("user-agent",
			ctx.Request.UserAgent()))
	}

	requestID := GetRequestID(ctx)
	if l.cfg.requestIDHeaderKey != "" && requestID != "" {
		requestAttributes = append(requestAttributes, slog.String(RequestIDKey, requestID))
	}

	// request headers
	if l.cfg.withRequestHeader {
		kv := make([]any, 0, len(ctx.Request.Header))

		for k, v := range ctx.Request.Header {
			if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
				continue
			}
			kv = append(kv, slog.Any(k, v))
		}

		requestAttributes = append(requestAttributes, slog.Group("header", kv...))
	}

	// request body
	if l.cfg.withRequestBody {
		body, n := GetRequestBody(ctx)
		requestAttributes = append(requestAttributes, slog.Int("length", n))
		requestAttributes = append(requestAttributes, slog.String("body", body.String()))
	}

	attributes := []slog.Attr{
		{
			Key:   "request",
			Value: slog.GroupValue(requestAttributes...),
		},
	}

	if status > 0 {
		responseAttributes := []slog.Attr{
			slog.Int("status", status),
		}

		attributes = append(attributes, slog.Attr{
			Key:   "response",
			Value: slog.GroupValue(responseAttributes...),
		})
	}

	// custom context values
	for k, v := range l.cfg.extraAttrs {
		attributes = append(attributes, slog.Any(k, v))
	}

	for k, v := range args {
		if k == "status" {
			continue
		}
		attributes = append(attributes, slog.Any(k, v))
	}

	if l.cfg.withStackTrace {
		trace := debug.Stack()
		attributes = append(attributes, slog.Any("stacktrace", string(trace)))
	}

	// skip [runtime.Callers, this function, l.XXXRequest, l.XXXRequest's caller]
	skip := 4
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])

	r := slog.NewRecord(time.Now(), lv, msg, pcs[0])
	r.AddAttrs(attributes...)

	_ = l.sl.Handler().Handle(ctx, r)
}

func DebugRequest(ctx *gin.Context, msg string, attrs map[string]any) {
	l := GetLoggerFromContext(ctx)
	l.DebugRequest(ctx, msg, attrs)
}

func InfoRequest(ctx *gin.Context, msg string, attrs map[string]any) {
	l := GetLoggerFromContext(ctx)
	l.InfoRequest(ctx, msg, attrs)
}

func WarnRequest(ctx *gin.Context, msg string, attrs map[string]any) {
	l := GetLoggerFromContext(ctx)
	l.WarnRequest(ctx, msg, attrs)
}

func ErrorRequest(ctx *gin.Context, msg string, attrs map[string]any) {
	l := GetLoggerFromContext(ctx)
	l.ErrorRequest(ctx, msg, attrs)
}
