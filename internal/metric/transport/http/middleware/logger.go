// Package middleware contains realization of the function and handlers that
// are executed before and/or after requests.
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns handler function which logs incoming request and outgoing response data.
// Function outputs request URI, method and handling time. Also the information includes
// response status and size.
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
			slog.Int("status", ctx.Writer.Status()),
			slog.Int("size", ctx.Writer.Size()),
		)
	}
}
