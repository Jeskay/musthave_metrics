// Package handlers contains functions that handle incoming http requests.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/metric"
)

// MetricString stores metric information in string format.
type MetricString struct {
	Name  string // Name of the metric
	Type  string // Type of the metric (gauge or counter)
	Value string // Value of the metric
}

// NewMetricString returns new instance of MetricString.
func NewMetricString(metric dto.Metrics) MetricString {
	mStr := MetricString{
		Name: metric.ID,
		Type: metric.MType,
	}
	if internal.MetricType(metric.MType) == internal.CounterMetric && metric.Delta != nil {
		mStr.Value = strconv.FormatInt(*metric.Delta, 10)
	} else if metric.Value != nil {
		mStr.Value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}
	return mStr
}

// ListMetrics handles metric list request.
//
//	Method: GET
//	Endpoint: /
//
// Example usage with curl:
//
//	curl -X GET http://localhost:9009/
//
// On success, returns HTML page with list of all available metrics.
func ListMetrics(svc *metric.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		list, err := svc.GetAllMetrics()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		metrics := make([]MetricString, len(list))
		for i, m := range list {
			metrics[i] = NewMetricString(m)
		}

		c.HTML(http.StatusOK, "/templates/list.tmpl", gin.H{
			"Metrics": metrics,
			"title":   "List of Metrics",
		})
	}
}
