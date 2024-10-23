package middleware

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

func Encoder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
			gzWriter, err := gzip.NewWriterLevel(ctx.Writer, gzip.BestSpeed)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
			}
			ctx.Writer = &gzipWriter{ctx.Writer, gzWriter}
			ctx.Header("Content-Encoding", "gzip")
			defer func() {
				ctx.Header("Content-Length", fmt.Sprint(ctx.Writer.Size()))
				gzWriter.Close()
			}()
		}
		ctx.Next()
	}
}
