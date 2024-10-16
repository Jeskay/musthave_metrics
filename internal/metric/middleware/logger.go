package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		t := time.Now()
		ctx.Next()
		logger.Info(
			"incoming request",
			slog.String("uri", ctx.Request.URL.RawPath),
			slog.String("method", ctx.Request.Method),
			slog.Duration("latency", time.Since(t)),
		)
		logger.Info(
			"response",
			slog.Int("status", ctx.Request.Response.StatusCode),
			slog.Int64("size", ctx.Request.Response.ContentLength),
		)
	}
}
