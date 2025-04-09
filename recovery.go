package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/litsea/gin-api/log"
)

func HandleRecovery() gin.RecoveryFunc {
	return func(ctx *gin.Context, err any) {
		rErr, ok := err.(error)
		if !ok {
			rErr = fmt.Errorf("%v", err)
		}

		Error(ctx, fmt.Errorf("panic error: %w", rErr))
	}
}

// Recovery middleware
func Recovery(recovery gin.RecoveryFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				l := log.GetLoggerFromContext(ctx)

				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var (
					brokenPipe bool
					msgErr     string
					rErr       error
				)
				if ne, ok := err.(*net.OpError); ok {
					rErr = ne
					var se *os.SyscallError
					if errors.As(ne.Err, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") {
							brokenPipe = true
							msgErr = "Panic error: broken pipe"
						} else if strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
							msgErr = "Panic error: connection reset by peer"
						}
					}
				}

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					l.ErrorRequest(ctx, msgErr, map[string]any{
						"status": http.StatusInternalServerError,
						"err":    rErr,
					})

					ctx.Abort()
					return
				}

				rErr, ok := err.(error)
				if !ok {
					rErr = fmt.Errorf("%v", err)
				}

				recovery(ctx, rErr)
			}
		}()

		ctx.Next()
	}
}
