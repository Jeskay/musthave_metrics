package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"sync"

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

type gzipHandler struct {
	gzPool sync.Pool
}

func NewGzipHandler() *gzipHandler {
	return &gzipHandler{
		gzPool: sync.Pool{
			New: func() interface{} {
				gz, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
				return gz
			},
		},
	}
}

func (g *gzipHandler) Handle(ctx *gin.Context) {
	if strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
		if gz, ok := g.gzPool.Get().(*gzip.Writer); ok {
			gz.Reset(ctx.Writer)
			ctx.Header("Content-Encoding", "gzip")
			ctx.Writer = &gzipWriter{ctx.Writer, gz}
			defer func() {
				if ctx.Writer.Size() < 0 {
					gz.Reset(io.Discard)
				}
				gz.Close()
				if ctx.Writer.Size() > -1 {
					ctx.Header("Content-Length", fmt.Sprint(ctx.Writer.Size()))
				}
				g.gzPool.Put(gz)
			}()
		}
	}
	ctx.Next()
}
