package handlers

import (
	"net/http"

	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/gin-gonic/gin"
)

func Ping(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc.DBHealth() {
			c.Writer.WriteHeader(http.StatusOK)
		} else {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
	}
}
