package log

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	customLoggerContextKey    = "litsea.gin-api.log"
	requestIDCtxKey           = "litsea.gin-api.log.request-id"
	defaultRequestIDHeaderKey = "X-Request-ID"
)

func Middleware(l Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if l != nil {
			SetLoggerToContext(c, l)
		}

		if l.GetRequestIDKey() != "" {
			requestID := c.GetHeader(l.GetRequestIDKey())
			if requestID == "" {
				requestID = uuid.New().String()
				c.Header(l.GetRequestIDKey(), requestID)
			}
			c.Set(requestIDCtxKey, requestID)
		}

		c.Next()
	}
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

func GetRequestID(ctx *gin.Context) string {
	requestID, ok := ctx.Get(requestIDCtxKey)
	if !ok {
		return ""
	}

	if id, ok := requestID.(string); ok {
		return id
	}

	return ""
}
