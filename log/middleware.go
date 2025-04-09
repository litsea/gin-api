package log

import (
	"bytes"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	customLoggerContextKey    = "litsea.gin-api.log"
	requestIDCtxKey           = "litsea.gin-api.log.request-id"
	requestBodyCtxKey         = "litsea.gin-api.log.request-body"
	defaultRequestIDHeaderKey = "X-Request-ID"
)

func Middleware(l Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if l != nil {
			SetLoggerToContext(c, l)
		}

		if l.Config().withRequestBody {
			rb := newRequestBody(RequestBodyMaxSize, l.Config().withRequestBody)
			err := rb.read(c)
			if err != nil {
				l.Error("log.Middleware: read requestBody failed", "err", fmt.Errorf("%w", err))
			} else {
				c.Set(requestBodyCtxKey, rb)
			}
		}

		if l.Config().requestIDHeaderKey != "" {
			requestID := c.GetHeader(l.Config().requestIDHeaderKey)
			if requestID == "" {
				requestID = uuid.New().String()
				c.Header(l.Config().requestIDHeaderKey, requestID)
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

func GetRequestBody(ctx *gin.Context) (*bytes.Buffer, int) {
	body, ok := ctx.Get(requestBodyCtxKey)
	if !ok {
		return bytes.NewBuffer([]byte{}), 0
	}

	if rb, ok := body.(*requestBody); ok {
		if rb.body == nil {
			return bytes.NewBuffer([]byte{}), 0
		}
		return rb.body, rb.bytes
	}

	return bytes.NewBuffer([]byte{}), 0
}
