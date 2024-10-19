package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
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

func GetMetricJson(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var metric dto.Metrics
		if err := c.ShouldBindJSON(&metric); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		switch metric.MType {
		case string(internal.CounterMetric):
			ok, v := svc.GetCounterMetric(metric.ID)
			if !ok {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.JSON(http.StatusOK, dto.NewCounterMetrics(metric.ID, v))
		case string(internal.GaugeMetric):
			ok, v := svc.GetGaugeMetric(metric.ID)
			if !ok {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.JSON(http.StatusOK, dto.NewGaugeMetrics(metric.ID, v))
		default:
			c.Writer.WriteHeader(http.StatusBadRequest)
		}
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
