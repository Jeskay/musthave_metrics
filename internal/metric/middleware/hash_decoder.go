package middleware

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jeskay/musthave_metrics/internal"
)

// HashDecoder returns function that handles requests with hash sum.
// If a request has hash sum header, handler function checks if the
// content of the request has been modified and aborts request when hashes do not align.
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
