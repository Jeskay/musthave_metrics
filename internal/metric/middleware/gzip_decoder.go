package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GzipDecoder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.Contains(ctx.GetHeader("Content-Encoding"), "gzip") {
			gzReader, err := gzip.NewReader(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
			}
			ctx.Request.Body = gzReader
			defer gzReader.Close()
		}

		ctx.Next()
	}
}
