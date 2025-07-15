package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SubnetChecker(network *net.IPNet) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := net.ParseIP(ctx.Request.RemoteAddr)
		if !network.Contains(ip) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.Next()
	}
}
