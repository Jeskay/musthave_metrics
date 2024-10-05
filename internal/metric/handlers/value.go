package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/gin-gonic/gin"
)

func GetCounterMetric(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		ok, value := svc.GetCounterMetric(name)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Writer.WriteString(strconv.FormatInt(value, 10))
		c.Writer.WriteHeader(http.StatusOK)
	}
}

func GetGaugeMetric(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		ok, value := svc.GetGaugeMetric(name)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Writer.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
		c.Writer.WriteHeader(http.StatusOK)
	}
}
