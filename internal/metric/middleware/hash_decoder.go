package middleware

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/Jeskay/musthave_metrics/internal"

	"github.com/gin-gonic/gin"
)

func HashDecoder(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, ok := ctx.Request.Header[http.CanonicalHeaderKey(internal.HashHeader)]
		if ok {
			hash := ctx.GetHeader(internal.HashHeader)
			if hash == "" {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			payload, err := io.ReadAll(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			ctx.Set(gin.BodyBytesKey, payload)
			hexData, err := hex.DecodeString(hash)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			nData, err := hashBytes(payload, key)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			if !bytes.Equal(nData, hexData) {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}
		ctx.Next()
	}
}
