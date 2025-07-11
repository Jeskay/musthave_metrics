package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"github.com/Jeskay/musthave_metrics/internal"
)

type hashWriter struct {
	gin.ResponseWriter
	payload *bytes.Buffer
	hashKey string
}

func (h *hashWriter) WriteString(s string) (int, error) {
	h.payload.Write([]byte(s))
	return h.ResponseWriter.Write([]byte(s))
}

func (h *hashWriter) Write(data []byte) (int, error) {
	h.payload.Write(data)
	return h.ResponseWriter.Write(data)
}

func (h *hashWriter) WriteHeader(code int) {
	h.ResponseWriter.WriteHeader(code)
}

// HashEncoder returns function that handles responses with hash sum.
// It replaces gin writer with hashWriter, that adds hash sum of the response to the headers.
func HashEncoder(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader(internal.HashHeader) == "" {
			ctx.Next()
			return
		}
		h := &hashWriter{ctx.Writer, &bytes.Buffer{}, key}
		ctx.Next()
		hashed, err := hashBytes(h.payload.Bytes(), h.hashKey)
		if err != nil {
			return
		}
		h.Header().Add(internal.HashHeader, hex.EncodeToString(hashed))
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
