package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/gin-gonic/gin"
)

func UpdateCounterMetric(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		v := c.Param("value")
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		svc.SetCounterMetric(name, value)
		c.Writer.WriteHeader(http.StatusOK)
	}
}

func UpdateGaugeMetric(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		v := c.Param("value")
		value, err := strconv.ParseFloat(v, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		svc.SetGaugeMetric(name, value)
		c.Writer.WriteHeader(http.StatusOK)
	}
}
