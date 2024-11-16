package middleware

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HashDecoder(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hash := ctx.GetHeader("HashSHA256")
		if hash != "" {
			payload, err := io.ReadAll(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
			}
			ctx.Set(gin.BodyBytesKey, payload)
			hex_data, err := hex.DecodeString(hash)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
			}
			n_data, err := hashBytes(payload, key)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
			}
			if !bytes.Equal(n_data, hex_data) {
				ctx.AbortWithStatus(http.StatusBadRequest)
			}
		}
		ctx.Next()
	}
}
