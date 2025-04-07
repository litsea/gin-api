package api

import (
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
