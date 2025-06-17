package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

func Decipher(privateKey *rsa.PrivateKey) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		header := ctx.Request.Header.Get("Ciphered")
		if strings.Contains(header, "true") {
			msg, err := io.ReadAll(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(400)
				return
			}
			cipherByte, err := rsa.DecryptOAEP(
				sha256.New(),
				rand.Reader,
				privateKey,
				msg,
				[]byte(""),
			)
			if err != nil {
				ctx.AbortWithStatus(400)
				return
			}
			ctx.Request.Body.Close()
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(cipherByte))
		}
		ctx.Next()
	}
}
