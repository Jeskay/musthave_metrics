package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/gin-gonic/gin"
)

// UpdateCounterMetricRaw handles raw update requests that target counter metrics.
//
//	Method: POST
//	Endpoint: /update/counter/{name}/{value}
//
// Example usage with curl:
//
//	curl -X POST http://localhost:9009/update/counter/test/100
//
//	On success, returns HTTP 200 OK.
//	On invalid value returns HTTP 400 Bad request.
func UpdateCounterMetricRaw(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		v := c.Param("value")
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err := svc.SetCounterMetric(name, value); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		c.Writer.WriteHeader(http.StatusOK)
	}
}

// UpdateMetricJson handles update requests in json format.
//
//	Method: POST
//	Endpoint: /update/
//
// Expected JSON body:
//
//	{
//		"id": "metric1",
//		"type": "counter",
//		"delta": 300
//	}
//
// Example usage with curl:
//
//	curl -X POST http://localhost:9009/update/ \
//			-H "Content-Type: application/json"
//			-d '{"id": "metric1", "type": "counter", "delta": 300}'
//
//	On success, returns HTTP 200 OK with updated metric.
//	On invalid JSON body returns HTTP 400 Bad request.
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
			if err := svc.SetGaugeMetric(metric.ID, *metric.Value); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			c.JSON(http.StatusOK, metric)
		}
	}
}

// UpdateMetricsJson handles update requests for multiple metrics in JSON format.
//
//	Method: POST
//	Endpoint: /updates/
//
// Expected JSON body:
//
//	[
//		{
//			"id": "metric1",
//			"type": "counter",
//			"delta": 100
//		},
//		{
//			"id": "metric1",
//			"type": "counter",
//			"delta": 200
//		}
//	]
//
// Example usage with curl:
//
//	curl -X POST http://localhost:9009/updates/ \
//			-H "Content-Type: application/json"
//			-d '[{"id": "metric1", "type": "counter", "delta": 300}, {"id": "metric2", "type": "counter", "delta": 200}]'
//
//	On success, returns HTTP 200 OK.
//	On invalid JSON format returns HTTP 400 Bad request.
func UpdateMetricsJson(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var metrics []dto.Metrics
		if err := c.ShouldBindJSON(&metrics); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}
		metrics = dto.OptimizeMetrics(metrics)
		if err := svc.SetMetrics(metrics); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

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

// UpdateGaugeMetricRaw handles raw update requests that target gauge metric.
//
//	Method: POST
//	Endpoint: /update/gauge/{name}/{value}
//
// Example usage with curl:
//
//	curl -X POST http://localhost:9009/update/gauge/test/100.1
//
//	On success, returns HTTP 200 OK.
//	On invalid value returns HTTP 400 Bad request.
func UpdateGaugeMetricRaw(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		v := c.Param("value")
		value, err := strconv.ParseFloat(v, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err := svc.SetGaugeMetric(name, value); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		c.Writer.WriteHeader(http.StatusOK)
	}
}
