package middleware

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

type hashWriter struct {
	gin.ResponseWriter
	hashKey string
}

func (h *hashWriter) WriteString(s string) (int, error) {
	return h.Write([]byte(s))
}

func (h *hashWriter) Write(data []byte) (int, error) {
	hashed, err := hashBytes(data, h.hashKey)
	if err != nil {
		return 0, err
	}
	h.Header().Add("HashSHA256", hex.EncodeToString(hashed))
	return h.Write(data)
}

func (h *hashWriter) WriteHeader(code int) {
	h.ResponseWriter.WriteHeader(code)
}

func HashEncoder(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader("HashSHA256") != "" {
			ctx.Writer = &hashWriter{ctx.Writer, key}
		}
		ctx.Next()
	}
}

func hashBytes(data []byte, key string) ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return nil, err
	}
	if _, err := h.Write([]byte(key)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}