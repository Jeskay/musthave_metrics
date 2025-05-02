package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/gin-gonic/gin"
)

// GetCounterMetric handles counter metric value request.
//
//	Method: GET
//	Endpoint: /value/counter/{name}
//
// Example usage with curl:
//
//	curl -X GET http://localhost:9009/value/counter/testMetric
//
//	On success, returns HTTP 200 OK with string value of requested metric.
//	On requesting invalid metric, returns HTTP 404 Not found.
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

// GetMetricJson handles metric value request in JSON format.
//
//	Method: POST
//	Endpoint: /value/
//
// Expected JSON body:
//
//	{
//		"id": "metric1",
//		"type": "counter"
//	}
//
// Example usage with curl:
//
//	curl -X POST http://localhost:9009/value/ \
//			-H "Content-Type: application/json"
//			-d '{"id": "metric1", "type": "counter"}'
//
//	On success, returns HTTP 200 OK with metric value.
//	On invalid JSON body or metric type returns HTTP 400 Bad request.
//	On requesting invalid metric returns HTTP 404 Not found.
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

// GetGaugeMetric handles gauge metric value request.
//
//	Method: GET
//	Endpoint: /value/gauge/{name}
//
// Example usage with curl:
//
//	curl -X GET http://localhost:9009/value/gauge/metric1
//
//	On success, returns HTTP 200 OK with metric value.
//	On requesting invalid metric returns HTTP 404 Not found.
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
