package api

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/litsea/gin-api/errcode"
)

func HandleNotFound() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		Error(ctx, errcode.ErrNotFound)
	}
}

func HandleMethodNotAllowed() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		Error(ctx, errcode.ErrMethodNotAllowed)
	}
}

func HandleHealthCheck() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Cache-Control", "no-store")
		ctx.Header("Access-Control-Allow-Origin", "*")
		Success(ctx, "OK")
	}
}

func RouteRegisterPprof(r *gin.Engine, token string) {
	if token == "" {
		return
	}

	g := r.Group("/debug", func(ctx *gin.Context) {
		req := &struct {
			Token string `form:"token"`
		}{}
		err := ctx.ShouldBind(req)
		if err != nil || req.Token != token {
			Error(ctx, errcode.ErrForbidden)
			ctx.Abort()
		}
		ctx.Next()
	})
	pprof.RouteRegister(g, "pprof")
}
