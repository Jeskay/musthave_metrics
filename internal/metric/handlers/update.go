package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/gin-gonic/gin"
)

func UpdateCounterMetricRaw(svc *metric.MetricService) gin.HandlerFunc {
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

func UpdateMetricJson(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var metric dto.Metrics
		if err := c.ShouldBindJSON(&metric); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if metric.MType == string(internal.CounterMetric) {
			if err := svc.SetCounterMetric(metric.ID, *metric.Delta); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			if ok, v := svc.GetCounterMetric(metric.ID); ok {
				metric.Delta = &v
				c.JSON(http.StatusOK, metric)
			}
		} else {
			svc.SetGaugeMetric(metric.ID, *metric.Value)
			c.JSON(http.StatusOK, metric)
		}
	}
}

func UpdateMetricsJson(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var metrics []dto.Metrics
		if err := c.ShouldBindJSON(&metrics); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}
		metrics = dto.OptimizeMetrics(metrics)
		svc.SetMetrics(metrics)
		keys := make([]string, len(metrics))
		for i, v := range metrics {
			keys[i] = v.ID
		}
		updatedMetrics, err := svc.GetMetrics(keys)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		c.JSON(http.StatusOK, updatedMetrics)
	}
}

func UpdateGaugeMetricRaw(svc *metric.MetricService) gin.HandlerFunc {
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
