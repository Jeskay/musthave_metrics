package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipDecoder returns handler function which decompresses the request
// body if it has been compressed using gzip.
func GzipDecoder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.Contains(ctx.GetHeader("Content-Encoding"), "gzip") && ctx.GetHeader("Content-Length") != "0" {
			gzReader, err := gzip.NewReader(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			ctx.Request.Body = gzReader
			defer gzReader.Close()
		}

		ctx.Next()
	}
}
