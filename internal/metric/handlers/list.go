package handlers

import (
	"net/http"
	"strconv"

	"github.com/Jeskay/musthave_metrics/internal"
	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/metric"
	"github.com/gin-gonic/gin"
)

type MetricString struct {
	Name  string
	Type  string
	Value string
}

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
